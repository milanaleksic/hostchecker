package hostchecker

// Shell defines an expectation for a freestyle shell command to be executed remotely
type Shell struct {
	Name     string `json:"name"`
	CLI      string `json:"cli"`
	Expected string `json:"expected"`
}

func (shell Shell) String() string {
	return shell.Name
}

// CheckExpectation verifies expectation sent as parameter. It will not use context - no need for remote access
func (shell Shell) CheckExpectation(context *runningContext) []error {
	response, err := context.executeRemoteCommand(shell.CLI)
	if err != nil {
		return only(err)
	} else if shell.Expected != "" && response != shell.Expected {
		return onlyF("Could not match output of process: %s (expected %s)", response, shell.Expected)
	}
	return nil
}
