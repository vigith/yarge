// A webserver to interface range cluster

package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
	// our packages
	"rangeexpr"
	"rangestore"
	"rangestore/etcdstore"
	"rangestore/filestore"
)

// globals
var store string      // name of the store
var params string     // path for filestore, server string for etcd, etc
var slowlog int       // in ms, log queries slower than these
var etcdroot string   // where does the yarge root start in etcd (useful for shared cluster)
var serveraddr string // server address
var fast bool         // is fast lookup okay
var roptimize bool    // do we have reverse lookup optimization
var debug bool        // debug
var help bool         // help

// future need to closure the function with more data to be passed?
func genericHandlerV1(fn func(http.ResponseWriter, *http.Request, interface{}), s interface{}) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) { // returns void, we don't care!
		fn(w, r, s)
	}
}

// handler for each request
// * log the request
// * log slow queries (clientip will the key to track a request)
func requestHandler(w http.ResponseWriter, r *http.Request, s interface{}) {
	var query string
	var err error

	var remoteaddr = fmt.Sprintf("%s:%s", r.Header.Get("X-Real-IP"), r.Header.Get("X-Real-Port"))
	if remoteaddr == ":" {
		remoteaddr = r.RemoteAddr
	}

	query, err = url.QueryUnescape(r.URL.RawQuery)
	if err != nil {
		log.Println("EROR> [%s] Request: [%s] Error: %s", remoteaddr, r.URL.RawQuery, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if debug {
		log.Printf("DBUG> [%s] %s", remoteaddr, query)
	}

	// exapand the query
	var results *[]string
	var errs []error

	defer func() {
		if _r := recover(); _r != nil {
			errs = append(errs, errors.New(fmt.Sprintf("[%s] %s [Panicked while Expanding Query]", remoteaddr, query)))
			results = &[]string{}
			log.Printf("[%s] %s [Panicked while Expanding Query]", remoteaddr, query)
		}
	}()

	// measure how long it took
	t0 := time.Now()
	// do the expand
	results, errs = expandQuery(query, s)
	t1 := time.Now()

	timetaken := time.Duration(t1.Sub(t0) / time.Microsecond)

	// return the results to the client
	// set the headers if we have errors
	if len(errs) > 0 {
		w.Header().Set("Range-Err-Count", fmt.Sprintf("%d", len(errs)))
		var _errs = make([]string, 0)
		for _, i := range errs {
			_errs = append(_errs, fmt.Sprintf("%s", i))
		}
		http.Error(w, strings.Join(_errs, ","), http.StatusInternalServerError)
		return
	}

	// set header with time taken to process thne request
	w.Header().Set("Range-Expand-Microsecond", fmt.Sprintf("%d", timetaken))
	// write the results
	_, err = fmt.Fprintf(w, "%s", strings.Join(*results, "\n"))
	if err != nil {
		log.Printf("ERROR> [%s] %s (Writing back to Client Failed [Reason: %s])\n", remoteaddr, query, err)
	}

	isSlow := timetaken > time.Duration(slowlog)*time.Microsecond
	//	log.Println(timetaken, time.Duration(slowlog)*time.Microsecond)
	if debug || isSlow {
		if isSlow {
			log.Printf("INFO> [SLOWQUERY] [%s] %s Result [%s] [Time Taken: %v]", remoteaddr, query, strings.Join(*results, "\n"), timetaken)
		} else {
			log.Printf("DBUG> [%s] %s Result [%s] [Time Taken: %v]", remoteaddr, query, strings.Join(*results, "\n"), timetaken)
		}
	}

	return
}

func expandQuery(query string, s interface{}) (*[]string, []error) {
	// parse the query
	var yr *rangeexpr.RangeExpr

	yr = &rangeexpr.RangeExpr{Buffer: query}
	// initialize
	yr.Init()
	yr.Expression.Init(query)
	// parse the query
	if err := yr.Parse(); err != nil {
		return &[]string{}, []error{errors.New("Parse Error")}
	}
	// build AST
	yr.Execute()

	// evaluate AST
	return yr.Evaluate(s)
}

func startServer(store interface{}) {
	// handling deploy requests
	http.HandleFunc("/v1/range/", genericHandlerV1(requestHandler, store))
	log.Printf("Range WebServer Started [%s]", serveraddr)
	http.ListenAndServe(serveraddr, nil)
	return
}

// init function to set up whatever state is required
// for real program execution
func init() {
	parseFlags()
	// handle help
	if help == true {
		printHelp()
		os.Exit(0)
	}
	return
}

func main() {
	// set log to get the code location
	log.SetFlags(log.Lshortfile)
	// create an connection to store
	var _store interface{}
	var err error
	switch store {
	case "teststore":
		_store, err = rangestore.ConnectTestStore("Test Store") // this can never return error
	case "filestore":
		var path = params
		var depth = -1
		_store, err = filestore.ConnectFileStore(path, depth, fast)
	case "etcdstore":
		var hosts = []string{params}
		_store, err = etcdstore.ConnectEtcdStore(hosts, roptimize, fast, etcdroot)
	default:
		log.Fatalf(`Unknown store [%s] (Supports only "filestore", "teststore", "etcdstore"\n`, store)
	}
	// if error, exit
	if err != nil {
		log.Fatalf("Error in Connecting to Store", err)
		return
	}

	startServer(_store)
}

// parse the flags
func parseFlags() {
	flag.StringVar(&store, "store", "teststore", "Store Name")
	flag.StringVar(&params, "params", "", "Store Parameters")
	flag.IntVar(&slowlog, "slowlog", 3, "Microseconds definition of Slow Query")
	flag.StringVar(&etcdroot, "etcdroot", "", "Root for Range in Etcd Cluster")
	flag.BoolVar(&fast, "fast", true, "Fast Lookup, return the first result")
	flag.BoolVar(&roptimize, "roptimize", true, "Reverse Lookup Optimization")
	flag.StringVar(&serveraddr, "serveraddr", "0.0.0.0:9999", "Server Address")
	flag.BoolVar(&debug, "debug", false, "Debug")
	flag.BoolVar(&help, "help", false, "Good Ol' Help")

	// parse the options
	flag.Parse()

	return
}

// print Help
func printHelp() {
	fmt.Println(
		`Usage: rangerserver [OPTIONS]
 --store ................ Store Name, it can be "teststore", "filestore" or "etcdstore" (default: "teststore")
 --params ............... Parameters for Store, (default: filestore - /var/yarge/, etcdstore - http://127.0.0.1:4001)
 --slowlog .............. Any Query that takes more than this param in ns will be logged (default: 10ns)
 --etcdroot ............. The yarge node root in etcd, useful for shared cluster (default: "")
 --fast ................. Enable Fast Lookup, return the first result for reverse lookups
 --roptimize ............ Enable reverse lookup optimization  
 --serveraddr ........... Server Listening Port (default: 0.0.0.0:9999)
 --debug ................ Debug
 --help ................. Good Ol' Help`,
	)
	return
}
