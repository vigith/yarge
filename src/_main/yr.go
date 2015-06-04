// Basic tool to talk to the rangeserver over http

package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		name := os.Args[0]
		fmt.Printf("Usage: %v \"EXPRESSION\"\n", name)
		fmt.Printf("Example: %v \"%%RANGE\n", name)
		os.Exit(1)
	}
	query := os.Args[1]
	res, err := http.Get(fmt.Sprintf("http://localhost:9999/v1/range/list?%s", url.QueryEscape(query)))
	if err != nil {
		log.Fatal(err)
	}
	results, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		log.Fatal(err)
	}
	//	if res.Header.Get() {
	//	}
	fmt.Printf("%s\n", results)
}
