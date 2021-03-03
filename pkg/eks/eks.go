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
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/4armed/kubeletmein/pkg/config"
	"github.com/kubicorn/kubicorn/pkg/logger"
	"github.com/spf13/cobra"
)

const (
	metadataIP = "169.254.169.254"
)

// Command runs the eks command
func Command(c *config.Config) *cobra.Command {

	cmd := &cobra.Command{
		Use:   "eks",
		Short: "Generate a kubeconfig on EKS.",
		RunE: func(cmd *cobra.Command, args []string) error {
			logger.Info("generating kubeconfig for current EKS node")
			err := doCommand(c)
			if err != nil {
				return fmt.Errorf("unable to generate kubeconfig: %v", err)
			}

			logger.Info("wrote kubeconfig")
			logger.Info("to use the kubeconfig, download aws-iam-authenticator to the current directory and make it executable by following the instructions at https://docs.aws.amazon.com/eks/latest/userguide/install-aws-iam-authenticator.html")
			logger.Info("then try: kubectl --kubeconfig %v get pods", c.KubeConfig)

			return err
		},
	}

	cmd.Flags().StringVarP(&c.KubeConfig, "kubeconfig", "k", "kubeconfig", "The filename to write the kubeconfig to")

	return cmd
}

func getUserData() (string, error) {
	// using AWS v2 metadata API (IMDSv2), get token first
	logger.Info("getting IMDSv2 token")
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodPut, "http://"+metadataIP+"/latest/api/token", nil)
	if err != nil {
		return "", err
	}
	// set token TTL to be 10 min
	req.Header.Set("X-aws-ec2-metadata-token-ttl-seconds", "600")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	tokenBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	metadataToken := string(tokenBytes)

	// now request the instance provisioning data
	logger.Info("getting user-data")
	req, err = http.NewRequest(http.MethodGet, "http://"+metadataIP+"/latest/user-data", nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("X-aws-ec2-metadata-token", metadataToken)

	resp, err = client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	userDataBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(userDataBytes), nil
}

type eksKubeConfigInfo struct {
	caData      string
	kubeMaster  string
	clusterName string
}

// parse user-data from the metadata service
func parseUserData(userData string) (*eksKubeConfigInfo, error) {
	// userData should contain the following lines:
	// B64_CLUSTER_CA=...
	// API_SERVER_URL=...
	// /etc/eks/bootstrap.sh <CLUSTER_NAME> ...
	re := regexp.MustCompile(`(?m)^B64_CLUSTER_CA=(.*)$`)
	caData := re.FindStringSubmatch(userData)
	if caData == nil {
		return nil, errors.New("Error while parsing user-data, could not find B64_CLUSTER_CA")
	}

	re = regexp.MustCompile(`(?m)^API_SERVER_URL=(.*)$`)
	k8sMaster := re.FindStringSubmatch(userData)
	if k8sMaster == nil {
		return nil, errors.New("Error while parsing user-data, could not find API_SERVER_URL")
	}

	re = regexp.MustCompile(`(?m)^/etc/eks/bootstrap.sh\s+(\S+)\s`)
	clusterName := re.FindStringSubmatch(userData)
	if clusterName == nil {
		return nil, errors.New("Error while parsing user-data, could not find cluster name from bootstrap.sh parameters")
	}

	result := &eksKubeConfigInfo{
		caData:      caData[1],
		kubeMaster:  k8sMaster[1],
		clusterName: clusterName[1],
	}
	return result, nil
}

func kubeConfigTemplate() string {
	// template from https://github.com/awslabs/amazon-eks-ami/blob/master/files/kubelet-kubeconfig
	// same information here: https://docs.aws.amazon.com/eks/latest/userguide/create-kubeconfig.html
	kubeconfig := "" +
		"apiVersion: v1\n" +
		"kind: Config\n" +
		"clusters:\n" +
		"- cluster:\n" +
		"    certificate-authority-data: B64_CA_DATA\n" +
		"    server: MASTER_ENDPOINT\n" +
		"  name: kubernetes\n" +
		"contexts:\n" +
		"- context:\n" +
		"    cluster: kubernetes\n" +
		"    user: kubelet\n" +
		"  name: kubelet\n" +
		"current-context: kubelet\n" +
		"users:\n" +
		"- name: kubelet\n" +
		"  user:\n" +
		"    exec:\n" +
		"      apiVersion: client.authentication.k8s.io/v1alpha1\n" +
		"      command: AWS_IAM_AUTHENTICATOR\n" +
		"      args:\n" +
		"      - \"token\"\n" +
		"      - \"-i\"\n" +
		"      - \"CLUSTER_NAME\"\n" //+
		// "      - --region\n" +
		// "      - \"AWS_REGION\"\n" +

	return kubeconfig
}

func doCommand(c *config.Config) error {

	// get user-data from the metadata service
	logger.Info("fetching cluster information from user-data from the metadata service")
	userData, err := getUserData()
	if err != nil {
		return err
	}

	kubeInfo, err := parseUserData(userData)
	if err != nil {
		return err
	}

	// construct file path for aws-iam-authenticator - assume it will be downloaded to the current dir
	dir, err := os.Getwd()
	if err != nil {
		return err
	}
	authenticatorPath := filepath.Join(dir, "aws-iam-authenticator")

	// template from https://github.com/awslabs/amazon-eks-ami/blob/master/files/kubelet-kubeconfig
	// same information here: https://docs.aws.amazon.com/eks/latest/userguide/create-kubeconfig.html
	kubeconfig := kubeConfigTemplate()

	kubeconfig = strings.ReplaceAll(kubeconfig, "B64_CA_DATA", kubeInfo.caData)
	kubeconfig = strings.ReplaceAll(kubeconfig, "MASTER_ENDPOINT", kubeInfo.kubeMaster)
	kubeconfig = strings.ReplaceAll(kubeconfig, "CLUSTER_NAME", kubeInfo.clusterName)
	kubeconfig = strings.ReplaceAll(kubeconfig, "AWS_IAM_AUTHENTICATOR", authenticatorPath)

	logger.Info("generating EKS node kubeconfig file at: %v", c.KubeConfig)
	err = ioutil.WriteFile(c.KubeConfig, []byte(kubeconfig), 0644)
	if err != nil {
		return fmt.Errorf("error while writing kubeconfig file: %v", err)
	}

	return err
}
