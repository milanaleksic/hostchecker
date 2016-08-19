package main

import "fmt"

type response struct {
	Name     string `json:"name"`
	Url      string `json:"url"`
	Codes    []int `json:"codes"`
	Response string `json:"response"`
	Server   string
}

func (s *response) newFailure(format string, args ...interface{}) *failure {
	return &failure{
		serviceName: s.Name,
		server: s.Server,
		msg: fmt.Sprintf(format, args...),
	}
}
