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

	"github.com/4armed/kubeletmein/pkg/common"
	"github.com/4armed/kubeletmein/pkg/config"
	"github.com/kubicorn/kubicorn/pkg/logger"
	"k8s.io/client-go/tools/clientcmd"
)

// Generate creates the kubeconfig for EKS
func Generate(c *config.Config) error {
	var userData string
	var region string
	var err error

	// get user-data
	if c.MetadataFile != "" {
		userDataBytes, err := common.FetchMetadataFromFile(c.MetadataFile)
		if err != nil {
			return err
		}
		userData = string(userDataBytes)
	} else {
		logger.Info("fetching cluster information from the metadata service")
		userData, err = getUserData()
		if err != nil {
			return err
		}
	}

	if c.Region == "" {
		region, err = getRegion()
		if err != nil {
			return err
		}
	}

	// These parsers should return an api.Config{} struct
	logger.Info("parsing user-data")
	kubeConfigData, err := ParseUserData(userData, region)
	if err != nil {
		return err
	}

	// Marshal to disk
	err = clientcmd.WriteToFile(*kubeConfigData, c.KubeConfig)
	if err != nil {
		return fmt.Errorf("unable to write kubeconfig file: %v", err)
	}

	logger.Info("wrote kubeconfig")
	logger.Info("now try: kubectl --kubeconfig %v get pods -A --field-selector spec.nodeName=%v", c.KubeConfig, c.NodeName)

	return err
}

// TODO: refactor all of this to be better mockable.
// At the moment we have these calls here which ignore the functions in
// metadata.go completely.
func getUserData() (string, error) {
	md, err := NewEC2MetadataClient()
	if err != nil {
		return "", err
	}

	return md.GetUserData()
}

func getRegion() (string, error) {
	md, err := NewEC2MetadataClient()
	if err != nil {
		return "", err
	}

	return md.Region()
}
