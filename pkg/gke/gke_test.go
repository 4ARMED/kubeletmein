package gke

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"os"
	"testing"

	"cloud.google.com/go/compute/metadata"
	"github.com/4armed/kubeletmein/pkg/common"
	"github.com/4armed/kubeletmein/pkg/mocks"
	"github.com/stretchr/testify/assert"
	yaml "gopkg.in/yaml.v2"
)

var (
	exampleKubeEnv = `CA_CERT: aWFtLi4uLmlycmVsZXZhbnQ=
KUBELET_CERT: aWFtLi4uLmlycmVsZXZhbnQ=
KUBELET_KEY: aWFtLi4uLmlycmVsZXZhbnQ=
KUBERNETES_MASTER_NAME: 1.1.1.1`
)

func TestMetadataFromGKEService(t *testing.T) {
	metadataClient := mocks.NewTestClient(func(req *http.Request) *http.Response {
		assert.Equal(t, "http://169.254.169.254/computeMetadata/v1/instance/attributes/kube-env", req.URL.String(), "should be equal")
		return &http.Response{
			StatusCode: 200,
			Body:       ioutil.NopCloser(bytes.NewBufferString(exampleKubeEnv)),
			Header:     make(http.Header),
		}
	})

	m := metadata.NewClient(metadataClient)
	kubeenv, err := fetchMetadataFromGKEService(m)
	if err != nil {
		t.Errorf("want kubeenv, got %q", err)
	}

	k := Kubeenv{}
	err = yaml.Unmarshal(kubeenv, &k)
	if err != nil {
		t.Errorf("unable to parse YAML from kube-env: %v", err)
	}

	assert.Equal(t, "1.1.1.1", k.KubeMasterName, "they should be equal")

}

func TestMetadataFromGKEFile(t *testing.T) {
	tempFile, err := ioutil.TempFile("", "")
	if err != nil {
		t.Errorf("couldn't create temp file for kube-env: %v", err)
	}
	_, err = tempFile.WriteString(exampleKubeEnv)
	if err != nil {
		t.Errorf("couldn't write kube-env to temp file: %v", err)
	}

	kubeenv, err := common.FetchMetadataFromFile(tempFile.Name())
	if err != nil {
		t.Errorf("want kubeenv, got %q", err)
	}

	// Clean up
	err = os.Remove(tempFile.Name())
	if err != nil {
		t.Errorf("couldn't remove tempFile: %v", err)
	}

	k := Kubeenv{}
	err = yaml.Unmarshal(kubeenv, &k)
	if err != nil {
		t.Errorf("unable to parse YAML from kube-env: %v", err)
	}

	assert.Equal(t, "1.1.1.1", k.KubeMasterName, "they should be equal")
}

func TestBootstrapGkeCmd(t *testing.T) {
	// TODO: Write test for end-to-end
}
