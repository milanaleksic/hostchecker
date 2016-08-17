package main

import "fmt"

type failure struct {
	server      string
	serviceName string
	msg         string
}

func (f failure) String() string {
	return fmt.Sprintf("On server %s the service %s check failed with reason: %s", f.server, f.serviceName, f.msg)
}