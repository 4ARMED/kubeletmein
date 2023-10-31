package eks

import (
	"encoding/base64"
	"errors"
	"regexp"
	"strings"

	"github.com/4armed/kubeletmein/pkg/common"
	"github.com/ghodss/yaml"
	"github.com/integrii/flaggy"
	"github.com/kubicorn/kubicorn/pkg/logger"
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

// ParseUserData takes a string input metadata, works out its type
// and returns a *clientcmdapi.Config ready for marshalling to disk.
func ParseUserData(userData, region string) (*clientcmdapi.Config, error) {

	var kubeConfigData *clientcmdapi.Config

	// Firstly, is it compressed?
	isGzipped, err := common.IsGzipped([]byte(userData))
	if err != nil {
		return nil, err
	}

	if isGzipped {
		uncompressedData, err := common.GunzipData([]byte(userData))
		if err != nil {
			return nil, err
		}

		// update userData
		userData = string(uncompressedData)
	}

	// Is this a cloud-config?
	re := regexp.MustCompile(`^#cloud-config`)
	match := re.MatchString(userData)
	if match {
		logger.Debug("assuming gzipped cloud-config")
		kubeConfigData, err = ParseCloudConfig([]byte(userData), region)
		if err != nil {
			return nil, err
		}
	} else {
		logger.Debug("shell script assuming, looking for /etc/eks/bootstrap.sh")
		kubeConfigData, err = ParseShellScript(userData, region)
		if err != nil {
			return nil, err
		}
	}

	return kubeConfigData, nil
}

// ParseCloudConfig parses gzipped cloud-config formatted YAML from cloud-init
// As a kubelet kubeconfig file is provided we basically pull this out as-is
// from cloud-config but merge in our CA data as `certificate-authority-data`
// to save us having to write out a cert file.
func ParseCloudConfig(cloudConfig []byte, region string) (*clientcmdapi.Config, error) {

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
		if v.Path == "/etc/eksctl/kubelet.env" {

			kubeletEnvContent := v.Content

			re := regexp.MustCompile(`CLUSTER_NAME=(.*)`)
			fetchedValue := re.FindStringSubmatch(kubeletEnvContent)

			if len(fetchedValue) == 0 {
				// different format of cloud-config
				continue
			}

			clusterName := fetchedValue[1]

			re = regexp.MustCompile(`API_SERVER_URL=(.*)`)
			fetchedValue = re.FindStringSubmatch(kubeletEnvContent)
			k8sMaster := fetchedValue[1]

			re = regexp.MustCompile(`B64_CLUSTER_CA=(.*)`)
			fetchedValue = re.FindStringSubmatch(kubeletEnvContent)
			base64EncodedCA := fetchedValue[1]

			base64DecodedCAData, err := base64.StdEncoding.DecodeString(base64EncodedCA)
			if err != nil {
				return nil, err
			}

			k = &clientcmdapi.Config{
				// Define a cluster stanza
				Clusters: map[string]*clientcmdapi.Cluster{
					clusterName: {
						Server:                   k8sMaster,
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
								clusterName,
								"--region",
								region,
							},
						},
					},
				},
				// Define a context and set as current
				Contexts: map[string]*clientcmdapi.Context{
					"kubeletmein": {
						Cluster:  clusterName,
						AuthInfo: "kubelet",
					},
				},
				CurrentContext: "kubeletmein",
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
func ParseShellScript(userData string, region string) (*clientcmdapi.Config, error) {
	// We must account for all arguments to bootstrap.sh in order to find the non-flag based cluster name
	// https://github.com/awslabs/amazon-eks-ami/blob/master/files/bootstrap.sh

	var clusterName = ""
	var useMaxPods = ""
	var b64ClusterCa = ""
	var apiServerEndpoint = ""
	var kubeletExtraArgs = ""
	var enableDockerBridge = ""
	var awsAPIRetryAttempts = ""
	var dockerConfigJSON = ""
	var pauseContainerAccount = ""
	var pauseContainerVersion = ""
	var dnsClusterIP = ""

	// We are looking for the /etc/eks/boostrap.sh command somewhere in the user-data.
	clusterNameAtStart := regexp.MustCompile(`(?m)/etc/eks/bootstrap.sh ([a-z0-9A-Z-_]*)\s*(--.*)`)
	eksBootstrapCmd := clusterNameAtStart.FindStringSubmatch(userData)
	if eksBootstrapCmd == nil {
		return nil, errors.New("error while parsing user-data, could not find /etc/eks/boostrap.sh")
	}

	eksBootstrapArgs := strings.Fields(eksBootstrapCmd[2])

	if eksBootstrapCmd[1] == "" {
		// The cluster name must be at the end
		clusterName = eksBootstrapArgs[len(eksBootstrapArgs)-1]

		// Now remove it
		eksBootstrapArgs[len(eksBootstrapArgs)-1] = ""
		eksBootstrapArgs = eksBootstrapArgs[:len(eksBootstrapArgs)-1]

	} else {
		clusterName = eksBootstrapCmd[1]
	}

	flaggy.ResetParser()

	flaggy.String(&useMaxPods, "", "use-max-pods", "")
	flaggy.String(&b64ClusterCa, "", "b64-cluster-ca", "")
	flaggy.String(&apiServerEndpoint, "", "apiserver-endpoint", "")
	flaggy.String(&kubeletExtraArgs, "", "kubelet-extra-args", "")
	flaggy.String(&enableDockerBridge, "", "enable-docker-bridge", "")
	flaggy.String(&awsAPIRetryAttempts, "", "aws-api-retry-attempts", "")
	flaggy.String(&dockerConfigJSON, "", "docker-config-json", "")
	flaggy.String(&pauseContainerAccount, "", "pause-container-account", "")
	flaggy.String(&pauseContainerVersion, "", "pause-container-version", "")
	flaggy.String(&dnsClusterIP, "", "dns-cluster-ip", "")

	flaggy.DefaultParser.ShowHelpOnUnexpected = false
	flaggy.ParseArgs(eksBootstrapArgs)

	logger.Debug("b64ClusterCa: %s", b64ClusterCa)
	base64DecodedCAData, err := base64.StdEncoding.DecodeString(checkVariable(b64ClusterCa, userData))
	if err != nil {
		return nil, err
	}

	clusterName = checkVariable(clusterName, userData)
	k8sMaster := checkVariable(apiServerEndpoint, userData)

	kubeconfigData := &clientcmdapi.Config{
		// Define a cluster stanza
		Clusters: map[string]*clientcmdapi.Cluster{
			clusterName: {
				Server:                   k8sMaster,
				CertificateAuthorityData: base64DecodedCAData,
			},
		},
		// Define auth based on the kubelet client cert retrieved
		AuthInfos: map[string]*clientcmdapi.AuthInfo{
			"kubelet": {
				Exec: &clientcmdapi.ExecConfig{
					APIVersion: "client.authentication.k8s.io/v1beta1",
					Command:    "aws",
					Args: []string{
						"eks",
						"get-token",
						"--cluster-name",
						clusterName,
						"--region",
						region,
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
				Cluster:  clusterName,
				AuthInfo: "kubelet",
			},
		},
		CurrentContext: "kubeletmein",
	}

	return kubeconfigData, nil
}

// Checks if a value looks like a shell variable and, if it is, finds the value in data
func checkVariable(value string, data string) string {
	re := regexp.MustCompile(`^\$([a-zA-Z0-9_-]*)`)
	variableName := re.FindStringSubmatch(value)
	if variableName != nil {
		re = regexp.MustCompile(`(?m)^` + regexp.QuoteMeta(variableName[1]) + `=(.*)`)
		fetchedValue := re.FindStringSubmatch(data)
		value = fetchedValue[1]
	}

	// Clean up any quotes
	unquotedValue := regexp.MustCompile(`^["'](.*)['"]$`).ReplaceAllString(value, `$1`)
	return unquotedValue
}
