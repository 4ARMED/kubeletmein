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
	"github.com/kubicorn/kubicorn/pkg/logger"
	"github.com/spf13/cobra"
)

// Config holds configuration values for GKE operations
type Config struct {
	bootstrapConfig string
	caCertPath      string
	kubeletKeyPath  string
	kubeletCertPath string
	kubeConfig      string
	certDir         string
	nodeName        string
}

// Command represents the gke command
func Command() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "gke",
		Short: "Retrieve kubelet creds in GKE",
		Run: func(cmd *cobra.Command, args []string) {
			if err := cmd.Help(); err != nil {
				logger.Debug("ignoring err %q", err.Error())
			}
		},
	}

	cmd.AddCommand(bootstrapCmd())
	cmd.AddCommand(generateCmd())

	return cmd

}
