package eks

import (
	"encoding/base64"
	"errors"
	"regexp"

	"github.com/4armed/kubeletmein/pkg/common"
	"github.com/ghodss/yaml"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
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
// As a kubelet kubeconfig file is provided we basically pull this out as-is
// from cloud-config but merge in our CA data as `certificate-authority-data`
// to save us having to write out a cert file.
func ParseCloudConfig(cloudConfig []byte) (*clientcmdapi.Config, error) {

	var cloudConfigData []byte
	k := &clientcmdapi.Config{}
	userData := CloudConfig{}
	var caData string

	// cloud-config is probably gzipped but let's check
	gzipped, err := common.IsGzipped(cloudConfig)
	if err != nil {
		return nil, err
	}

	if gzipped {
		cloudConfigData, err = common.GunzipData(cloudConfig)
		if err != nil {
			return nil, err
		}
	} else {
		cloudConfigData = cloudConfig
	}

	err = yaml.Unmarshal(cloudConfigData, &userData)
	if err != nil {
		return nil, err
	}

	for _, v := range userData.WriteFiles {
		if v.Path == "/etc/eksctl/ca.crt" {
			caData = v.Content
		}
		if v.Path == "/etc/eksctl/kubeconfig.yaml" {
			k, err = clientcmd.Load([]byte(v.Content))
			if err != nil {
				return nil, err
			}
		}
	}

	if caData != "" {
		contextName := k.CurrentContext
		clusterName := k.Contexts[contextName].Cluster
		k.Clusters[clusterName].CertificateAuthorityData = []byte(caData)
	}

	return k, nil
}

// ParseShellScript parses shell-script format user-data seen on managed nodegroups
func ParseShellScript(userData string) (*clientcmdapi.Config, error) {
	// userData should contain the following lines:
	// B64_CLUSTER_CA=...
	// API_SERVER_URL=...
	// /etc/eks/bootstrap.sh <CLUSTER_NAME> ...
	re := regexp.MustCompile(`(?m)^B64_CLUSTER_CA=(.*)$`)
	caData := re.FindStringSubmatch(userData)
	if caData == nil {
		return nil, errors.New("Error while parsing user-data, could not find B64_CLUSTER_CA")
	}

	base64DecodedCAData, err := base64.StdEncoding.DecodeString(caData[1])
	if err != nil {
		return nil, err
	}

	re = regexp.MustCompile(`(?m)^API_SERVER_URL=(.*)$`)
	k8sMaster := re.FindStringSubmatch(userData)
	if k8sMaster == nil {
		return nil, errors.New("Error while parsing user-data, could not find API_SERVER_URL")
	}

	re = regexp.MustCompile(`(?m)^/etc/eks/bootstrap.sh\s+(\S+)\s`)
	clusterName := re.FindStringSubmatch(userData)
	if clusterName == nil {
		return nil, errors.New("Error while parsing user-data, could not find cluster name from bootstrap.sh parameters")
	}

	kubeconfigData := &clientcmdapi.Config{
		// Define a cluster stanza
		Clusters: map[string]*clientcmdapi.Cluster{
			clusterName[1]: {
				Server:                   clusterName[1],
				CertificateAuthorityData: base64DecodedCAData,
			},
		},
		// Define auth based on the kubelet client cert retrieved
		AuthInfos: map[string]*clientcmdapi.AuthInfo{
			"kubelet": {
				Exec: &clientcmdapi.ExecConfig{
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
					Env: []clientcmdapi.ExecEnvVar{
						{
							Name:  "AWS_STS_REGIONAL_ENDPOINTS",
							Value: "regional",
						},
					},
				},
			},
		},
		// Define a context and set as current
		Contexts: map[string]*clientcmdapi.Context{
			"kubeletmein": {
				Cluster:  clusterName[1],
				AuthInfo: "kubelet",
			},
		},
		CurrentContext: "kubeletmein",
	}

	return kubeconfigData, nil
}
