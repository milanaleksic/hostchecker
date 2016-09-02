package hostchecker

import (
	"fmt"
)

// Shell defines an expectation for a freestyle shell command to be executed remotely
type Shell struct {
	Name     string `json:"name"`
	CLI      string `json:"cli"`
	Expected string `json:"expected"`
	Server   string
}

func (shell *Shell) newFailure(format string, args ...interface{}) *Failure {
	return &Failure{
		serviceName: shell.Name,
		server:      shell.Server,
		msg:         fmt.Sprintf(format, args...),
	}
}

func (shell Shell) String() string {
	return shell.Name
}

// CheckExpectation verifies expectation sent as parameter. It will not use context - no need for remote access
func (shell Shell) CheckExpectation(expec *Expectation, context *runningContext) (failures []Failure) {
	response, err := context.executeRemoteCommand(shell.CLI)
	if err != nil {
		return append(failures, *shell.newFailure(err.Error()))
	} else if shell.Expected != "" && response != shell.Expected {
		return append(failures, *shell.newFailure("Could not match output of process: %s (expected %s)", response, shell.Expected))
	}
	return
}
