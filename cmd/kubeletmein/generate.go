package main

import (
	"fmt"
	"net/http"

	"github.com/4armed/kubeletmein/pkg/autodetect"
	"github.com/4armed/kubeletmein/pkg/config"
	"github.com/4armed/kubeletmein/pkg/do"
	"github.com/4armed/kubeletmein/pkg/eks"
	"github.com/4armed/kubeletmein/pkg/gke"
	"github.com/kubicorn/kubicorn/pkg/logger"
	"github.com/spf13/cobra"
)

// Generate runs the generate command
func Generate(c *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:              "generate",
		TraverseChildren: true,
		Short:            "Generate a kubeconfig",
		RunE: func(cmd *cobra.Command, args []string) error {

			if c.Provider == "autodetect" {
				logger.Info("running autodetect")
				hc := &http.Client{}
				ac, err := autodetect.New(hc, nil)
				if err != nil {
					return err
				}

				detectedProvider := ac.GetProvider()
				if detectedProvider == "" {
					return fmt.Errorf("could not autodetect provider. :-/")
				}
				c.Provider = detectedProvider
			}

			switch c.Provider {
			case "gke":
				logger.Info("GKE detected")
				err := gke.Generate(c)
				if err != nil {
					return err
				}
			case "do":
				logger.Info("DigitalOcean detected")
				err := do.Generate(c)
				if err != nil {
					return err
				}
			case "eks":
				logger.Info("EKS detected")
				err := eks.Generate(c)
				if err != nil {
					return err
				}
			case "autodetect":
				logger.Debug("autodetect? We should not have got here")
			default:
				return fmt.Errorf("Invalid provider: [%s]", c.Provider)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&c.Provider, "provider", "p", "autodetect", "The provider to generate for (eks, gke, do, autodetect)")
	cmd.Flags().StringVarP(&c.KubeConfig, "kubeconfig", "k", "kubeconfig.yaml", "The filename to write the kubeconfig to")
	cmd.Flags().StringVarP(&c.BootstrapConfig, "bootstrap-kubeconfig", "b", "bootstrap-kubeconfig.yaml", "The filename to write the bootstrap kubeconfig to")
	cmd.Flags().StringVarP(&c.MetadataFile, "metadata-file", "f", "", "Don't try to parse metadata, load from the specified filename instead")
	cmd.Flags().StringVarP(&c.NodeName, "node-name", "n", "", "Node name to use for CSR. Default is to detect from metadata if needed")
	cmd.Flags().StringVar(&c.Region, "region", "", "Region used in generation of kubeconfig")

	return cmd
}
