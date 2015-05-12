package main

import (
	"fmt"
	"log"
	"os"

	"rangeexpr"
)

func main() {
	if len(os.Args) < 2 {
		name := os.Args[0]
		fmt.Printf("Usage: %v \"EXPRESSION\"\n", name)
		fmt.Printf("Example: %v \"a\n", name)
		os.Exit(1)
	}
	expression := os.Args[1]
	r := &rangeexpr.RangeExpr{Buffer: expression}
	r.Init()
	r.Expression.Init(expression)
	if err := r.Parse(); err != nil {
		log.Fatal(err)
	}
	r.Execute()
	fmt.Printf("= %v\n", r.Evaluate())
}
