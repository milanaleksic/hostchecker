package main

import (
	"fmt"
	"os"
)

// TODO: check age of JARs

func main() {
	expectations := readExpectationsFromJSON()

	var failures []failure
	for _, expectation := range expectations {
		failures = append(failures, checkServer(expectation)...)
	}
	if len(failures) > 0 {
		fmt.Println("\n\nFollowing violations are present:")
		for index, failure := range failures {
			fmt.Printf("%d. %s\n", index+1, failure.String())
		}
		os.Exit(1)
	}
}
