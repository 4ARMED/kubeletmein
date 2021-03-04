package common

import (
	"io/ioutil"

	"github.com/kubicorn/kubicorn/pkg/logger"
)

// FetchMetadataFromFile reads targets files and returns "metadata"
func FetchMetadataFromFile(metadataFile string) ([]byte, error) {

	logger.Debug("FetchMetadataFromFile: %s", metadataFile)
	metadata, err := ioutil.ReadFile(metadataFile)
	if err != nil {
		return nil, err
	}

	return metadata, nil
}
