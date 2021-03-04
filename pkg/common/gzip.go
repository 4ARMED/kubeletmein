package common

import (
	"bytes"
	"compress/gzip"
)

// GunzipData ungzips the supplied data
func GunzipData(data []byte) ([]byte, error) {
	b := bytes.NewBuffer(data)

	r, err := gzip.NewReader(b)
	if err != nil {
		return nil, err
	}

	var result bytes.Buffer
	_, err = result.ReadFrom(r)
	if err != nil {
		return nil, err
	}

	return result.Bytes(), nil

}
