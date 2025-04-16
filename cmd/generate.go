package cmd

import (
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var generateCommand = &cobra.Command{
	Use: "generate",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) > 1 {
			return errors.New("requires at most one arg")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		gradleProjectRootPath, err := getWorkingDirectory(args)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			return
		}

		catalogFile, err := openVersionCatalogFile(gradleProjectRootPath)
		defer func(catalogFile *os.File) {
			err := catalogFile.Close()
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				return
			}
		}(catalogFile)

		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			return
		}

		foundFiles, err := findBuildGradle(gradleProjectRootPath, 3, 0)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error during listing up build.gradle files: %v\n", err)
			return
		}

		if len(foundFiles) > 0 {
			fmt.Println("Found build.gradle files:")
			for _, file := range foundFiles {
				fmt.Println(file)
			}
		} else {
			fmt.Printf("No build.gradle files found within %s (up to depth %d).\n", gradleProjectRootPath, 3)
		}

		catalog, err := extractVersionCatalog(foundFiles)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error during exracting libs.versions.toml  : %v\n", err)
			return
		}

		err = WriteCatalog(catalogFile, catalog)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error during writing libs.versions.toml  : %v\n", err)
			return
		}

		fmt.Println("!!! DONE !!!")
		fmt.Printf("Generated: %s", catalogFile.Name())
	},
	Short: "Generate libs.versions.toml",
}

func init() {
	rootCmd.AddCommand(generateCommand)
}
