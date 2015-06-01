// We have to load the test etcd cluster, this load happens by
// using the range interface querying the filestore.
// NOTE: We use port 13824 as test etcd cluster to make sure
// we won't end up messing up production cluster by mistake

// Setup
// -----
// Starting Test Etcd Cluster:
// /path/to/etcd --listen-client-urls 'http://localhost:13824' --advertise-client-urls 'http://localhost:13824'
// Testing using etcdctl:
// /path/to/etcdctl --debug -C 127.0.0.1:13824 ls --recursive

package main

import (
	"fmt"
	"github.com/coreos/go-etcd/etcd"
	"log"
	"rangeexpr"
	"rangestore"
	"strings"
)

func main() {
	log.SetFlags(log.Lshortfile)
	var storedir = "../rangestore/t"
	log.Printf("Connecting to FileStore [storedir: %s]", storedir)
	var store, err = rangestore.ConnectFileStore(storedir, -1, false)
	if err != nil {
		log.Fatal("Error in Connecting to Store", err)
	}

	var host = "127.0.0.1"
	var port = 13824
	log.Printf("Connecting to Etcd [http://%s:%d]", host, port)
	machines := []string{fmt.Sprintf("http://%s:%d", host, port)}
	client := etcd.NewClient(machines)
	// defer the close
	defer func() {
		log.Println("Closing Etcd Connection")
		client.Close()
	}()

	var res *[]string
	var errs []error
	var query string

	var response *etcd.Response

	response, err = client.Get("_range_store", false, false)
	// we have to make sure we have a clean testing area
	if err == nil {
		log.Printf("Looks like the test-etcd (http://%s:%d) is already populated [%s : %s]", host, port, response.Node.Key, response.Node.Value)
		log.Fatal("Please reformat your test-etcd cluster")
	} else if err != nil && err.(*etcd.EtcdError).ErrorCode != 100 {
		// don't fatal out if the error is, key not present, it is expected
		log.Fatal(err)
	}

	// make a note saying, we have started populating data
	_, _ = client.Set("_range_store", "loading", 0)

	// we can load the etcd in 4 major steps
	// 1. get the top level range and create dirs
	// 2. do the same for second level and third level
	// 3. for the fourth level (leaf nodes) create the dirs
	// 4. pull KEYS and push the key/value pairs to nodes

	// step 1,2,3
	log.Println("Steps 1,2,3 (create dirs) - START")
	for _, i := range []string{"%RANGE", "%%RANGE", "%%%RANGE", "%%%%RANGE"} {
		query = i
		res, errs = exandQuery(query, store)
		// if error, crap out
		if len(errs) > 0 {
			log.Fatal(errs)
		}
		// iterate over the range results
		for _, j := range *res {
			err = createEtcdDir(j, client)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
	log.Println("Steps 1,2,3 (create dirs) - DONE")
	// now we have variable res containing all leaf nodes

	// step 4
	// * get keys
	// * for each key: get values
	// * create the key/value pair
	log.Println("Steps 4 (create leaf nodes) - START")
	for _, i := range *res {
		// get all keys
		keys, errs := exandQuery(fmt.Sprintf("%%%s:KEYS", i), store)
		if len(errs) > 0 {
			log.Fatal(errs)
		}
		// we need a marker for leaf node
		*keys = append(*keys, "_leaf")
		// for each key
		for _, j := range *keys {
			value, errs := exandQuery(fmt.Sprintf("%%%s:%s", i, j), store)
			if len(errs) > 0 {
				log.Fatal(errs)
			}
			err = createEtcdKeyValue(i, j, value, client)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
	log.Println("Steps 4 (create leaf nodes) - DONE")

	// make a note saying, data has been populated
	_, _ = client.Set("_range_store", "loaded", 0)

	return
}

// creates key/value in etcd
func createEtcdKeyValue(dir string, key string, value *[]string, client *etcd.Client) error {
	dir = strings.Replace(dir, "-", "/", -1)
	_, err := client.Set(fmt.Sprintf("%s/%s", dir, key), strings.Join(*value, "\t"), 0)
	return err
}

// creates dir in etcd
func createEtcdDir(dir string, client *etcd.Client) error {
	dir = strings.Replace(dir, "-", "/", -1)
	_, err := client.CreateDir(dir, 0)
	return err
}

// calls filestore and expands the query
func exandQuery(query string, store *rangestore.FileStore) (*[]string, []error) {
	if strings.HasSuffix(query, "_leaf") {
		return &[]string{"_leaf"}, nil
	}
	var r *rangeexpr.RangeExpr
	r = &rangeexpr.RangeExpr{Buffer: query}
	r.Init()
	r.Expression.Init(query)
	if err := r.Parse(); err != nil {
		log.Fatal(err)
	}
	r.Execute()
	return r.Evaluate(store)
}
