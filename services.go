package hostchecker

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	serviceListOutputRegex = regexp.MustCompile(`([a-z-]+)[^\d]*(\d+)$`)
	psOutputRegex          = regexp.MustCompile(`([^\s]+)\s+([^\s]+)\s+([^\s]+)`)
)

// UpstartService defines an Upstart service expectation (how recently it should have been started for example)
type UpstartService struct {
	Name             string `json:"name"`
	User             string `json:"user"`
	NewerThanSeconds int    `json:"newerThanSeconds"`
	Ports            []int  `json:"ports"`
}

func (s UpstartService) String() string {
	return s.Name
}

// CheckExpectation verifies expectation sent as parameter. It will use context to execute remote commands
func (s UpstartService) CheckExpectation(context *runningContext) (errors []error) {
	app, err := context.executeRemoteCommand(fmt.Sprintf(`status %s`, s.Name))
	if err != nil {
		return only(err)
	} else if !serviceListOutputRegex.MatchString(app) {
		return onlyF("Could not match output of service listing in '%s'", app)
	}
	found := serviceListOutputRegex.FindStringSubmatch(app)
	pid := found[2]
	if !strings.Contains(app, "running") {
		return onlyF("Service not up")
	}

	userAndServiceStartTime, err := context.executeRemoteCommand(fmt.Sprintf(`ps -p %s -o user,pid,etime | tail -1`, pid))
	if err != nil {
		return only(err)
	} else if !psOutputRegex.MatchString(userAndServiceStartTime) {
		return onlyF("Could not match output of process analysis in '%s' (service down?)", userAndServiceStartTime)
	}
	found = psOutputRegex.FindStringSubmatch(userAndServiceStartTime)
	user := found[1]
	// (we know it already) pid := found[2]
	elapsedTime := found[3]

	if fail := s.checkUser(user); fail != nil {
		errors = append(errors, fail)
	}
	if fail := s.checkPorts(context, pid); fail != nil {
		errors = append(errors, fail)
	}
	if fail := s.checkOld(elapsedTime); fail != nil {
		errors = append(errors, fail)
	}
	return
}

func (s *UpstartService) checkPorts(context *runningContext, pid string) error {
	for _, port := range s.Ports {
		pidHoldingPort, err := context.executeRemoteCommand(fmt.Sprintf(`lsof -nP | grep :%d | grep LISTEN | awk '{print $2}'`, port))
		if err != nil {
			return err
		} else if pidHoldingPort == "" {
			return fmt.Errorf("Port %d is not being taken by any process", port)
		}
		if pidHoldingPort == pid {
			continue
		}
		ppidHoldingPort, err := context.executeRemoteCommand(fmt.Sprintf(`cat /proc/%s/stat | awk '{print $4}'`, pidHoldingPort))
		if err != nil {
			return err
		} else if ppidHoldingPort != pid {
			return fmt.Errorf("Port %d is being taken by the process PID=%s. Neither that PID nor its parent (%s) is of the service (%s)",
				port, pidHoldingPort, ppidHoldingPort, pid)
		}
	}
	return nil
}

func (s *UpstartService) checkOld(elapsedTime string) error {
	if s.NewerThanSeconds != 0 {
		timeInSeconds := extractTimeInSeconds(elapsedTime)
		if timeInSeconds > s.NewerThanSeconds {
			return fmt.Errorf("Service is older than %d seconds (age is %d seconds)", s.NewerThanSeconds, timeInSeconds)
		}
	}
	return nil
}

func (s *UpstartService) checkUser(user string) error {
	if user != s.User {
		return fmt.Errorf("User is not correct for this service: %s != (expected) %s", user, s.User)
	}
	return nil
}

// CustomService defines a process expectation which will be identified via `ps` query for a certain regex
type CustomService struct {
	UpstartService
	Regex string `json:"regex"`
}

// CheckExpectation verifies expectation sent as parameter. It will use context to execute remote commands
func (s CustomService) CheckExpectation(context *runningContext) (errors []error) {
	jettyUsernameAndStartTime, err := context.executeRemoteCommand(fmt.Sprintf(`ps ax -o user,pid,etime,command | grep '%s'`, s.Regex))
	if err != nil {
		return only(err)
	} else if !psOutputRegex.MatchString(jettyUsernameAndStartTime) {
		return onlyF("Custom service has not been found on this server")
	}
	found := psOutputRegex.FindStringSubmatch(jettyUsernameAndStartTime)
	user := found[1]
	pid := found[2]
	elapsedTime := found[3]

	if fail := s.checkUser(user); fail != nil {
		errors = append(errors, fail)
	}
	if fail := s.checkPorts(context, pid); fail != nil {
		errors = append(errors, fail)
	}
	if fail := s.checkOld(elapsedTime); fail != nil {
		errors = append(errors, fail)
	}
	return
}
