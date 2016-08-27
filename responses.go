package hostchecker

import "fmt"

// Response defines an expectation for a URL request response, which can be utilized to check if a server is indeed up and properly running
type Response struct {
	Name     string `json:"name"`
	URL      string `json:"url"`
	Codes    []int  `json:"codes"`
	Response string `json:"response"`
	Server   string
}

func (s *Response) newFailure(format string, args ...interface{}) *Failure {
	return &Failure{
		serviceName: s.Name,
		server:      s.Server,
		msg:         fmt.Sprintf(format, args...),
	}
}
