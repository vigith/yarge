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
	"strings"
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
	flag.BoolVar(&help, "help", false, "Help")
	flag.BoolVar(&timing, "timing", false, "display timing")
	flag.StringVar(&vip, "vip", "localhost", "VIP endpoint")
	flag.Parse()
	return
}

func printHelp() {
	fmt.Printf(
		`	Usage: %s [OPTIONS] <query>
	eg: %s %%RANGE
	--debug .................... Debug
	--help ..................... Good Ol' Help
	--timing ................... Execution Time as provided by rangeserver
	--vip ...................... Range VIP Endpoint (default: localhost:9999)

Documentation: https://github.com/vigith/yarge
`, os.Args[0], os.Args[0])
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

	_url := fmt.Sprintf("http://%s/v1/range/list?%s", vip, url.QueryEscape(query))
	if debug {
		fmt.Println("Range URL: ", _url)
	}
	res, err := http.Get(_url)
	// fatal out if we have error
	if err != nil || res.StatusCode != 200 {
		_url_human, _ := url.QueryUnescape(_url)
		results, _ := ioutil.ReadAll(res.Body)
		log.Printf("ERROR, URL: (%s) Error: (%s) HTTP_Errors: (%v)\n", _url_human, strings.TrimSuffix(string(results), "\n"), err)
		os.Exit(1)
	}

	results, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	res.Body.Close()

	// show the time taken for range expansion (as per the rangeserver)
	if timing || debug {
		fmt.Printf("Range-Expand-Microsecond: %s \n", res.Header.Get("Range-Expand-Microsecond"))
	}

	// if error, print the error
	if res.Header.Get("Range-Err-Count") != "" {
		fmt.Printf("%s", res.Header.Get("Range-Err-Count"))
	} else { // else print the result
		fmt.Printf("%s\n", results)
	}

	return
}
