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
	Server           string
}

func (s UpstartService) String() string {
	return s.Name
}

// CheckExpectation verifies expectation sent as parameter. It will use context to execute remote commands
func (s UpstartService) CheckExpectation(expec *Expectation, context *runningContext) (failures []Failure) {
	app, err := context.executeRemoteCommand(fmt.Sprintf(`status %s`, s.Name))
	if err != nil {
		return append(failures, *s.newFailure(err.Error()))
	} else if !serviceListOutputRegex.MatchString(app) {
		return append(failures, *s.newFailure("Could not match output of service listing in '%s'", app))
	}
	found := serviceListOutputRegex.FindStringSubmatch(app)
	pid := found[2]
	if !strings.Contains(app, "running") {
		return append(failures, *s.newFailure("Service not up"))
	}

	userAndServiceStartTime, err := context.executeRemoteCommand(fmt.Sprintf(`ps -p %s -o user,pid,etime | tail -1`, pid))
	if err != nil {
		return append(failures, *s.newFailure(err.Error()))
	} else if !psOutputRegex.MatchString(userAndServiceStartTime) {
		return append(failures, *s.newFailure("Could not match output of process analysis in '%s' (service down?)", userAndServiceStartTime))
	}
	found = psOutputRegex.FindStringSubmatch(userAndServiceStartTime)
	user := found[1]
	// (we know it already) pid := found[2]
	elapsedTime := found[3]

	if failure := s.checkUser(user); failure != nil {
		failures = append(failures, *failure)
	}
	if failure := s.checkPorts(context, pid); failure != nil {
		failures = append(failures, *failure)
	}
	if failure := s.checkOld(elapsedTime); failure != nil {
		failures = append(failures, *failure)
	}
	return
}

func (s *UpstartService) newFailure(format string, args ...interface{}) *Failure {
	return &Failure{
		serviceName: s.Name,
		server:      s.Server,
		msg:         fmt.Sprintf(format, args...),
	}
}

func (s *UpstartService) checkPorts(context *runningContext, pid string) *Failure {
	for _, port := range s.Ports {
		pidHoldingPort, err := context.executeRemoteCommand(fmt.Sprintf(`lsof -nP | grep :%d | grep LISTEN | awk '{print $2}'`, port))
		if err != nil {
			return s.newFailure(err.Error())
		} else if pidHoldingPort == "" {
			return s.newFailure("Port %d is not being taken by any process", port)
		}
		if pidHoldingPort == pid {
			continue
		}
		ppidHoldingPort, err := context.executeRemoteCommand(fmt.Sprintf(`cat /proc/%s/stat | awk '{print $4}'`, pidHoldingPort))
		if err != nil {
			return s.newFailure(err.Error())
		} else if ppidHoldingPort != pid {
			return s.newFailure("Port %d is being taken by the process PID=%s. Neither that PID nor its parent (%s) is of the service (%s)",
				port, pidHoldingPort, ppidHoldingPort, pid)
		}
	}
	return nil
}

func (s *UpstartService) checkOld(elapsedTime string) *Failure {
	if s.NewerThanSeconds != 0 {
		timeInSeconds := extractTimeInSeconds(elapsedTime)
		if timeInSeconds > s.NewerThanSeconds {
			return s.newFailure("Service is older than %d seconds (age is %d seconds)", s.NewerThanSeconds, timeInSeconds)
		}
	}
	return nil
}

func (s *UpstartService) checkUser(user string) *Failure {
	if user != s.User {
		return s.newFailure("User is not correct for this service: %s != (expected) %s", user, s.User)
	}
	return nil
}

// CustomService defines a process expectation which will be identified via `ps` query for a certain regex
type CustomService struct {
	UpstartService
	Regex string `json:"regex"`
}

// CheckExpectation verifies expectation sent as parameter. It will use context to execute remote commands
func (s CustomService) CheckExpectation(expec *Expectation, context *runningContext) (failures []Failure) {
	jettyUsernameAndStartTime, err := context.executeRemoteCommand(fmt.Sprintf(`ps ax -o user,pid,etime,command | grep '%s'`, s.Regex))
	if err != nil {
		return append(failures, *s.newFailure(err.Error()))
	} else if !psOutputRegex.MatchString(jettyUsernameAndStartTime) {
		return append(failures, *s.newFailure("Custom service has not been found on this server"))
	}
	found := psOutputRegex.FindStringSubmatch(jettyUsernameAndStartTime)
	user := found[1]
	pid := found[2]
	elapsedTime := found[3]

	if failure := s.checkUser(user); failure != nil {
		failures = append(failures, *failure)
	}
	if failure := s.checkPorts(context, pid); failure != nil {
		failures = append(failures, *failure)
	}
	if failure := s.checkOld(elapsedTime); failure != nil {
		failures = append(failures, *failure)
	}
	return
}
