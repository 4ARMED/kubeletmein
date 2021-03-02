package mocks

import (
	"net/http"
)

// RoundTripFunc wraps the RoundTrip func...
type RoundTripFunc func(req *http.Request) *http.Response

// RoundTrip implements the RoundTrip func
func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

// NewTestClient returns an http client using the supplied function as transport
func NewTestClient(fn RoundTripFunc) *http.Client {
	return &http.Client{
		Transport: RoundTripFunc(fn),
	}
}
