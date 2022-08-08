package eks

import (
	"path/filepath"
	"testing"

	"github.com/4armed/kubeletmein/pkg/common"
	"github.com/stretchr/testify/assert"
)

func TestParseCloudConfigGzip(t *testing.T) {

	cloudConfigData, err := common.FetchMetadataFromFile(filepath.Join("testdata", "cloud-config.txt.gz"))
	if err != nil {
		t.Errorf("err: %v", err)
	}

	kubeConfig, err := ParseCloudConfig(cloudConfigData, region)
	if err != nil {
		t.Errorf("err: %v", err)
	}

	// if you change the testdata, make sure the cluster name matches here
	assert.Contains(t, kubeConfig.Clusters, "test3.eu-west-1.eksctl.io")
}

func TestParseCloudConfigGzip2(t *testing.T) {

	cloudConfigData, err := common.FetchMetadataFromFile(filepath.Join("testdata", "cloud-config-two.txt.gz"))
	if err != nil {
		t.Errorf("err: %v", err)
	}

	kubeConfig, err := ParseCloudConfig(cloudConfigData, region)
	if err != nil {
		t.Errorf("err: %v", err)
	}

	// if you change the testdata, make sure the cluster name matches here
	assert.Contains(t, kubeConfig.Clusters, "test3")
}

func TestParseCloudConfigNoGzip(t *testing.T) {

	cloudConfigData, err := common.FetchMetadataFromFile(filepath.Join("testdata", "cloud-config.txt.gz"))
	if err != nil {
		t.Errorf("err: %v", err)
	}

	// gunzip CloudConfigData
	gzipCloudCloudConfigData, err := common.GunzipData(cloudConfigData)
	if err != nil {
		t.Errorf("err: %v", err)
	}

	kubeConfig, err := ParseCloudConfig(gzipCloudCloudConfigData, region)
	if err != nil {
		t.Errorf("err: %v", err)
	}

	// if you change the testdata, make sure the cluster name matches here
	assert.Contains(t, kubeConfig.Clusters, "test3.eu-west-1.eksctl.io")
}

func TestParseShellScriptManagementConsole(t *testing.T) {

	userData, err := common.FetchMetadataFromFile(filepath.Join("testdata", "userdata-shell.txt"))
	if err != nil {
		t.Errorf("err: %v", err)
	}

	kubeConfig, err := ParseShellScript(string(userData))
	if err != nil {
		t.Errorf("err: %v", err)
	}

	// if you change the testdata, make sure the cluster name matches here
	assert.Contains(t, kubeConfig.Clusters, "test4")

}

func TestParseShellScriptCustom(t *testing.T) {

	userData, err := common.FetchMetadataFromFile(filepath.Join("testdata", "userdata-custom.txt"))
	if err != nil {
		t.Errorf("err: %v", err)
	}

	kubeConfig, err := ParseShellScript(string(userData))
	if err != nil {
		t.Errorf("err: %v", err)
	}

	// if you change the testdata, make sure the cluster name matches here
	assert.Contains(t, kubeConfig.Clusters, "kubeletmein")

}
