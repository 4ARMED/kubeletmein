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
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/4armed/kubeletmein/pkg/config"
	"github.com/kubicorn/kubicorn/pkg/logger"
	"github.com/spf13/cobra"
	yaml "gopkg.in/yaml.v2"
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

// bootstrapCmd represents the bootstrap command
func bootstrapDoCmd(c *config.Config) *cobra.Command {
	m := Metadata{}
	userData := []byte{}
	var kubeMaster string
	var err error

	cmd := &cobra.Command{
		Use:              "do",
		TraverseChildren: true,
		Short:            "Write out a bootstrap kubeconfig for the kubelet LoadClientCert function on Digital Ocean",
		RunE: func(cmd *cobra.Command, args []string) error {

			if c.MetadataFile == "" {
				logger.Info("fetching kubelet creds from metadata service")
				resp, err := http.Get("http://" + metadataIP + "/metadata/v1/user-data")
				if err != nil {
					panic(err)
				}
				defer resp.Body.Close()
				userData, err = ioutil.ReadAll(resp.Body)
				if err != nil {
					return err
				}
			} else {
				logger.Info("fetching kubelet creds from file: %v", c.MetadataFile)
				userData, err = ioutil.ReadFile(c.MetadataFile)
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
			logger.Info("now generate a new node certificate with: kubeletmein generate")

			return err
		},
	}

	return cmd
}
