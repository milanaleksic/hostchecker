package main

import (
	"golang.org/x/crypto/ssh"
	"fmt"
)

type service struct {
	Name             string `json:"name"`
	User             string `json:"user"`
	NewerThanSeconds int `json:"newerThanSeconds"`
	Ports            []int  `json:"ports"`
}

type customService struct {
	service
	Regex string `json:"regex"`
}

func (s *service) checkPorts(client *ssh.Client, pid string) {
	for _, port := range s.Ports {
		pidHoldingPort := executeRemoteCommand(client, fmt.Sprintf(`lsof -nP | grep :%d | grep LISTEN | awk '{print $2}'`, port))
		if pidHoldingPort == "" {
			log.Fatalf("Port %d is not being taken by any process", port)
		}
		if pidHoldingPort == pid {
			continue
		}
		ppidHoldingPort := executeRemoteCommand(client, fmt.Sprintf(`cat /proc/%s/stat | awk '{print $4}'`, pidHoldingPort))
		if ppidHoldingPort != pid {
			log.Fatalf("Port %d is being taken by the process PID=%s. Neither that PID nor its parent (%s) is of the service %s (%s)",
				port, pidHoldingPort, ppidHoldingPort, s.Name, pid)
		}
	}
}

func (s *service) checkOld(elapsedTime string) {
	if s.NewerThanSeconds != 0 {
		timeInSeconds := extractTimeInSeconds(elapsedTime)
		if timeInSeconds > s.NewerThanSeconds {
			log.Fatalf("Service %s is older than %d seconds (age is %d seconds)", s.Name, s.NewerThanSeconds, timeInSeconds)
		}
	}
}

func (s *service) checkUser(user string) {
	if user != s.User {
		log.Fatalf("User is not correct for this service: %s != (expected) %s", user, s.User)
	}
}