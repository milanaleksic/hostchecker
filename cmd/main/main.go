package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/milanaleksic/hostchecker"
	"github.com/milanaleksic/hostchecker/failure"
)

func main() {
	fileName := flag.String("filename", "expectations.json", "which expectation file should be used")
	flag.Parse()

	expectations := hostchecker.ReadExpectationsFromJSON(*fileName)

	var failures []*failure.Failure
	for _, expectation := range expectations {
		failures = append(failures, expectation.CheckServer()...)
	}
	if len(failures) > 0 {
		fmt.Println("\n\nFollowing violations are present:")
		for index, fail := range failures {
			fmt.Printf("%d. %s\n", index+1, fail.String())
		}
		os.Exit(1)
	} else {
		fmt.Println("\n\nNo violations found")
	}
}
