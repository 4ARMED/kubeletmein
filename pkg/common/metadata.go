package common

import "io/ioutil"

// FetchMetadataFromFile eads targets files and returns "metadata"
func FetchMetadataFromFile(metadataFile string) ([]byte, error) {
	metadata, err := ioutil.ReadFile(metadataFile)
	if err != nil {
		return nil, err
	}

	return metadata, nil
}
