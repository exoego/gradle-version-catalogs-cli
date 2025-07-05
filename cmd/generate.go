package cmd

import (
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
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
		if len(args) > 2 {
			return errors.New("requires at most two arg")
		}
		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		gradleProjectRootPath, err := getWorkingDirectory(args)
		if err != nil {
			return err
		}

		useAutoLatest, err := cmd.Flags().GetBool("auto-latest")
		if err != nil {
			return fmt.Errorf("error option: %w", err)
		}

		gradleDirPath := filepath.Join(gradleProjectRootPath, "gradle")
		if _, err := os.Stat(gradleDirPath); os.IsNotExist(err) {
			return fmt.Errorf("not a Gradle project seemingly: %s", gradleDirPath)
		}

		foundFiles, err := findBuildGradle(gradleProjectRootPath, 3, 0)
		if err != nil {
			return fmt.Errorf("error during listing up build.gradle files: %w", err)
		}

		for _, file := range foundFiles {
			fmt.Printf("found build file: %s%s", file, LineBreak)
		}

		outputPath := filepath.Join(gradleProjectRootPath, "gradle", "libs.versions.toml")
		prevCatalog, err := ReadCatalog(outputPath)
		if err != nil {
			return fmt.Errorf("failed to read the existing libs.versions.toml: %w", err)
		}

		variableDefFiles := []string{filepath.Join(gradleProjectRootPath, "gradle.properties")}
		catalog, err := extractVersionCatalog(*prevCatalog, foundFiles, variableDefFiles)
		if err != nil {
			return fmt.Errorf("failed to extract libs.versions.toml: %w", err)
		}

		if useAutoLatest {
			searchLatestVersions(catalog)
		}

		embedResult, err := embedReferenceToLibs(foundFiles)
		if err != nil {
			return fmt.Errorf("failed to rewrite build files: %w", err)
		}

		if embedResult.UpdatedBuildSrc {
			var settingsGradlePath string
			if embedResult.WrittenInKotlin {
				settingsGradlePath = "settings.gradle.kts"
			} else {
				settingsGradlePath = "settings.gradle"
			}

			fullPath := filepath.Join(gradleProjectRootPath, "buildSrc", settingsGradlePath)
			err := writeBuildSrcSettings(fullPath)
			if err != nil {
				return fmt.Errorf("failed to write %s: %w", fullPath, err)
			}
			fmt.Printf("Updated: %s%s", fullPath, LineBreak)
			fmt.Printf("         ^ This file is used to resolve Version Catalog (libs.versions.toml) in buildSrc.%s", LineBreak)
		}

		err = WriteCatalog(outputPath, catalog)
		if err != nil {
			return fmt.Errorf("failed to write libs.versions.toml: %w", err)
		}

		fmt.Printf("Generated: %s%s", outputPath, LineBreak)

		return err
	},
}

func init() {
	rootCmd.AddCommand(generateCommand)
	generateCommand.Flags().Bool("auto-latest", true, "auto select latest version if none is specified")
}
