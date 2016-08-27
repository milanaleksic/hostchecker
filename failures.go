package hostchecker

import "fmt"

// Failure is a representation of a failed constraint (aka Expectation)
type Failure struct {
	server      string
	serviceName string
	msg         string
}

func (f Failure) String() string {
	return fmt.Sprintf("On server %s the service %s check failed with reason: %s", f.server, f.serviceName, f.msg)
}
