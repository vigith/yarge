// Basic tool to talk to the rangeserver over http

package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"
)

//globals
var debug bool
var help bool
var timing bool
var vip string
var query string

func init() {
	parseFlags()
	if help == true {
		printHelp()
		os.Exit(0)
	}
	return
}

func parseFlags() {
	flag.BoolVar(&debug, "debug", false, "enable debug")
	flag.BoolVar(&help, "help", false, "enable help")
	flag.BoolVar(&timing, "timing", false, "enable timing")
	flag.StringVar(&vip, "vip", "localhost", "vip endpoint")
	flag.StringVar(&query, "query", "", "query")

	flag.Parse()
	return
}

func printHelp() {
	fmt.Println(
		`Usage: rangerclient [OPTIONS]
	--debug......................Run client in debug mode
	--help.......................Prints this documentation
	--timing.....................profiling of execution time
	--query......................query string
	--vip........................Range vip endpoint`)
	return
}

func main() {
	t1 := time.Now()
	if query == "" {
		fmt.Println("Need a query to execute")
		os.Exit(0)
	}
	if vip == "localhost" {
		vip = "localhost:9999"
	}
	res, err := http.Get(fmt.Sprintf("http://%s/v1/range/list?%s", vip, url.QueryEscape(query)))
	if err != nil {
		log.Fatal(err)
	}
	results, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		log.Fatal("%s", err)
	}
	if res.Header.Get("Range-Err-Count") != "" {
		fmt.Printf("%s", results)
	} else {
		fmt.Printf("%s\n", results)
	}

	t2 := time.Now()
	if timing == true {
		fmt.Printf("Took %v to run", t2.Sub(t1))
	}

}
