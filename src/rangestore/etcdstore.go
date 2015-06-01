// etcd will queried when this store is used

package rangestore

import (
	"errors"
	"fmt"
	"github.com/coreos/go-etcd/etcd"
	"log"
	"strings"
)

const _leaf = "_leaf"

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

// reads the child clusters of this cluster.
// returns only those nodes for which this cluster is parent
func (e *EtcdStore) listClusters(cluster string) ([]string, error) {
	var dir = e.clusterToPath(cluster)
	var children = make([]string, 0)
	// list the nodes under this cluster.
	// we are sure when this call was made, the check
	// has been made to sure this is not a leaf node
	response, _, _, _, err := e.retrieveFromEtcd(dir, false, false)
	// if there is an error, return err
	if err != nil {
		return []string{}, err
	}
	// if response is NOT for a dir
	if !response.Node.Dir {
		var _err = fmt.Sprintf("Expected value of lookup [%s] to be a dir, recieved a leaf file", dir)
		log.Printf(_err)
		return []string{}, errors.New(_err)
	}

	for _, n := range response.Node.Nodes {
		children = append(children, n.Value)
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
	log.Println("there", object)
	// if all is good
	return response, response.Node.Key, response.Node.Value, true, nil
}
