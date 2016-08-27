package main

import (
	"fmt"
	"os"
	"github.com/milanaleksic/hostchecker"
	"flag"
)

func main() {
	fileName := flag.String("filename", "expectations.json", "which expectation file should be used")
	flag.Parse()

	expectations := hostchecker.ReadExpectationsFromJSON(*fileName)

	var failures []hostchecker.Failure
	for _, expectation := range expectations {
		failures = append(failures, expectation.CheckServer()...)
	}
	if len(failures) > 0 {
		fmt.Println("\n\nFollowing violations are present:")
		for index, failure := range failures {
			fmt.Printf("%d. %s\n", index + 1, failure.String())
		}
		os.Exit(1)
	} else {
		fmt.Println("\n\nNo violations found")
	}
}
