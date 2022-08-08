package eks

import (
	"path/filepath"
	"testing"

	"github.com/4armed/kubeletmein/pkg/common"
	"github.com/stretchr/testify/assert"
)

var (
	region string = "eu-west-1"
)

func TestGetUserDataCloudConfig(t *testing.T) {
	// Shell script

}

func TestGetUserDataShellScript(t *testing.T) {
	// returns cloud-config gzipped
}

func TestParseUserDataFromFile(t *testing.T) {
	userDataBytes, err := common.FetchMetadataFromFile(filepath.Join("testdata", "userdata-shell.txt"))
	userData := string(userDataBytes)
	if err != nil {
		t.Errorf("err: %v", err)
	}

	kubeConfigData, err := ParseUserData(userData, region)
	if err != nil {
		t.Errorf("err: %v", err)
	}

	assert.Equal(t, "https://36C6589489FA0EE0A77C38D3A0552682.gr7.eu-west-1.eks.amazonaws.com", kubeConfigData.Clusters["test4"].Server, "it should match")
}
