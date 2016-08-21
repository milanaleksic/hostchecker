package hostchecker

import (
	"fmt"
	"io/ioutil"
	"encoding/json"
)

// Expectation is a definition of an expectation for a certain server.
// Since SSH is used, one needs to provide the username and password to access the remote server,
// as well as definitions of expectations: servies, expected URL responses etc
type Expectation struct {
	Server          string `json:"server"`
	User            string `json:"user"`
	Password        string  `json:"password"`
	UpstartServices []*Service `json:"upstart"`
	CustomServices  []*CustomService `json:"custom"`
	Responses	[]*Response `json:"responses"`
}

// ReadExpectationsFromJSON provides the expectation structs from external (JSON) DSL.
func ReadExpectationsFromJSON(filename string) []Expectation {
	var target []Expectation
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