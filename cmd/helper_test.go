package cmd

import (
	"bytes"
	"github.com/stretchr/testify/assert"
	"os"
	"path/filepath"
	"testing"
)

func writeFile(t *testing.T, parent string, fileName string, content string) {
	filePath := filepath.Join(parent, fileName)

	enclosingDir := filepath.Dir(filePath)
	assert.NoError(t, os.MkdirAll(enclosingDir, 0750))

	err := os.WriteFile(filePath, []byte(content), 0750)
	assert.NoError(t, err)
}

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
