package hostchecker

import (
	"bytes"
	"fmt"
	"log"
	"strings"

	"golang.org/x/crypto/ssh"
)

type runningContext struct {
	sshClient *ssh.Client
}

func newRunningContextWithSSH(username, password, server string) *runningContext {
	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
	}

	client, err := ssh.Dial("tcp", server, config)
	if err != nil {
		log.Fatal("Failed to dial: " + err.Error())
	}
	return &runningContext{
		sshClient: client,
	}
}

func (context *runningContext) executeRemoteCommand(command string) (string, error) {
	if context.sshClient == nil {
		log.Fatalf("Tried to execute command '%s' but SSH client was not set up", command)
	}
	session, err := context.sshClient.NewSession()
	if err != nil {
		log.Fatal("Failed to create session: " + err.Error())
	}
	defer session.Close()

	var b bytes.Buffer
	session.Stdout = &b
	if err := session.Run(command); err != nil {
		return "", fmt.Errorf("Failed to run: %s, err=%s", command, err.Error())
	}
	return strings.TrimSpace(b.String()), nil
}

func (context *runningContext) Close() error {
	if context.sshClient != nil {
		return context.sshClient.Close()
	}
	return nil
}
