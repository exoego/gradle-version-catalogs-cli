package cmd

import (
	"errors"
	"fmt"
	"github.com/spf13/cobra"
)

var generateCommand = &cobra.Command{
	Use:   "generate [PATH]",
	Short: "Generate libs.versions.toml",
	Long: `
Collects library versions in multiple build.gradle(.kts) and generates libs.versions.toml in PATH/gradle.
If no PATH is provided, the current working directory is used.
Some manual intervention may be required.

Caution:
  If libs.version.toml already exists, it will be overwritten.
`,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) > 1 {
			return errors.New("requires at most one arg")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		gradleProjectRootPath, err := getWorkingDirectory(args)
		if err != nil {
			return err
		}

		catalogFile, err := openVersionCatalogFile(gradleProjectRootPath)
		if err != nil {
			return err
		}

		foundFiles, err := findBuildGradle(gradleProjectRootPath, 3, 0)
		if err != nil {
			return fmt.Errorf("error during listing up build.gradle files: %w", err)
		}

		for _, file := range foundFiles {
			fmt.Printf("found build file: %s", file)
		}

		catalog, err := extractVersionCatalog(foundFiles)
		if err != nil {
			return fmt.Errorf("failed to extract libs.versions.toml: %w", err)
		}

		err = embedReferenceToLibs(foundFiles, catalog)
		if err != nil {
			return fmt.Errorf("failed to rewrite build files: %w", err)
		}

		err = WriteCatalog(catalogFile, catalog)
		if err != nil {
			return fmt.Errorf("failed to write libs.versions.toml: %w", err)
		}

		fmt.Println("!!! DONE !!!")
		fmt.Println(fmt.Sprintf("Generated: %s", catalogFile.Name()))

		err = catalogFile.Close()
		return err
	},
}

func init() {
	rootCmd.AddCommand(generateCommand)

}
