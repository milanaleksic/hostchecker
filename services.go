package hostchecker

import (
	"fmt"
)

// Service defines an Upstart service expectation (how recently it should have been started for example)
type Service struct {
	Name             string `json:"name"`
	User             string `json:"user"`
	NewerThanSeconds int    `json:"newerThanSeconds"`
	Ports            []int  `json:"ports"`
	Server           string
}

// CustomService defines a process expectation which will be identified via `ps` query for a certain regex
type CustomService struct {
	Service
	Regex string `json:"regex"`
}

func (s *Service) newFailure(format string, args ...interface{}) *Failure {
	return &Failure{
		serviceName: s.Name,
		server:      s.Server,
		msg:         fmt.Sprintf(format, args...),
	}
}

func (s *Service) checkPorts(executeRemoteCommand func(string) (string, error), pid string) *Failure {
	for _, port := range s.Ports {
		pidHoldingPort, err := executeRemoteCommand(fmt.Sprintf(`lsof -nP | grep :%d | grep LISTEN | awk '{print $2}'`, port))
		if err != nil {
			return s.newFailure(err.Error())
		} else if pidHoldingPort == "" {
			return s.newFailure("Port %d is not being taken by any process", port)
		}
		if pidHoldingPort == pid {
			continue
		}
		ppidHoldingPort, err := executeRemoteCommand(fmt.Sprintf(`cat /proc/%s/stat | awk '{print $4}'`, pidHoldingPort))
		if err != nil {
			return s.newFailure(err.Error())
		} else if ppidHoldingPort != pid {
			return s.newFailure("Port %d is being taken by the process PID=%s. Neither that PID nor its parent (%s) is of the service (%s)",
				port, pidHoldingPort, ppidHoldingPort, pid)
		}
	}
	return nil
}

func (s *Service) checkOld(elapsedTime string) *Failure {
	if s.NewerThanSeconds != 0 {
		timeInSeconds := extractTimeInSeconds(elapsedTime)
		if timeInSeconds > s.NewerThanSeconds {
			return s.newFailure("Service is older than %d seconds (age is %d seconds)", s.NewerThanSeconds, timeInSeconds)
		}
	}
	return nil
}

func (s *Service) checkUser(user string) *Failure {
	if user != s.User {
		return s.newFailure("User is not correct for this service: %s != (expected) %s", user, s.User)
	}
	return nil
}
