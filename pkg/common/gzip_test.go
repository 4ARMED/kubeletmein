package common

import (
	"encoding/base64"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	base64GzipTestString = "H4sIAE4aQWAAAytJLS4BAAx+f9gEAAAA"
)

func TestIsGzippedWithGzip(t *testing.T) {
	gzipTestString, _ := base64.StdEncoding.DecodeString(base64GzipTestString)
	result, err := IsGzipped(gzipTestString)
	if err != nil {
		t.Errorf("err: %v", err)
	}

	assert.Equal(t, true, result, "should be equal")
}

func TestIsGzippedWithoutGzip(t *testing.T) {
	result, err := IsGzipped([]byte(base64GzipTestString))
	if err != nil {
		t.Errorf("err: %v", err)
	}

	assert.Equal(t, false, result, "should be equal")
}
