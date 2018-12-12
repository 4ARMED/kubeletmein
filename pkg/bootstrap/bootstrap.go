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
	"github.com/4armed/kubeletmein/pkg/config"
	"github.com/spf13/cobra"
)

// Command represents the bootstrap command
func Command() *cobra.Command {

	c := &config.Config{}

	cmd := &cobra.Command{
		Use:   "bootstrap",
		Short: "Retrieve kubelet creds",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	cmd.AddCommand(bootstrapDoCmd(c))
	cmd.AddCommand(bootstrapGkeCmd(c))

	cmd.PersistentFlags().StringVarP(&c.BootstrapConfig, "bootstrap-kubeconfig", "b", "bootstrap-kubeconfig", "The filename to write the bootstrap kubeconfig to")
	cmd.PersistentFlags().StringVarP(&c.CaCertPath, "ca-cert", "a", "ca-certificates.crt", "The filename to write the apiserver CA cert to")
	cmd.PersistentFlags().StringVarP(&c.MetadataFile, "metadata-file", "f", "", "Don't try to parse metadata, load from the specified filename instead.")

	return cmd

}
