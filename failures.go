package main

import "fmt"

type failure struct {
	s   service
	msg string
}

func (f *failure) String() string {
	return fmt.Sprintf("%+v", f)
}