package cmd

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestRoot(t *testing.T) {
	assert.NoError(t, rootCmd.Execute())
}

func Test1(t *testing.T) {
	os.Args = append(os.Args, "generate")
	os.Args = append(os.Args, "./path/to/gradle/project")
	assert.NoError(t, generateCommand.Execute())
}

func Test2(t *testing.T) {
	os.Args = append(os.Args, "generate")
	os.Args = append(os.Args, "./path/to/gradle/project")
	os.Args = append(os.Args, "extra-arg")
	assert.Error(t, generateCommand.Execute())
}
