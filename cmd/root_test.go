package cmd

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestRoot(t *testing.T) {
	os.Args = []string{"cli"}
	assert.NoError(t, rootCmd.Execute())
}
