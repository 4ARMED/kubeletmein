package common

import (
	"bufio"
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

// GzipData compresses data
func GzipData(data []byte) ([]byte, error) {
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)

	_, err := gz.Write(data)
	if err != nil {
		return nil, err
	}

	if err = gz.Flush(); err != nil {
		return nil, err
	}

	if err = gz.Close(); err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

// IsGzipped tests whether data is gzipped
func IsGzipped(data []byte) (bool, error) {
	r := bytes.NewReader(data)
	bReader := bufio.NewReader(r)
	testBytes, err := bReader.Peek(2) // Read first 2 bytes
	if err != nil {
		return false, err
	}

	// Check if first two bytes are 0x1f8b
	if testBytes[0] == 31 && testBytes[1] == 139 {
		// This is gzip
		return true, nil
	}

	return false, nil
}
