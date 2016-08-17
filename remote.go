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

func executeRemoteCommand(client *ssh.Client, command string) (string, error) {
	session, err := client.NewSession()
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

func checkServer(expec expectation) (failures []failure) {
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
	failures = append(failures, checkUpstartServices(expec, client)...)
	failures = append(failures, checkCustomServices(expec, client)...)

	return
}

func checkUpstartServices(expec expectation, client *ssh.Client) (failures []failure) {
	for _, upstartService := range expec.UpstartServices {
		fmt.Printf("Checking upstart service %s\n", upstartService.Name)
		app, err := executeRemoteCommand(client, fmt.Sprintf(`status %s`, upstartService.Name))
		if err != nil {
			failures = append(failures, *upstartService.newFailure(err.Error()))
			continue
		} else if !serviceListOutputRegex.MatchString(app) {
			failures = append(failures, *upstartService.newFailure("Could not match output of service listing in '%s'", app))
			continue
		}
		found := serviceListOutputRegex.FindStringSubmatch(app)
		pid := found[2]
		if !strings.Contains(app, "running") {
			failures = append(failures, *upstartService.newFailure("Service not up"))
			continue
		}

		userAndServiceStartTime, err := executeRemoteCommand(client, fmt.Sprintf(`ps -p %s -o user,pid,etime | tail -1`, pid))
		if err != nil {
			failures = append(failures, *upstartService.newFailure(err.Error()))
			continue
		} else if !psOutputRegex.MatchString(userAndServiceStartTime) {
			failures = append(failures, *upstartService.newFailure("Could not match output of process analysis in '%s' (service down?)", userAndServiceStartTime))
			continue
		}
		found = psOutputRegex.FindStringSubmatch(userAndServiceStartTime)
		user := found[1]
		// (we know it already) pid := found[2]
		elapsedTime := found[3]

		upstartService.checkUser(user)
		upstartService.checkPorts(client, pid)
		upstartService.checkOld(elapsedTime)
	}
	return
}

func checkCustomServices(expec expectation, client *ssh.Client) (failures []failure) {
	for _, customService := range expec.CustomServices {
		fmt.Printf("Checking custom service %s\n", customService.Name)

		jettyUsernameAndStartTime, err := executeRemoteCommand(client, fmt.Sprintf(`ps ax -o user,pid,etime,command | grep '%s'`, customService.Regex))
		if err != nil {
			failures = append(failures, *customService.newFailure(err.Error()))
			continue
		} else if !psOutputRegex.MatchString(jettyUsernameAndStartTime) {
			failures = append(failures, *customService.newFailure("Custom service has not been found on this server"))
			continue
		}
		found := psOutputRegex.FindStringSubmatch(jettyUsernameAndStartTime)
		user := found[1]
		pid := found[2]
		elapsedTime := found[3]

		customService.checkUser(user)
		customService.checkPorts(client, pid)
		customService.checkOld(elapsedTime)
	}
	return
}