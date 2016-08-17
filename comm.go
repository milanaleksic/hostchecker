package main

import (
	"bytes"
	"strings"
	"golang.org/x/crypto/ssh"
	"fmt"
	"log"
	"regexp"
)

var (
	serviceListOutputRegex = regexp.MustCompile(`([a-z-]+)[^\d]*(\d+)$`)
	psOutputRegex = regexp.MustCompile(`([^\s]+)\s+([^\s]+)\s+([^\s]+)`)
)

func executeRemoteCommand(client *ssh.Client, command string) string {
	session, err := client.NewSession()
	if err != nil {
		log.Fatal("Failed to create session: " + err.Error())
	}
	defer session.Close()

	var b bytes.Buffer
	session.Stdout = &b
	if err := session.Run(command); err != nil {
		log.Fatalf("Failed to run: %s, err=%s", command, err.Error())
	}
	return strings.TrimSpace(b.String())
}

func checkServer(expec expectation) {
	config := &ssh.ClientConfig{
		User: expec.User,
		Auth: []ssh.AuthMethod{
			ssh.Password(expec.Password),
		},
	}

	client, err := ssh.Dial("tcp", expec.Server, config)
	if err != nil {
		log.Fatal("Failed to dial: " + err.Error())
	}

	fmt.Printf("\nChecking services on host %s\n", expec.Server)
	for _, upstartService := range expec.UpstartServices {
		app := executeRemoteCommand(client, fmt.Sprintf(`status %s`, upstartService.Name))
		if !serviceListOutputRegex.MatchString(app) {
			log.Fatalf("Could not match output of service listing in '%s'\n", app)
		}
		found := serviceListOutputRegex.FindStringSubmatch(app)
		serviceName := found[1]
		pid := found[2]
		if !strings.Contains(app, "running") {
			log.Fatalf("Service not up: %s\n", serviceName)
		}

		userAndServiceStartTime := executeRemoteCommand(client, fmt.Sprintf(`ps -p %s -o user,pid,etime | tail -1`, pid))
		if !psOutputRegex.MatchString(userAndServiceStartTime) {
			log.Fatalf("Could not match output of process analysis in '%s' (service down?)\n", userAndServiceStartTime)
		}
		found = psOutputRegex.FindStringSubmatch(userAndServiceStartTime)
		user := found[1]
		// (we know it already) pid := found[2]
		elapsedTime := found[3]

		fmt.Printf("Service %s is up, has pid: %s, is executed under user %s and is running for (HH:MM:SS) %s\n", serviceName, pid, user, elapsedTime)

		upstartService.checkUser(user)
		upstartService.checkPorts(client, pid)
		upstartService.checkOld(elapsedTime)

	}
	for _, customService := range expec.CustomServices {
		jettyUsernameAndStartTime := executeRemoteCommand(client, fmt.Sprintf(`ps ax -o user,pid,etime,command | grep '%s'`, customService.Regex))
		if !psOutputRegex.MatchString(jettyUsernameAndStartTime) {
			log.Fatalf("Custom service %s has not been found on this server", customService.Name)
		}
		found := psOutputRegex.FindStringSubmatch(jettyUsernameAndStartTime)
		user := found[1]
		pid := found[2]
		elapsedTime := found[3]

		fmt.Printf("Custom service %s is up, has pid: %s, is executed under user %s and is running for (HH:MM:SS) %s\n", customService.Name, pid, user, elapsedTime)

		customService.checkUser(user)
		customService.checkPorts(client, pid)
		customService.checkOld(elapsedTime)
	}
}