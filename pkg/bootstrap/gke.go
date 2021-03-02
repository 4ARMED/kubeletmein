// Copyright © 2018 Marc Wickenden <marc@4armed.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package bootstrap

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/http"

	"cloud.google.com/go/compute/metadata"
	"github.com/4armed/kubeletmein/pkg/config"
	"github.com/kubicorn/kubicorn/pkg/logger"
	"github.com/spf13/cobra"
	yaml "gopkg.in/yaml.v2"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

// Kubeenv stores the kube-env YAML
type Kubeenv struct {
	CaCert         string `yaml:"CA_CERT"`
	KubeletCert    string `yaml:"KUBELET_CERT"`
	KubeletKey     string `yaml:"KUBELET_KEY"`
	KubeMasterName string `yaml:"KUBERNETES_MASTER_NAME"`
}

// bootstrapCmd represents the bootstrap command
func bootstrapGkeCmd(c *config.Config) *cobra.Command {
	m := metadata.NewClient(&http.Client{})
	k := Kubeenv{}
	var kubeenv []byte
	var err error

	cmd := &cobra.Command{
		Use:              "gke",
		TraverseChildren: true,
		Short:            "Write out a bootstrap kubeconfig for the kubelet LoadClientCert function on GKE",
		RunE: func(cmd *cobra.Command, args []string) error {

			if c.MetadataFile == "" {
				logger.Info("fetching kubelet creds from metadata service")
				kubeenv, err = fetchMetadataFromService(m)
				if err != nil {
					return err
				}
			} else {
				logger.Info("fetching kubelet creds from file: %v", c.MetadataFile)
				kubeenv, err = fetchMetadataFromFile(c.MetadataFile)
				if err != nil {
					return err
				}
			}

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
			logger.Info("now generate a new node certificate with: kubeletmein generate")

			return err
		},
	}

	cmd.Flags().StringVarP(&c.KubeletCertPath, "kubelet-cert", "c", "kubelet.crt", "The filename to write the kubelet cert to")
	cmd.Flags().StringVarP(&c.KubeletKeyPath, "kubelet-key", "k", "kubelet.key", "The filename to write the kubelet key to")

	return cmd
}

func fetchMetadataFromService(m *metadata.Client) ([]byte, error) {
	ke, err := m.InstanceAttributeValue("kube-env")
	if err != nil {
		return nil, err
	}

	return []byte(ke), nil
}

func fetchMetadataFromFile(metadataFile string) ([]byte, error) {
	kubeenv, err := ioutil.ReadFile(metadataFile)
	if err != nil {
		return nil, err
	}

	return kubeenv, nil
}
