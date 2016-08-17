package main

import (
	"golang.org/x/crypto/ssh"
	"bytes"
	"flag"
	"log"
	"strings"
	"regexp"
	"fmt"
	"io/ioutil"
	"encoding/json"
	"strconv"
)

// TODO: check age of JARs

func main() {
	explain := flag.Bool("explain", false, "application should explain expectations")
	flag.Parse()

	expectations := readExpectationsFromJson()
	if *explain {
		fmt.Printf("Expectations: %+v", expectations)
		return
	}

	var failures []failure
	for _, expectation := range expectations {
		checkServer(expectation, failures)
	}
}
