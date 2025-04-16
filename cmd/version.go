package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

var (
	Revision      = "dev"
	VersionString = "dev"
)

func getVersion() string {
	return fmt.Sprintf(`Version: %s
Revision: %s
OS: %s
Arch: %s`, VersionString, Revision, runtime.GOOS, runtime.GOARCH)
}

var versionCmd = &cobra.Command{
	Use: "version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(getVersion())
	},
	Short: "Show version info",
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
