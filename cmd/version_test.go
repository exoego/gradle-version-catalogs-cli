package cmd

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"runtime"
	"testing"
)

func TestGetVersion(t *testing.T) {
	os.Args = []string{"cli", "version"}
	stdout, err := CaptureStdout(t, generateCommand.Execute)
	assert.NoError(t, err)

	want := fmt.Sprintf(`Version: %s
Revision: %s
OS: %s
Arch: %s`, VersionString, Revision, runtime.GOOS, runtime.GOARCH)

	assert.Equal(t, want, stdout)
}
