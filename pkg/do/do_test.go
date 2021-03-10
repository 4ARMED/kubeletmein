package do

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/4armed/kubeletmein/pkg/common"
	"github.com/4armed/kubeletmein/pkg/mocks"
	metadata "github.com/digitalocean/go-metadata"
	"github.com/ghodss/yaml"
	"github.com/stretchr/testify/assert"
)

var (
	exampleUserData = `k8saas_ca_cert: aWFtLi4uLmlycmVsZXZhbnQ=
k8saas_bootstrap_token: aWFtLi4uLmlycmVsZXZhbnQ=
k8saas_master_domain_name: 1.1.1.1`
)

func TestMetadataFromDOService(t *testing.T) {
	mockClient := mocks.NewTestClient(func(req *http.Request) *http.Response {
		assert.Equal(t, "http://169.254.169.254/metadata/v1/user-data", req.URL.String(), "should be equal")
		return &http.Response{
			StatusCode: 200,
			Body:       ioutil.NopCloser(bytes.NewBufferString(exampleUserData)),
			Header:     make(http.Header),
		}
	})

	metadataClientOptions := metadata.WithHTTPClient(mockClient)
	metadataClient := metadata.NewClient(metadataClientOptions)

	userData, err := fetchMetadataFromDOService(metadataClient)
	if err != nil {
		t.Errorf("want user-data, got %q", err)
	}

	m := Metadata{}
	err = yaml.Unmarshal(userData, &m)
	if err != nil {
		t.Errorf("unable to parse YAML from user-data: %v", err)
	}

	assert.Equal(t, "1.1.1.1", m.KubeMaster, "they should be equal")

}

func TestMetadataFromDOFile(t *testing.T) {
	cwd, err := os.Getwd()
	if err != nil {
		t.Errorf("err: %v", err)
	}

	testFile := filepath.Join(cwd, "testdata", "user-data.txt")
	kubeenv, err := common.FetchMetadataFromFile(testFile)
	if err != nil {
		t.Errorf("want kubeenv, got %q", err)
	}

	m := Metadata{}
	err = yaml.Unmarshal(kubeenv, &m)
	if err != nil {
		t.Errorf("unable to parse YAML from user-data: %v", err)
	}

	assert.Equal(t, "2b2afd78-3773-426b-b67b-cbeb469ed434.internal.k8s.ondigitalocean.com", m.KubeMaster, "they should be equal")
}

func TestBootstrapDOCmd(t *testing.T) {
	// TODO: Write test for end-to-end
}
