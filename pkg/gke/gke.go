package gke

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"

	"cloud.google.com/go/compute/metadata"
	"github.com/4armed/kubeletmein/pkg/common"
	"github.com/4armed/kubeletmein/pkg/config"
	"github.com/4armed/kubeletmein/pkg/kubelet/certificate/bootstrap"
	"github.com/ghodss/yaml"
	"github.com/kubicorn/kubicorn/pkg/logger"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

// Kubeenv stores the kube-env YAML
type Kubeenv struct {
	CaCert         string `yaml:"CA_CERT" json:"CA_CERT"`
	KubeletCert    string `yaml:"KUBELET_CERT" json:"KUBELET_CERT"`
	KubeletKey     string `yaml:"KUBELET_KEY" json:"KUBELET_KEY"`
	KubeMasterName string `yaml:"KUBERNETES_MASTER_NAME" json:"KUBERNETES_MASTER_NAME"`
}

// Generator provides a struct through which we call our funcs
type Generator struct {
	mc      *metadata.Client
	kubeEnv Kubeenv
	config  *config.Config
}

// Generate creates the kubeconfig for GKE
func Generate(c *config.Config) error {
	metadataClient := metadata.NewClient(&http.Client{})
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

		nodeName, err := generator.fetchHostNameFromGCEService()
		if err != nil {
			return err
		}

		c.NodeName = nodeName
	}

	logger.Info("using bootstrap-config to request new cert for node: %v", c.NodeName)
	err := bootstrap.LoadClientCert(context.TODO(), c.KubeConfig, c.BootstrapConfig, c.CertDir, types.NodeName(c.NodeName))
	if err != nil {
		return fmt.Errorf("unable to create certificate: %v", err)
	}

	logger.Info("got new cert and wrote kubeconfig")
	logger.Info("now try: kubectl --kubeconfig %v get pods", c.KubeConfig)

	return nil
}

func (g *Generator) fetchMetadataFromGKEService() ([]byte, error) {
	ke, err := g.mc.InstanceAttributeValue("kube-env")
	if err != nil {
		return nil, err
	}

	return []byte(ke), nil
}

func (g *Generator) bootstrapKubeletConfig() error {
	var kubeenv []byte
	var err error

	if g.config.MetadataFile == "" {
		logger.Info("fetching kubelet creds from metadata service")
		kubeenv, err = g.fetchMetadataFromGKEService()
		if err != nil {
			return err
		}
	} else {
		logger.Info("fetching kubelet creds from file: %v", g.config.MetadataFile)
		kubeenv, err = common.FetchMetadataFromFile(g.config.MetadataFile)
		if err != nil {
			return err
		}
	}

	logger.Debug("kubeenv: %v", string(kubeenv))

	err = yaml.Unmarshal(kubeenv, &g.kubeEnv)
	if err != nil {
		return fmt.Errorf("unable to parse YAML from kube-env: %v", err)
	}

	logger.Debug("decoding ca cert")
	caCert, err := base64.StdEncoding.DecodeString(g.kubeEnv.CaCert)
	if err != nil {
		return fmt.Errorf("unable to decode ca cert: %v", err)
	}

	logger.Debug("decoding kubelet cert")
	kubeletCert, err := base64.StdEncoding.DecodeString(g.kubeEnv.KubeletCert)
	if err != nil {
		return fmt.Errorf("unable to decode kubelet cert: %v", err)
	}

	logger.Debug("decoding kubelet key")
	kubeletKey, err := base64.StdEncoding.DecodeString(g.kubeEnv.KubeletKey)
	if err != nil {
		return fmt.Errorf("unable to decode kubelet key: %v", err)
	}

	logger.Info("generating bootstrap-kubeconfig file at: %v", g.config.BootstrapConfig)
	kubeconfigData := clientcmdapi.Config{
		// Define a cluster stanza
		Clusters: map[string]*clientcmdapi.Cluster{"local": {
			Server:                   "https://" + g.kubeEnv.KubeMasterName,
			InsecureSkipTLSVerify:    false,
			CertificateAuthorityData: caCert,
		}},
		// Define auth based on the kubelet client cert retrieved
		AuthInfos: map[string]*clientcmdapi.AuthInfo{"kubelet": {
			ClientCertificateData: kubeletCert,
			ClientKeyData:         kubeletKey,
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

func (g *Generator) fetchHostNameFromGCEService() (string, error) {
	hostname, err := g.mc.InstanceName()
	if err != nil {
		return "", err
	}

	return hostname, nil
}
