package main

import (
	"fmt"
	"io/ioutil"
	"encoding/json"
)

type expectation struct {
	Server          string `json:"server"`
	User            string `json:"user"`
	Password        string  `json:"password"`
	UpstartServices []*service `json:"upstart"`
	CustomServices  []*customService `json:"custom"`
	Responses	[]*response `json:"responses"`
}

func readExpectationsFromJSON() []expectation {
	var target []expectation
	filename := "expectations.json"
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Printf("Could not read data from file filename=%v, err=%v\n", filename, err)
	} else {
		err := json.Unmarshal(data, &target)
		if err != nil {
			fmt.Printf("Could not parse data from file filename=%v, err=%v\n", filename, err)
		}
	}
	for _, e := range target {
		for _, s := range e.UpstartServices {
			s.Server = e.Server
		}
		for _, s := range e.CustomServices {
			s.Server = e.Server
		}
		for _, r := range e.Responses {
			r.Server = e.Server
		}
	}
	return target
}