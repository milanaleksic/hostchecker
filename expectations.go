package hostchecker

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/milanaleksic/hostchecker/failure"
)

// Verifiable represents a named "thing" that can be verified using an expectation and a running context
type Verifiable interface {
	String() string
	CheckExpectation(context *runningContext) []error
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
	Shells          []Shell          `json:"shell"`
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
	return target
}

func (expec *Expectation) String() string {
	return expec.Server
}

// CheckServer is a host verifier entry point: for a known expectation, provide failures if they exist
func (expec *Expectation) CheckServer() (failures []*failure.Failure) {
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
		for _, f := range verifiable.CheckExpectation(context) {
			failures = append(failures, failure.New(expec, verifiable, f))
		}
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
	for _, v := range expec.Shells {
		verifiables = append(verifiables, v)
	}
	return
}

func (expec *Expectation) demandsSSH() bool {
	return len(expec.CustomServices) > 0 || len(expec.UpstartServices) > 0 || len(expec.Shells) > 0
}
