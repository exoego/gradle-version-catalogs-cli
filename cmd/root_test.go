package cmd

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestRootViaExecute(t *testing.T) {
	os.Args = []string{"cli"}
	stdout, err := CaptureStdout(t, func() error {
		Execute()
		return nil
	})
	assert.NoError(t, err)
	assert.Contains(t, stdout, "Usage:")
}
