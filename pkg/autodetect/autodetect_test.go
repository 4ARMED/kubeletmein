package autodetect

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testServer struct {
	t             *testing.T
	cloudProvider string
}

func newMetadataServer(t *testing.T, testServer *testServer) *httptest.Server {
	mux := http.NewServeMux()

	switch testServer.cloudProvider {
	case "gke":
		// Google Cloud
		mux.HandleFunc("/", testServer.gkeHandler)
	case "do":
		// Digital Ocean
		mux.HandleFunc("/metadata/v1/id", testServer.doHandler)
	case "eks":
		// AWS EKS
		mux.HandleFunc("/", testServer.eksHandler)
	}

	return httptest.NewServer(mux)
}

func (s *testServer) gkeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Metadata-Flavor", "Google")
	w.Header().Set("Server", "Metadata Server for VM")

	responseData := `0.1/
	computeMetadata/`

	w.Write([]byte(responseData))
}

func (s *testServer) doHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-type", "text/plain; charset=utf-8")

	responseData := `235891500`

	w.Write([]byte(responseData))
}

func (s *testServer) eksHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Server", "EC2ws")

	responseData := `1.0
	2007-01-19
	2007-03-01
	2007-08-29
	2007-10-10
	2007-12-15
	2008-02-01
	2008-09-01
	2009-04-04
	2011-01-01
	2011-05-01
	2012-01-12
	2014-02-25
	2014-11-05
	2015-10-20
	2016-04-19
	2016-06-30
	2016-09-02
	2018-03-28
	2018-08-17
	2018-09-24
	2019-10-01
	2020-10-27
	latest`

	w.Write([]byte(responseData))
}

func TestAutoDetectGKE(t *testing.T) {
	ts := &testServer{
		t:             t,
		cloudProvider: "gke",
	}

	server := newMetadataServer(t, ts)
	defer server.Close()

	metadataServerURL = server.URL

	pc, err := New(&http.Client{}, nil)
	if err != nil {
		t.Errorf("err: %v", err)
	}

	provider := pc.GetProvider()

	assert.Equal(t, "gke", provider, "should be equal")
}

func TestCheckProviderGKE(t *testing.T) {
	ts := &testServer{
		t:             t,
		cloudProvider: "gke",
	}

	server := newMetadataServer(t, ts)
	defer server.Close()

	metadataServerURL = server.URL
	hc := &http.Client{}
	provider := PublicCloudProviders["gke"]
	result := checkProvider(hc, provider)

	assert.True(t, result, "should be true")
}

func TestAutoDetectDigitalOcean(t *testing.T) {
	ts := &testServer{
		t:             t,
		cloudProvider: "do",
	}

	server := newMetadataServer(t, ts)
	defer server.Close()

	metadataServerURL = server.URL

	pc, err := New(&http.Client{}, nil)
	if err != nil {
		t.Errorf("err: %v", err)
	}

	provider := pc.GetProvider()

	assert.Equal(t, "do", provider, "should be equal")

}

func TestCheckProviderDigitalOcean(t *testing.T) {
	ts := &testServer{
		t:             t,
		cloudProvider: "do",
	}

	server := newMetadataServer(t, ts)
	defer server.Close()

	metadataServerURL = server.URL
	hc := &http.Client{}
	provider := PublicCloudProviders["do"]
	result := checkProvider(hc, provider)

	assert.True(t, result, "should be true")
}

func TestAutoDetectEKS(t *testing.T) {
	ts := &testServer{
		t:             t,
		cloudProvider: "eks",
	}

	server := newMetadataServer(t, ts)
	defer server.Close()

	metadataServerURL = server.URL

	pc, err := New(&http.Client{}, nil)
	if err != nil {
		t.Errorf("err: %v", err)
	}

	provider := pc.GetProvider()

	assert.Equal(t, "eks", provider, "should be equal")

}

func TestCheckProviderEKS(t *testing.T) {
	ts := &testServer{
		t:             t,
		cloudProvider: "eks",
	}

	server := newMetadataServer(t, ts)
	defer server.Close()

	metadataServerURL = server.URL
	hc := &http.Client{}
	provider := PublicCloudProviders["eks"]
	result := checkProvider(hc, provider)

	assert.True(t, result, "should be true")
}
