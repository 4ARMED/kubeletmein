package bootstrap

import "io/ioutil"

func fetchMetadataFromFile(metadataFile string) ([]byte, error) {
	metadata, err := ioutil.ReadFile(metadataFile)
	if err != nil {
		return nil, err
	}

	return metadata, nil
}
