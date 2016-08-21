package hostchecker

import (
	"bytes"
	"strings"
	"golang.org/x/crypto/ssh"
	"fmt"
	"log"
	"regexp"
	"net/http"
	"io/ioutil"
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

// CheckServer is a host verifier entry point: for a known expectation, provide failures if they exist
func CheckServer(expec Expectation) (failures []Failure) {
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
	failures = append(failures, checkResponses(expec)...)

	return
}

func checkUpstartServices(expec Expectation, client *ssh.Client) (failures []Failure) {
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

		if failure := upstartService.checkUser(user); failure != nil {
			failures = append(failures, *failure)
		}
		if failure := upstartService.checkPorts(client, pid); failure != nil {
			failures = append(failures, *failure)
		}
		if failure := upstartService.checkOld(elapsedTime); failure != nil {
			failures = append(failures, *failure)
		}
	}
	return
}

func checkCustomServices(expec Expectation, client *ssh.Client) (failures []Failure) {
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

		if failure := customService.checkUser(user); failure != nil {
			failures = append(failures, *failure)
		}
		if failure := customService.checkPorts(client, pid); failure != nil {
			failures = append(failures, *failure)
		}
		if failure := customService.checkOld(elapsedTime); failure != nil {
			failures = append(failures, *failure)
		}
	}
	return
}

func checkResponses(expec Expectation) (failures []Failure) {
	for _, response := range expec.Responses {
		fmt.Printf("Checking Response %s\n", response.Name)

		resp, err := http.Get(response.URL)
		if err != nil {
			failures = append(failures, *response.newFailure(err.Error()))
			continue
		}

		codeFound := false;
		for _, code := range response.Codes {
			if code == resp.StatusCode {
				codeFound = true
				break
			}
		}
		if !codeFound {
			failures = append(failures, *response.newFailure("Code (%d) is not as expected (%+v)", resp.StatusCode, response.Codes))
			continue
		}

		if response.Response != "" {
			data, err := ioutil.ReadAll(resp.Body)
			resp.Body.Close()
			if err != nil {
				failures = append(failures, *response.newFailure(err.Error()))
				continue
			}

			if string(data) != response.Response {
				failures = append(failures, *response.newFailure("Response (%s) is not as expected (%+v)", data, response.Response))
				continue
			}
		}
	}
	return
}
