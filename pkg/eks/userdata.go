package eks

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"regexp"

	"github.com/4armed/kubeletmein/pkg/common"
	"github.com/ghodss/yaml"
	api "k8s.io/client-go/tools/clientcmd/api"
)

// File contains details of files to write through cloud-config
type File struct {
	Encoding    string `json:"encoding"`
	Content     string `json:"content"`
	Owner       string `json:"owner"`
	Path        string `json:"path"`
	Permissions string `json:"permissions"`
}

// CloudConfig contains the parsed cloud-config data
type CloudConfig struct {
	WriteFiles []File `json:"write_files"`
}

// ParseCloudConfig parses gzipped cloud-config formatted YAML from cloud-init
func ParseCloudConfig(cloudConfig []byte) (api.Config, error) {

	kubeConfigData := api.Config{}
	var userData CloudConfig
	var caData string

	// Assume our cloud-config is gzipped....we probably should check this
	ungzippedCloudConfig, err := common.GunzipData(cloudConfig)
	if err != nil {
		return kubeConfigData, err
	}

	fmt.Println("about to unmarshal userdata")

	err = yaml.Unmarshal(ungzippedCloudConfig, &userData)
	if err != nil {
		return kubeConfigData, err
	}

	fmt.Println("about to unmarshal /tmp/kubeconfig.yaml")

	testFile, err := ioutil.ReadFile("/tmp/kubeconfig.yaml")
	if err != nil {
		return kubeConfigData, err
	}

	err = yaml.Unmarshal(testFile, &kubeConfigData)
	if err != nil {
		return kubeConfigData, err
	}

	fmt.Println("from file", kubeConfigData)

	for _, v := range userData.WriteFiles {
		if v.Path == "/etc/eksctl/ca.crt" {
			caData = string(v.Content)
		}
		if v.Path == "/etc/eksctl/kubeconfig.yaml" {
			fmt.Println(v.Content)
			err = yaml.Unmarshal([]byte(v.Content), &kubeConfigData)
			if err != nil {
				return kubeConfigData, err
			}
		}

	}

	// fmt.Println("clusters:", kubeConfigData.Clusters)
	fmt.Println(kubeConfigData)
	fmt.Println("caData:", caData)

	// if caData != "" {
	// 	kubeConfigData.Clusters..CertificateAuthority
	// }

	// kubeConfigData.Clusters

	return kubeConfigData, nil
}

// ParseShellScript parses shell-script format user-data seen on managed nodegroups
func ParseShellScript(userData string) (api.Config, error) {
	// userData should contain the following lines:
	// B64_CLUSTER_CA=...
	// API_SERVER_URL=...
	// /etc/eks/bootstrap.sh <CLUSTER_NAME> ...
	re := regexp.MustCompile(`(?m)^B64_CLUSTER_CA=(.*)$`)
	caData := re.FindStringSubmatch(userData)
	if caData == nil {
		return api.Config{}, errors.New("Error while parsing user-data, could not find B64_CLUSTER_CA")
	}

	base64DecodedCAData, err := base64.StdEncoding.DecodeString(caData[1])
	if err != nil {
		return api.Config{}, err
	}

	re = regexp.MustCompile(`(?m)^API_SERVER_URL=(.*)$`)
	k8sMaster := re.FindStringSubmatch(userData)
	if k8sMaster == nil {
		return api.Config{}, errors.New("Error while parsing user-data, could not find API_SERVER_URL")
	}

	re = regexp.MustCompile(`(?m)^/etc/eks/bootstrap.sh\s+(\S+)\s`)
	clusterName := re.FindStringSubmatch(userData)
	if clusterName == nil {
		return api.Config{}, errors.New("Error while parsing user-data, could not find cluster name from bootstrap.sh parameters")
	}

	kubeconfigData := api.Config{
		// Define a cluster stanza
		Clusters: map[string]*api.Cluster{
			clusterName[1]: {
				Server:                   clusterName[1],
				CertificateAuthorityData: base64DecodedCAData,
			},
		},
		// Define auth based on the kubelet client cert retrieved
		AuthInfos: map[string]*api.AuthInfo{
			"kubelet": {
				Exec: &api.ExecConfig{
					APIVersion: "client.authentication.k8s.io/v1alpha1",
					Command:    "aws",
					Args: []string{
						"eks",
						"get-token",
						"--cluster-name",
						clusterName[1],
						"--region",
						"eu-west-1",
					},
					Env: []api.ExecEnvVar{
						{
							Name:  "AWS_STS_REGIONAL_ENDPOINTS",
							Value: "regional",
						},
					},
				},
			},
		},
		// Define a context and set as current
		Contexts: map[string]*api.Context{
			"kubeletmein": {
				Cluster:  clusterName[1],
				AuthInfo: "kubelet",
			},
		},
		CurrentContext: "kubeletmein",
	}

	return kubeconfigData, nil
}
