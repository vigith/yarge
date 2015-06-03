// etcd will queried when this store is used

package etcdstore

import (
	"errors"
	"fmt"
	"github.com/coreos/go-etcd/etcd"
	"log"
	"strings"
)

const _leaf = "_leaf"
const _sep = "\t"
const _roptimize = "/_roptimize"

type EtcdStore struct {
	hosts      []string     // http://host1:port,..
	ROptimize  bool         // reverse lookup optimization
	FastLookup bool         // fast return, will return the first match
	client     *etcd.Client // etcd connection object
	storenode  string       // path to where the range store is etcd
}

// Connect to the Etcd Store
// 1. make sure Etcd Store is running
// 2. check whether _range_store value is set to 'loaded'
// 3. create the interface for client connections
func ConnectEtcdStore(hosts []string, roptimize, fast bool, node string) (e *EtcdStore, err error) {
	if len(node) > 0 && node[len(node)-1] == '/' {
		node = node[:len(node)-1]
	}
	client := etcd.NewClient(hosts)
	e = &EtcdStore{hosts: hosts, ROptimize: roptimize, FastLookup: fast, client: client, storenode: node}
	// test the consistency of etcd store
	var obj = fmt.Sprintf("%s/%s", e.storenode, "_range_store")
	var response *etcd.Response
	response, err = client.Get(obj, false, false)
	// we have to make sure we have a clean etcd store
	if err != nil && err.(*etcd.EtcdError).ErrorCode == 100 {
		log.Printf("ERROR: Looks like the etcd-store is not loaded with data to serve as rangestore [Key: %s, Value: NOT FOUND]\n", obj)
		return nil, errors.New("ERROR: etcd-store is NOT LOADED with data")
	} else if err != nil {
		log.Println("ERROR: Etcd Store Returned Error")
		log.Println(err)
		return nil, err
	} else if err == nil {
		// if we get no error, it means key if present. We need to make sure it is set to 'loaded'
		if response.Node.Value != "loaded" {
			log.Printf("ERROR: Looks like the etcd-store is not ready to serve as rangestore [Key: %s, Value: %s]\n", response.Node.Key, response.Node.Value)
			return nil, errors.New("ERROR: etcd-store is NOT READY to serve")
		}
	}

	return e, nil
}

func (e *EtcdStore) DisconnectEtcdStore() {
	e.client.Close()
	return
}

////////////////////
// LOOKUP CLUSTER //
////////////////////

// LOGIC
// -----
// * for the first element in cluster create results array
//   * check whether the cluster is a leaf node
//   * if yes, call KeyLookup, with key == NODES
//   * if not, call listClusters
// * if more elements are there, repeat the above
//   (unlike, filestore we don't call ArrayToSet since etcd
//   is populated by a program, not by morals)
func (e *EtcdStore) ClusterLookup(cluster *[]string) (*[]string, error) {
	// store the resuls
	var results = make([]string, 0)
	// for each cluster, do a lookup
	// (this will only happen only for nested lookups eg, %%..)
	for _, elem := range *cluster {
		// handle RANGE separately
		if elem == "RANGE" {
			elem = "/"
		}
		var err error
		isLeaf, err := e.checkIsLeafNode(elem)
		if err != nil {
			return &[]string{}, err
		}
		// if it is a leaf node, we need do a KeyLookup (NODES)
		if isLeaf {
			// by default, lookup for NODES
			result, err := e.KeyLookup(&[]string{elem}, "NODES")
			if err != nil {
				return &[]string{}, err
			}
			results = append(results, *result...)
		} else { // we need to return the children
			result, err := e.listClusters(elem)
			if err != nil {
				return &[]string{}, err
			}
			results = append(results, result...)
		}

	}

	return &results, nil
}

func (e *EtcdStore) KeyLookup(cluster *[]string, key string) (*[]string, error) {
	// store the resuls
	var results = make([]string, 0)
	// this will most likely be single element arrays
	// can't think of a reason otherwise
	for _, elem := range *cluster {
		dir := e.clusterToPath(elem)
		node := fmt.Sprintf("%s/%s", dir, key)
		var result []string
		if key == "KEYS" {
			response, _, _, found, err := e.retrieveFromEtcd(dir, false, false)
			// if there is an error, return err
			if err != nil { // got error
				return &[]string{}, err
			} else if !found { // key not found
				return &[]string{}, errors.New(fmt.Sprintf("KeyLookup for [%s:%s] Failed (Error: No KEY Found)", elem, key))
			}

			// if response is NOT for a dir
			if !response.Node.Dir {
				var _err = fmt.Sprintf("Expected value of lookup [%s] to be a dir, recieved a leaf file", dir)
				log.Printf(_err)
				return &[]string{}, errors.New(_err)
			}

			for _, n := range response.Node.Nodes {
				// replace '/' with '-', also root will always be '/'
				_node := strings.Split(n.Key, "/")
				if len(_node) > 0 {
					result = append(result, _node[len(_node)-1])
				}
			}
		} else {
			// 1. read the key in etcd
			// 2. append the result
			_, _, value, found, err := e.retrieveFromEtcd(node, false, false)
			if err != nil {
				return &[]string{}, errors.New(fmt.Sprintf("KeyLookup for [%s:%s] Failed (Error: %s)", elem, key, err))
			} else if !found {
				return &[]string{}, errors.New(fmt.Sprintf("KeyLookup for [%s:%s] Failed (Error: No KEY Found)", elem, key))
			}
			result = strings.Split(value, _sep)
		}
		// append the result with results
		results = append(results, result...)
	}

	return &results, nil
}

////////////////////
// LOOKUP REVERSE //
////////////////////

// same as KeyReverseLookupAttr where attr == NODES
func (e *EtcdStore) KeyReverseLookup(key string) (*[]string, error) {
	return e.KeyReverseLookupAttr(key, "NODES")
}

// same as KeyReverseLookupAttr where attr == NODES and hint == ""
func (e *EtcdStore) KeyReverseLookupAttr(key string, attr string) (*[]string, error) {
	// optimization, for nodes don't do the tough thing
	if e.ROptimize && attr == "NODES" {
		return e.optimizedNodeReverseLookup(key)
	}
	return e.KeyReverseLookupHint(key, attr, "")
}

// given a key, it will search for the cluster where the attr has that key,
// hint is to limit the scope of search
func (e *EtcdStore) KeyReverseLookupHint(key string, attr string, hint string) (*[]string, error) {
	var clusters *[]string
	var err error
	var results = make([]string, 0)
	var seen bool

	clusters, err = e.getAllLeafNodes(hint)
	if err != nil {
		return &results, nil
	}

	for _, elem := range *clusters {

		// 1. read the key in etcd
		// 2. append the result
		_, _, value, found, err := e.retrieveFromEtcd(fmt.Sprintf("%s/%s", e.clusterToPath(elem), attr), false, false)
		if err != nil {
			return &[]string{}, errors.New(fmt.Sprintf("KeyLookup for [%s:%s] Failed (Error: %s)", elem, key, err))
		} else if !found {
			continue
		} else {
			result := strings.Split(value, _sep)
			for _, i := range result {
				if i == key {
					results = append(results, elem)
					seen = true
					break
				}
			}
			if seen && e.FastLookup {
				return &results, nil
			}
		}
	}
	return &results, nil
}

////////////////////////
// Internal Functions //
////////////////////////

// given a cluster name, it will convert to cluster
// in the file system
func (e *EtcdStore) clusterToPath(cluster string) string {
	return fmt.Sprintf("%s/%s", e.storenode, strings.Replace(cluster, "-", "/", -1))
}

// Get all the leaf cluster nodes for a given dir
// it is not efficient since we have to walk down all the path
func (e *EtcdStore) getAllLeafNodes(root string) (*[]string, error) {
	var results = make([]string, 0)
	var err error

	// root a leaf node
	isleaf, err := e.checkIsLeafNode(root)
	if err != nil {
		return &[]string{}, err
	}
	if isleaf {
		return &[]string{root}, nil
	}

	err = e._getAllLeafNodes(e.clusterToPath(root), &results)

	if err != nil {
		return &[]string{}, err
	}

	// fix the results in place
	for i := 0; i < len(results); i++ {
		results[i] = strings.Replace(strings.Trim(results[i], "/"), "/", "-", -1)
	}

	return &results, nil
}

// the function that really does the work
func (e *EtcdStore) _getAllLeafNodes(root string, results *[]string) error {
	response, _, _, found, err := e.retrieveFromEtcd(root, false, false)
	// if there is an error, return err
	if err != nil { // got error
		return err
	} else if !found { // key not found
		return errors.New(fmt.Sprintf("DirLookup for [%s] Failed (Error: No DIR Found)", root))
	}

	// if response is NOT for a dir
	if !response.Node.Dir {
		var _err = fmt.Sprintf("Expected value of lookup [%s] to be a dir, recieved a leaf file", root)
		log.Printf(_err)
		return errors.New(_err)
	}

	for _, n := range response.Node.Nodes {
		status, err := e.checkIsLeafNode(n.Key)
		if err != nil {
			return err
		}
		if status {
			*results = append(*results, n.Key)
		} else {
			_ = e._getAllLeafNodes(n.Key, results)
		}
	}

	return nil
}

// reads the child clusters of this cluster.
// returns only those nodes for which this cluster is parent
func (e *EtcdStore) listClusters(cluster string) ([]string, error) {
	var dir = e.clusterToPath(cluster)
	var children = make([]string, 0)
	// list the nodes under this cluster.
	// we are sure when this call was made, the check
	// has been made to sure this is not a leaf node
	response, _, _, found, err := e.retrieveFromEtcd(dir, false, false)
	// if there is an error, return err
	if err != nil { // got error
		return []string{}, err
	} else if !found { // key not found
		return []string{}, nil
	}

	// if response is NOT for a dir
	if !response.Node.Dir {
		var _err = fmt.Sprintf("Expected value of lookup [%s] to be a dir, recieved a leaf file", dir)
		log.Printf(_err)
		return []string{}, errors.New(_err)
	}

	for _, n := range response.Node.Nodes {
		// replace '/' with '-', also root will always be '/'
		_node := strings.Replace(n.Key[1:], "/", "-", -1)
		children = append(children, _node)
	}

	return children, nil
}

// Checks whether the cluster is in leaf or not
// It will return error if the cluster doesn't exist,
// false if not a leaf node, true otherwise
func (e *EtcdStore) checkIsLeafNode(cluster string) (found bool, err error) {
	var dir = e.clusterToPath(cluster)
	_, _, _, found, err = e.retrieveFromEtcd(fmt.Sprintf("%s/%s", dir, _leaf), false, false)
	// found and err are set properly by retrieveFromEtcd
	return found, err
}

// given a object, return the response from the etcd cluster
func (e *EtcdStore) retrieveFromEtcd(object string, sort, recursive bool) (response *etcd.Response, key string, value string, found bool, err error) {
	response, err = e.client.Get(object, sort, recursive)

	// Check whether the error is Key NOT Found
	if err != nil && err.(*etcd.EtcdError).ErrorCode == 100 {
		return response, object, "", false, nil
	} else if err != nil { // error could be cluster is unreachble, etc..
		log.Printf("ERROR: Etcd Store Returned Error: %s\n", err)
		return response, object, "", false, err
	}

	// if all is good
	return response, response.Node.Key, response.Node.Value, true, nil
}

// same as KeyReverseLookupAttr where attr == NODES and hint == ""
func (e *EtcdStore) optimizedNodeReverseLookup(key string) (*[]string, error) {
	_, _, value, found, err := e.retrieveFromEtcd(fmt.Sprintf("%s/%s", _roptimize, key), false, false)
	if !found {
		return &[]string{}, nil
	} else if err != nil {
		return &[]string{}, err
	}
	values := strings.Split(value, _sep)
	return &values, nil
}
