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

package gke

import (
	"fmt"

	"github.com/kubicorn/kubicorn/pkg/logger"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/kubernetes/pkg/kubelet/certificate/bootstrap"
)

// generateCmd represents the generate command
func generateCmd() *cobra.Command {
	config := &Config{}

	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate valid cert and kubeconfig from bootstrap config",
		RunE: func(cmd *cobra.Command, args []string) error {
			logger.Info("using bootstrap-config to request new cert for node: %v", config.nodeName)
			err := bootstrap.LoadClientCert(config.kubeConfig, config.bootstrapConfig, config.certDir, types.NodeName(config.nodeName))
			if err != nil {
				return fmt.Errorf("unable to create certificate: %v", err)
			}

			logger.Info("got new cert and wrote kubeconfig")
			logger.Info("now try: kubectl --kubeconfig %v get pods", config.kubeConfig)

			return err
		},
	}

	cmd.Flags().StringVarP(&config.bootstrapConfig, "bootstrap-kubeconfig", "b", "bootstrap-kubeconfig", "The filename to write the bootstrap kubeconfig to")
	cmd.Flags().StringVarP(&config.kubeConfig, "kubeconfig", "k", "kubeconfig", "The filename to write the kubeconfig to")
	cmd.Flags().StringVarP(&config.certDir, "cert-dir", "d", "pki", "Directory into which the new cert will be written")
	cmd.Flags().StringVarP(&config.nodeName, "node-name", "n", "", "Node name to use for CSR")
	cmd.MarkFlagRequired("node-name")

	return cmd
}
