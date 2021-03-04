// Copyright © 2021 Amiran Alavidze @airman604 and Marc Wickenden @marcwickenden
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

package eks

import (
	"fmt"
	"regexp"

	"github.com/4armed/kubeletmein/pkg/common"
	"github.com/4armed/kubeletmein/pkg/config"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/kubicorn/kubicorn/pkg/logger"
	"github.com/spf13/cobra"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

var (
	// KubeConfigData will hold the kubeconfig data we will marshal to a file
	KubeConfigData = clientcmdapi.Config{}
)

// Command runs the eks command
func Command(c *config.Config) *cobra.Command {

	cmd := &cobra.Command{
		Use:   "eks",
		Short: "Generate a kubeconfig on EKS.",
		RunE: func(cmd *cobra.Command, args []string) error {

			logger.Info("generating kubeconfig for current EKS node")
			err := generateKubeConfig(c)
			if err != nil {
				return fmt.Errorf("unable to generate kubeconfig: %v", err)
			}

			logger.Info("now try: kubectl --kubeconfig %v get pods", c.KubeConfig)

			return err
		},
	}

	cmd.Flags().StringVarP(&c.KubeConfig, "kubeconfig", "k", "kubeconfig", "The filename to write the kubeconfig to")

	return cmd
}

// func getUserData() (string, error) {
// 	md, err := NewEC2MetadataClient()
// 	if err != nil {
// 		return "", err
// 	}

// 	userData, err := md.GetUserData()
// 	if err != nil {
// 		return "", err
// 	}

// 	return userData, nil
// }

func getUserData() (string, error) {
	md := ec2metadata.New(session.New())

	userData, err := md.GetUserData()
	if err != nil {
		return "", err
	}

	return userData, nil
}

func generateKubeConfig(c *config.Config) error {
	var userData string
	var err error

	// get user-data
	if c.MetadataFile != "" {
		userDataBytes, err := common.FetchMetadataFromFile(c.MetadataFile)
		userData = string(userDataBytes)
		if err != nil {
			return err
		}
	} else {
		logger.Info("fetching cluster information from user-data from the metadata service")
		userData, err = getUserData()
		if err != nil {
			return err
		}
	}

	// These parsers should return an api.Config{} struct
	logger.Info("determining type of user-data")
	re := regexp.MustCompile(`Content-Type: text/x-shellscript`)
	match := re.MatchString(userData)
	if match {
		logger.Info("text/x-shellscript detected")
		KubeConfigData, err = ParseShellScript(userData)
		if err != nil {
			return err
		}
	} else {
		logger.Info("assuming gzipped cloud-config")
		KubeConfigData, err = ParseCloudConfig([]byte(userData))
		if err != nil {
			return err
		}
	}

	// Marshal to disk
	err = clientcmd.WriteToFile(KubeConfigData, c.KubeConfig)
	if err != nil {
		return fmt.Errorf("unable to write kubeconfig file: %v", err)
	}

	logger.Info("wrote kubeconfig")

	return err
}
