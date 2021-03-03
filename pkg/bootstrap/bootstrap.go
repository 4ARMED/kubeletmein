package bootstrap

import (
	"github.com/4armed/kubeletmein/pkg/config"
	"github.com/4armed/kubeletmein/pkg/do"
	"github.com/4armed/kubeletmein/pkg/gke"
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

	cmd.AddCommand(do.BootstrapCmd(c))
	cmd.AddCommand(gke.BootstrapCmd(c))

	cmd.PersistentFlags().StringVarP(&c.BootstrapConfig, "bootstrap-kubeconfig", "b", "bootstrap-kubeconfig", "The filename to write the bootstrap kubeconfig to")
	cmd.PersistentFlags().StringVarP(&c.CaCertPath, "ca-cert", "a", "ca-certificates.crt", "The filename to write the apiserver CA cert to")
	cmd.PersistentFlags().StringVarP(&c.MetadataFile, "metadata-file", "f", "", "Don't try to parse metadata, load from the specified filename instead.")

	return cmd

}
