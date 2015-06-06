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
)

//globals
var debug bool
var help bool
var timing bool
var vip string

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

	flag.Parse()
	return
}

func printHelp() {
	fmt.Println(
		`	Usage: rangerclient [OPTIONS] <query>
	eg: yr %RANGE
	--debug .................... Run client in debug mode
	--help ..................... Prints this documentation
	--timing ................... profiling of execution time
	--vip ...................... Range vip endpoint`)
	os.Exit(0)
	return
}

func main() {
	query := ""
	if len(os.Args) == 1 {
		printHelp()
	}
	query = os.Args[len(os.Args)-1] //trying to get the last element of an array
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
	if timing == true {
		fmt.Printf("Range-Expand-Microsecond : %sms \n", res.Header.Get("Range-Expand-Microsecond"))
	}
}
