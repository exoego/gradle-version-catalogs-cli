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

func Test1(t *testing.T) {
	os.Args = []string{"cli", "generate"}
	assert.NoError(t, generateCommand.Execute())
}

func Test2(t *testing.T) {
	os.Args = []string{"cli", "generate", "./path/to/gradle/project"}
	assert.NoError(t, generateCommand.Execute())
}

func Test3(t *testing.T) {
	os.Args = []string{"cli", "generate", "./path/to/gradle/project", "extra-arg"}
	assert.Error(t, generateCommand.Execute())
}
