package common

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMetadataFromFile(t *testing.T) {
	testMetadataFile := filepath.Join("testdata", "userdata-shell.txt")
	metadata, err := FetchMetadataFromFile(testMetadataFile)
	if err != nil {
		t.Errorf("err: %v", err)
	}

	assert.Contains(t, string(metadata), "Content-Type: text/x-shellscript", "should contain")
}
