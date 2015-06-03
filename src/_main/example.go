// An Example Code on how to interface with the Server

package main

import (
	"fmt"
	"log"
	"os"

	"rangeexpr"
	"rangestore"
	"rangestore/etcdstore"
	"rangestore/filestore"
)

func main() {
	if len(os.Args) < 3 {
		name := os.Args[0]
		fmt.Printf("Usage: %v \"STORE\" \"EXPRESSION\"\n", name)
		fmt.Printf("Example: %v file \"%%RANGE\n", name)
		os.Exit(1)
	}
	store := os.Args[1]
	expression := os.Args[2]
	r := &rangeexpr.RangeExpr{Buffer: expression}
	r.Init()
	r.Expression.Init(expression)
	if err := r.Parse(); err != nil {
		log.Fatal(err)
	}
	r.Execute()

	var _store interface{}
	var err error
	switch store {
	case "teststore":
		_store, err = rangestore.ConnectTestStore("Test Store") // this can never return error
	case "filestore":
		var path = "../rangestore/filestore/t"
		var depth = -1
		var fast = false
		_store, err = filestore.ConnectFileStore(path, depth, fast)
	case "etcdstore":
		var hosts = []string{"http://127.0.0.1:13824"}
		var roptimize = false
		var fast = false
		var node = ""
		_store, err = etcdstore.ConnectEtcdStore(hosts, roptimize, fast, node)
	default:
		log.Fatalf(`Unknown store [%s] (Supports only "filestore", "teststore", "etcdstore"\n`, store)
	}
	//	var store, err = rangestore.ConnectFileStore("../rangestore/t", 3, false)
	if err != nil {
		log.Fatalf("Error in Connecting to Store", err)
		return
	}
	res, errs := r.Evaluate(_store)
	if len(errs) == 0 {
		fmt.Printf("Result = %s\n", *res)
	} else {
		fmt.Printf("errors = %s\n", errs)
	}
}
