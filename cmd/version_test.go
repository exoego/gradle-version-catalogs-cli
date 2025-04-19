package cmd

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"runtime"
	"testing"
)

func CaptureStdout(t *testing.T, process func() error) (string, error) {
	t.Helper()
	original := os.Stdout
	defer func() {
		os.Stdout = original
	}()
	r, w, _ := os.Pipe()
	os.Stdout = w
	err := process()
	if err != nil {
		return "", err
	}
	err = w.Close()
	if err != nil {
		return "", err
	}
	var buffer bytes.Buffer
	if _, err := buffer.ReadFrom(r); err != nil {
		return "", err
	}
	s := buffer.String()
	return s[:len(s)-1], nil
}

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
