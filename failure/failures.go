package failure

import (
	"fmt"
)

// Failure is a representation of a failed constraint (aka Expectation)
type Failure struct {
	server      string
	serviceName string
	msg         string
}

func (f Failure) String() string {
	return fmt.Sprintf("On server %s the service %s check failed with reason: %s", f.server, f.serviceName, f.msg)
}

// New creates a nee failure to be used by end-user programs to understand what went wrong where
func New(server fmt.Stringer, name fmt.Stringer, errorFmt error) *Failure {
	return &Failure{
		serviceName: name.String(),
		server:      server.String(),
		msg:         errorFmt.Error(),
	}
}
