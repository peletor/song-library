package tests

import (
	"github.com/gavv/httpexpect/v2"
	"net/url"
	"testing"
)

const (
	host = "localhost:8080"
)

func httpExpect(t *testing.T) *httpexpect.Expect {
	hostURL := url.URL{
		Scheme: "http",
		Host:   host,
	}
	return httpexpect.Default(t, hostURL.String())
}
