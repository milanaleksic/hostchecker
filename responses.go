package hostchecker

import (
	"io/ioutil"
	"net/http"
)

// Response defines an expectation for a URL request response, which can be utilized to check if a server is indeed up and properly running
type Response struct {
	Name     string `json:"name"`
	URL      string `json:"url"`
	Codes    []int  `json:"codes"`
	Response string `json:"response"`
}

func (response Response) String() string {
	return response.Name
}

// CheckExpectation verifies expectation sent as parameter. It will not use context - no need for remote access
func (response Response) CheckExpectation(context *runningContext) []error {
	resp, err := http.Get(response.URL)
	if err != nil {
		return only(err)
	}

	codeFound := false
	for _, code := range response.Codes {
		if code == resp.StatusCode {
			codeFound = true
			break
		}
	}
	if !codeFound {
		return onlyF("Code (%d) is not as expected (%+v)", resp.StatusCode, response.Codes)
	}

	if response.Response != "" {
		data, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return only(err)
		}

		if string(data) != response.Response {
			return onlyF("Response (%s) is not as expected (%+v)", data, response.Response)
		}
	}
	return nil
}
