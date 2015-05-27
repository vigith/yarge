// When using FileStore, we will have yamls files to store the data.

// FIXME:
// Since I am not planning to use this in production (I prefer etcd as store)
// I would NOT be using any optimizations to cache the process file
// contents (ie, result of yaml parser) Please give me a patch (one way to clear
// the cache is by sending some signals to reread the file and update the cache.
// Also we should make sure we reread only the specific files in question).

package rangestore

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

var _config = "cluster.yaml"

type FileStore struct {
	StorePath string
	MaxDepth  int
}

// check whether the StorePath Exists, etc
func ConnectFileStore(dir string, depth int) (f *FileStore, err error) {
	// removing trailing path seperator
	if os.IsPathSeparator(dir[len(dir)-1]) {
		dir = dir[:len(dir)-1]
	}
	var fi os.FileInfo
	// check whether the dir exists
	fi, err = os.Stat(dir)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Path [%s] is not a FileStore directory (ERROR: %s)", dir, err))
	}
	// check whether it is a dir
	if !fi.IsDir() {
		return nil, errors.New(fmt.Sprintf("Path [%s] is not a directory", dir))
	}
	f = &FileStore{StorePath: dir, MaxDepth: depth}
	return f, nil
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
//   but do an ArraytoSet with the results array
func (f *FileStore) ClusterLookup(cluster *[]string) (*[]string, error) {
	// store the resuls
	var results = make([]string, 0)
	// for each cluster, do a lookup
	// (this will only happen only for nested lookups eg, %%..)
	for _, elem := range *cluster {
		var err error
		isLeaf, err := f.checkIsLeafNode(elem)
		if err != nil {
			return &[]string{}, err
		}
		// if it is a leaf node, we need do a KeyLookup (NODES)
		if isLeaf {
			// by default, lookup for NODES
			result, err := f.KeyLookup(&[]string{elem}, "NODES")
			if err != nil {
				return &[]string{}, err
			}
			results = append(results, *result...)
		} else { // we need to return the children
			result, err := f.listClusters(elem)
			if err != nil {
				return &[]string{}, err
			}
			results = append(results, result...)
		}

	}

	return &results, nil
}

func (f *FileStore) KeyLookup(cluster *[]string, key string) (*[]string, error) {
	// store the resuls
	var results = make([]string, 0)
	// this will most likely be single element arrays
	// can't think of a reason otherwise
	for _, elem := range *cluster {
		// 1. read the config
		// 2. do a key lookup
		// 3. append the result
		content, err := f.readClusterConfig(elem)
		if err != nil {
			return &[]string{}, errors.New(fmt.Sprintf("KeyLookup for [%s] Failed (Error: %s)", elem, err))
		}
		result, err := yamlKeyLookup(content, key)
		if err != nil {
			return &[]string{}, errors.New(fmt.Sprintf("KeyLookup for [%s] Failed (Error: %s)", elem, err))
		}
		results = append(results, *result...)
	}

	return &results, nil
}

////////////////////
// LOOKUP REVERSE //
////////////////////

func (f *FileStore) KeyReverseLookup(key string) (*[]string, error) {
	return &[]string{}, errors.New("KeyReverseLookup Failed, returning empty")
}

func (f *FileStore) KeyReverseLookupAttr(key string, attr string) (*[]string, error) {
	return &[]string{}, errors.New("KeyReverseLookupAttr Failed, returning empty")
}

func (f *FileStore) KeyReverseLookupHint(key string, attr string, hint string) (*[]string, error) {
	return &[]string{}, errors.New("KeyReverseLookupHint Failed, returning empty")
}

////////////////////////
// Internal Functions //
////////////////////////

// given a cluster name, it will convert to cluster
// in the file system
func (f *FileStore) clusterToPath(cluster string) string {
	return fmt.Sprintf("%s/%s", f.StorePath, strings.Replace(cluster, "-", "/", -1))
}

// reads the child clusters of this cluster.
// returns only those nodes for which this cluster is parent
func (f *FileStore) listClusters(cluster string) ([]string, error) {
	var dir = f.clusterToPath(cluster)
	var children = make([]string, 0)
	files, err := ioutil.ReadDir(dir)
	// if there is an error, return err
	if err != nil {
		return []string{}, err
	}
	for _, f := range files {
		if f.IsDir() {
			children = append(children, f.Name())
		}
	}
	return children, nil
}

// Checks whether the cluster is in leaf or not
// It will return error if the cluster doesn't exist,
// false if not a leaf node, true otherwise
func (f *FileStore) checkIsLeafNode(cluster string) (bool, error) {
	var err error
	var dir = f.clusterToPath(cluster)
	var fi os.FileInfo
	// check whether it is a dir
	fi, err = os.Stat(dir)
	if err != nil {
		return false, errors.New(fmt.Sprintf("cluser [%s] is NOT FOUND in FileStore [%s w.r.t %s] (ERROR: %s)", cluster, dir, f.StorePath, err))
	}
	if !fi.IsDir() {
		return false, errors.New(fmt.Sprintf("cluser [%s] is NOT A DIRECTORY in FileStore [%s w.r.t %s] (ERROR: %s)", cluster, dir, f.StorePath, err))
	}

	// now check whether this dir has "cluster.yaml" as its direct child
	_, err = os.Stat(fmt.Sprintf("%s/%s", dir, _config))
	// if err is nil, it means file exists
	if err == nil {
		return true, nil
	} else if os.IsNotExist(err) {
		return false, nil
	} else {
		return false, errors.New(fmt.Sprintf("cluser [%s] is NEITHER a LeafNode or a Cluster Dir in FileStore [%s w.r.t %s] (ERROR: %s)", cluster, dir, f.StorePath, err))
	}

	// not a dir
	return true, nil
}

// Given a cluster name, it will read the corresponding cluster config
// and return the file content as a string
func (f *FileStore) readClusterConfig(cluster string) (content []byte, err error) {
	var dir = f.clusterToPath(cluster)
	content, err = ioutil.ReadFile(fmt.Sprintf("%s/%s", dir, _config))
	if err != nil {
		return []byte{}, err
	}

	return content, nil
}
