package do

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/4armed/kubeletmein/pkg/common"
	"github.com/4armed/kubeletmein/pkg/config"
	"github.com/4armed/kubeletmein/pkg/kubelet/certificate/bootstrap"
	metadata "github.com/digitalocean/go-metadata"
	"github.com/ghodss/yaml"
	"github.com/kubicorn/kubicorn/pkg/logger"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

const (
	metadataIP = "169.254.169.254"
)

// Metadata stores the Kubernetes-related YAML
type Metadata struct {
	CaCert       string `yaml:"k8saas_ca_cert" json:"k8saas_ca_cert"`
	KubeletToken string `yaml:"k8saas_bootstrap_token" json:"k8saas_bootstrap_token"`
	KubeMaster   string `yaml:"k8saas_master_domain_name" json:"k8saas_master_domain_name"`
}

// Generate creates the kubeconfig for DigitalOcean
func Generate(c *config.Config) error {
	if !c.SkipBootstrap {
		err := bootstrapKubeletConfig(c)
		if err != nil {
			return err
		}
	}

	// TODO: Tidy up duplicate of metadata client
	if c.NodeName == "" {
		logger.Debug("fetching nodename from metadata service")
		metadataClient := metadata.NewClient()
		nodeName, err := fetchHostNameFromDOService(metadataClient)
		if err != nil {
			return err
		}

		c.NodeName = nodeName
	}

	logger.Info("using bootstrap-config to request new cert for node: %v", c.NodeName)
	logger.Debug("using bootstrap-config: %v and targeting kubeconfig file: %v", c.BootstrapConfig, c.KubeConfig)
	err := bootstrap.LoadClientCert(context.TODO(), c.KubeConfig, c.BootstrapConfig, c.CertDir, types.NodeName(c.NodeName))
	if err != nil {
		return fmt.Errorf("unable to create certificate: %v", err)
	}

	logger.Info("got new cert and wrote kubeconfig")
	logger.Info("now try: kubectl --kubeconfig %v get pods", c.KubeConfig)

	return nil
}

func fetchMetadataFromDOService(metadataClient *metadata.Client) ([]byte, error) {
	logger.Info("fetching kubelet creds from metadata service")

	userData, err := metadataClient.UserData()
	if err != nil {
		return nil, err
	}

	return []byte(userData), nil
}

func bootstrapKubeletConfig(c *config.Config) error {
	metadataClient := metadata.NewClient()
	m := Metadata{}
	userData := []byte{}
	var err error

	if c.MetadataFile == "" {
		userData, err = fetchMetadataFromDOService(metadataClient)
		if err != nil {
			return err
		}
	} else {
		logger.Info("fetching kubelet creds from file: %v", c.MetadataFile)
		userData, err = common.FetchMetadataFromFile(c.MetadataFile)
		if err != nil {
			return err
		}
	}

	err = yaml.Unmarshal([]byte(userData), &m)
	if err != nil {
		return fmt.Errorf("unable to parse YAML from user-data: %v", err)
	}

	logger.Debug("encoding ca cert")
	var caCert []byte
	base64.StdEncoding.Encode(caCert, []byte(m.CaCert))

	logger.Info("generating bootstrap-kubeconfig file at: %v", c.BootstrapConfig)
	kubeconfigData := clientcmdapi.Config{
		// Define a cluster stanza
		Clusters: map[string]*clientcmdapi.Cluster{"local": {
			Server:                   "https://" + m.KubeMaster,
			InsecureSkipTLSVerify:    false,
			CertificateAuthorityData: caCert,
		}},
		// Define auth based on the kubelet client cert retrieved
		AuthInfos: map[string]*clientcmdapi.AuthInfo{"kubelet": {
			Token: m.KubeletToken,
		}},
		// Define a context and set as current
		Contexts: map[string]*clientcmdapi.Context{"service-account-context": {
			Cluster:  "local",
			AuthInfo: "kubelet",
		}},
		CurrentContext: "service-account-context",
	}

	// Marshal to disk
	err = clientcmd.WriteToFile(kubeconfigData, c.BootstrapConfig)
	if err != nil {
		return fmt.Errorf("unable to write bootstrap-kubeconfig file: %v", err)
	}

	logger.Info("wrote bootstrap-kubeconfig")

	return err
}

func fetchHostNameFromDOService(metadataClient *metadata.Client) (string, error) {
	hostname, err := metadataClient.Hostname()
	if err != nil {
		return "", err
	}

	return hostname, nil
}
