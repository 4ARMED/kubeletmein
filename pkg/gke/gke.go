package gke

import (
	"context"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"

	"cloud.google.com/go/compute/metadata"
	"github.com/4armed/kubeletmein/pkg/common"
	"github.com/4armed/kubeletmein/pkg/config"
	"github.com/4armed/kubeletmein/pkg/kubelet/certificate/bootstrap"
	"github.com/ghodss/yaml"
	"github.com/kubicorn/kubicorn/pkg/logger"
	"github.com/spf13/cobra"
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

// Command runs the gke command
func Command(c *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:              "gke",
		TraverseChildren: true,
		Short:            "Generate a kubeconfig on GKE",
		RunE: func(cmd *cobra.Command, args []string) error {

			if !c.SkipBootstrap {
				err := bootstrapKubeletConfig(c)
				if err != nil {
					return err
				}
			}

			logger.Info("using bootstrap-config to request new cert for node: %v", c.NodeName)
			err := bootstrap.LoadClientCert(context.TODO(), c.KubeConfig, c.BootstrapConfig, c.CertDir, types.NodeName(c.NodeName))
			if err != nil {
				return fmt.Errorf("unable to create certificate: %v", err)
			}

			logger.Info("got new cert and wrote kubeconfig")
			logger.Info("now try: kubectl --kubeconfig %v get pods", c.KubeConfig)

			return err
		},
	}

	cmd.Flags().BoolVarP(&c.SkipBootstrap, "skip-bootstrap", "s", false, "Skip bootstrap and use existing bootstrap-kubeconfig")
	cmd.Flags().StringVarP(&c.KubeletCertPath, "kubelet-cert", "c", "kubelet.crt", "The filename to write the kubelet cert to")
	cmd.Flags().StringVarP(&c.KubeletKeyPath, "kubelet-key", "k", "kubelet.key", "The filename to write the kubelet key to")
	cmd.Flags().StringVarP(&c.BootstrapConfig, "bootstrap-kubeconfig", "b", "bootstrap-kubeconfig", "The filename to write the bootstrap kubeconfig to")
	cmd.Flags().StringVarP(&c.CertDir, "cert-dir", "d", "pki", "Directory into which the new cert will be written")
	cmd.Flags().StringVarP(&c.NodeName, "node-name", "n", "", "Node name to use for CSR")
	cmd.MarkFlagRequired("node-name")

	return cmd
}

func fetchMetadataFromGKEService(metadataClient *metadata.Client) ([]byte, error) {
	ke, err := metadataClient.InstanceAttributeValue("kube-env")
	if err != nil {
		return nil, err
	}

	return []byte(ke), nil
}

func bootstrapKubeletConfig(c *config.Config) error {
	metadataClient := metadata.NewClient(&http.Client{})
	k := Kubeenv{}
	var kubeenv []byte
	var err error

	if c.MetadataFile == "" {
		logger.Info("fetching kubelet creds from metadata service")
		kubeenv, err = fetchMetadataFromGKEService(metadataClient)
		if err != nil {
			return err
		}
	} else {
		logger.Info("fetching kubelet creds from file: %v", c.MetadataFile)
		kubeenv, err = common.FetchMetadataFromFile(c.MetadataFile)
		if err != nil {
			return err
		}
	}

	logger.Debug("kubeenv: %v", kubeenv)

	err = yaml.Unmarshal(kubeenv, &k)
	if err != nil {
		return fmt.Errorf("unable to parse YAML from kube-env: %v", err)
	}

	logger.Debug("decoding ca cert")
	caCert, err := base64.StdEncoding.DecodeString(k.CaCert)
	if err != nil {
		return fmt.Errorf("unable to decode ca cert: %v", err)
	}
	logger.Info("writing ca cert to: %v", c.CaCertPath)
	err = ioutil.WriteFile(c.CaCertPath, caCert, 0644)
	if err != nil {
		return fmt.Errorf("unable to write ca cert to file: %v", err)
	}

	logger.Debug("decoding kubelet cert")
	kubeletCert, err := base64.StdEncoding.DecodeString(k.KubeletCert)
	if err != nil {
		return fmt.Errorf("unable to decode kubelet cert: %v", err)
	}

	logger.Info("writing kubelet cert to: %v", c.KubeletCertPath)
	err = ioutil.WriteFile(c.KubeletCertPath, kubeletCert, 0644)
	if err != nil {
		return fmt.Errorf("unable to write kubelet cert to file: %v", err)
	}

	logger.Debug("decoding kubelet key")
	kubeletKey, err := base64.StdEncoding.DecodeString(k.KubeletKey)
	if err != nil {
		return fmt.Errorf("unable to decode kubelet key: %v", err)
	}

	logger.Info("writing kubelet key to: %v", c.KubeletKeyPath)
	err = ioutil.WriteFile(c.KubeletKeyPath, kubeletKey, 0644)
	if err != nil {
		return fmt.Errorf("unable to write kubelet key to file: %v", err)
	}

	logger.Info("generating bootstrap-kubeconfig file at: %v", c.BootstrapConfig)
	kubeconfigData := clientcmdapi.Config{
		// Define a cluster stanza
		Clusters: map[string]*clientcmdapi.Cluster{"local": {
			Server:                "https://" + k.KubeMasterName,
			InsecureSkipTLSVerify: false,
			CertificateAuthority:  c.CaCertPath,
		}},
		// Define auth based on the kubelet client cert retrieved
		AuthInfos: map[string]*clientcmdapi.AuthInfo{"kubelet": {
			ClientCertificate: c.KubeletCertPath,
			ClientKey:         c.KubeletKeyPath,
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
