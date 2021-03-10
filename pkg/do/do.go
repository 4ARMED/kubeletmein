package do

import (
	"context"
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

// Generator provides a struct through which we call our funcs
type Generator struct {
	mc       *metadata.Client
	metadata Metadata
	config   *config.Config
}

// Generate creates the kubeconfig for DigitalOcean
func Generate(c *config.Config) error {
	metadataClient := metadata.NewClient()
	generator := &Generator{
		mc:     metadataClient,
		config: c,
	}

	if !c.SkipBootstrap {
		err := generator.bootstrapKubeletConfig()
		if err != nil {
			return err
		}
	}

	if c.NodeName == "" {
		logger.Debug("fetching nodename from metadata service")
		nodeName, err := generator.fetchHostNameFromDOService()
		if err != nil {
			return err
		}

		c.NodeName = nodeName
	}

	logger.Info("using %s to request new cert for node: %v", c.BootstrapConfig, c.NodeName)
	logger.Debug("using bootstrap-config: %v and targeting kubeconfig file: %v", c.BootstrapConfig, c.KubeConfig)
	err := bootstrap.LoadClientCert(context.TODO(), c.KubeConfig, c.BootstrapConfig, c.CertDir, types.NodeName(c.NodeName))
	if err != nil {
		return fmt.Errorf("unable to create certificate: %v", err)
	}

	logger.Info("got new cert and wrote kubeconfig")
	logger.Info("now try: kubectl --kubeconfig %v get pods", c.KubeConfig)

	return nil
}

func (g *Generator) fetchMetadataFromDOService() ([]byte, error) {
	logger.Info("fetching kubelet creds from metadata service")

	userData, err := g.mc.UserData()
	if err != nil {
		return nil, err
	}

	return []byte(userData), nil
}

func (g *Generator) bootstrapKubeletConfig() error {
	userData := []byte{}
	var err error

	if g.config.MetadataFile == "" {
		userData, err = g.fetchMetadataFromDOService()
		if err != nil {
			return err
		}
	} else {
		logger.Info("fetching kubelet creds from file: %v", g.config.MetadataFile)
		userData, err = common.FetchMetadataFromFile(g.config.MetadataFile)
		if err != nil {
			return err
		}
	}

	err = yaml.Unmarshal([]byte(userData), &g.metadata)
	if err != nil {
		return fmt.Errorf("unable to parse YAML from user-data: %v", err)
	}

	logger.Info("generating bootstrap-kubeconfig file at: %v", g.config.BootstrapConfig)
	kubeconfigData := clientcmdapi.Config{
		// Define a cluster stanza
		Clusters: map[string]*clientcmdapi.Cluster{"local": {
			Server:                   "https://" + g.metadata.KubeMaster,
			InsecureSkipTLSVerify:    false,
			CertificateAuthorityData: []byte(g.metadata.CaCert),
		}},
		// Define auth based on the kubelet client cert retrieved
		AuthInfos: map[string]*clientcmdapi.AuthInfo{"kubelet": {
			Token: g.metadata.KubeletToken,
		}},
		// Define a context and set as current
		Contexts: map[string]*clientcmdapi.Context{"service-account-context": {
			Cluster:  "local",
			AuthInfo: "kubelet",
		}},
		CurrentContext: "service-account-context",
	}

	// Marshal to disk
	err = clientcmd.WriteToFile(kubeconfigData, g.config.BootstrapConfig)
	if err != nil {
		return fmt.Errorf("unable to write bootstrap-kubeconfig file: %v", err)
	}

	logger.Info("wrote bootstrap-kubeconfig")

	return err
}

func (g *Generator) fetchHostNameFromDOService() (string, error) {
	hostname, err := g.mc.Hostname()
	if err != nil {
		return "", err
	}

	return hostname, nil
}
