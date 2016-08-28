package hostchecker

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

// Verifiable represents a named "thing" that can be verified using an expectation and a running context
type Verifiable interface {
	String() string
	CheckExpectation(expec *Expectation, context *runningContext) (failures []Failure)
}

// Expectation is a definition of an expectation for a certain server.
// Since SSH is used, one needs to provide the username and password to access the remote server,
// as well as definitions of expectations: services, expected URL responses etc
type Expectation struct {
	Server          string           `json:"server"`
	User            string           `json:"user"`
	Password        string           `json:"password"`
	UpstartServices []UpstartService `json:"upstart"`
	CustomServices  []CustomService  `json:"custom"`
	Responses       []Response       `json:"responses"`
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

// CheckServer is a host verifier entry point: for a known expectation, provide failures if they exist
func (expec *Expectation) CheckServer() (failures []Failure) {
	fmt.Printf("\nChecking services on host %s\n", expec.Server)
	var context *runningContext
	if expec.demandsSSH() {
		context = newRunningContextWithSSH(expec.User, expec.Password, expec.Server)
		defer context.sshClient.Close()
	} else {
		context = &runningContext{}
	}
	for _, verifiable := range expec.getAllVerifiables() {
		fmt.Printf("Checking verifiable %s of type %T\n", verifiable.String(), verifiable)
		failures = append(failures, verifiable.CheckExpectation(expec, context)...)
	}
	return
}

func (expec *Expectation) getAllVerifiables() (verifiables []Verifiable) {
	for _, v := range expec.UpstartServices {
		verifiables = append(verifiables, v)
	}
	for _, v := range expec.CustomServices {
		verifiables = append(verifiables, v)
	}
	for _, v := range expec.Responses {
		verifiables = append(verifiables, v)
	}
	return
}

func (expec *Expectation) demandsSSH() bool {
	return len(expec.CustomServices) > 0 || len(expec.UpstartServices) > 0
}
