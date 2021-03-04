package do

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/4armed/kubeletmein/pkg/common"
	"github.com/4armed/kubeletmein/pkg/config"
	"github.com/4armed/kubeletmein/pkg/kubelet/certificate/bootstrap"
	metadata "github.com/digitalocean/go-metadata"
	"github.com/ghodss/yaml"
	"github.com/kubicorn/kubicorn/pkg/logger"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

const (
	metadataIP = "169.254.169.254"
)

// Metadata stores the Kubernetes-related YAML
type Metadata struct {
	CaCert       string `yaml:"k8saas_ca_cert"`
	KubeletToken string `yaml:"k8saas_bootstrap_token"`
	KubeMaster   string `yaml:"k8saas_master_domain_name"`
}

// Command runs the digitalocean command
func Command(c *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:              "do",
		TraverseChildren: true,
		Short:            "Generate a kubeconfig on Digital Ocean",
		RunE: func(cmd *cobra.Command, args []string) error {

			if !c.SkipBootstrap {
				err := bootstrapKubeletConfig(c)
				if err != nil {
					return err
				}
			}

			logger.Info("using bootstrap-config to request new cert for node: %v", c.NodeName)
			logger.Debug("using bootstrap-config: %v and targeting kubeconfig file: %v", c.BootstrapConfig, c.KubeConfig)
			err := bootstrap.LoadClientCert(context.TODO(), c.KubeConfig, c.BootstrapConfig, c.CertDir, types.NodeName(c.NodeName))
			if err != nil {
				return fmt.Errorf("unable to create certificate: %v", err)
			}

			logger.Info("got new cert and wrote kubeconfig")
			logger.Info("now try: kubectl --kubeconfig %v get pods", c.KubeConfig)

			return err
		},
	}

	cmd.Flags().BoolVarP(&c.SkipBootstrap, "skip-bootstrap", "s", false, "Skip bootstrap and use existing kubeconfig")
	cmd.Flags().StringVarP(&c.KubeletCertPath, "kubelet-cert", "c", "kubelet.crt", "The filename to write the kubelet cert to")
	cmd.Flags().StringVarP(&c.KubeletKeyPath, "kubelet-key", "k", "kubelet.key", "The filename to write the kubelet key to")
	cmd.Flags().StringVarP(&c.BootstrapConfig, "bootstrap-kubeconfig", "b", "bootstrap-kubeconfig", "The filename to write the bootstrap kubeconfig to")
	cmd.Flags().StringVarP(&c.CertDir, "cert-dir", "d", "pki", "Directory into which the new cert will be written")
	cmd.Flags().StringVarP(&c.NodeName, "node-name", "n", "", "Node name to use for CSR")
	cmd.MarkFlagRequired("node-name")

	return cmd
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
	var kubeMaster string
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

	logger.Info("writing ca cert to: %v", c.CaCertPath)
	err = ioutil.WriteFile(c.CaCertPath, []byte(m.CaCert), 0644)
	if err != nil {
		return fmt.Errorf("unable to write ca cert to file: %v", err)
	}

	if os.Getenv("KUBERNETES_SERVICE_HOST") != "" && os.Getenv("KUBERNETES_SERVICE_PORT_HTTPS") != "" {
		kubeMaster = os.Getenv("KUBERNETES_SERVICE_HOST") + ":" + os.Getenv("KUBERNETES_SERVICE_PORT_HTTPS")
	} else {
		kubeMaster = m.KubeMaster
	}

	logger.Info("generating bootstrap-kubeconfig file at: %v", c.BootstrapConfig)
	kubeconfigData := clientcmdapi.Config{
		// Define a cluster stanza
		Clusters: map[string]*clientcmdapi.Cluster{"local": {
			Server:                "https://" + kubeMaster,
			InsecureSkipTLSVerify: false,
			CertificateAuthority:  c.CaCertPath,
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
