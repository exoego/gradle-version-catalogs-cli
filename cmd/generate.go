package cmd

import (
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"strings"
)

func openVersionCatalogFile(root os.File) (*os.File, error) {
	// Open the sub directory "gradle" under the root
	gradleDirPath := filepath.Join(root.Name(), "gradle")
	gradleDir, err := os.Open(gradleDirPath)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Not a Gradle project seemingly: %s", gradleDirPath))
	}
	defer gradleDir.Close()

	catalogFile, err := os.OpenFile(filepath.Join(gradleDir.Name(), "libs.version.toml"), os.O_RDWR|os.O_CREATE, 0644)
	if errors.Is(err, os.ErrNotExist) {
		// empty
		catalogFile.WriteString("")
	}
	defer catalogFile.Close()

	return catalogFile, nil
}

func findBuildGradle(root string, depth int, currentDepth int) ([]string, error) {
	var buildGradleFiles []string

	if currentDepth > depth {
		return buildGradleFiles, nil
	}

	entries, err := os.ReadDir(root)
	if err != nil {
		fmt.Printf("Error reading directory %s: %v\n", root, err)
		return buildGradleFiles, nil // Continue even if one directory fails
	}

	for _, entry := range entries {
		path := filepath.Join(root, entry.Name())
		if !entry.IsDir() && (strings.HasSuffix(entry.Name(), ".gradle") || strings.HasSuffix(entry.Name(), ".gradle.kts")) {
			buildGradleFiles = append(buildGradleFiles, path)
		} else if entry.IsDir() {
			subFiles, err := findBuildGradle(path, depth, currentDepth+1)
			if err != nil {
				return nil, err // Propagate the error if needed
			}
			buildGradleFiles = append(buildGradleFiles, subFiles...)
		}
	}

	return buildGradleFiles, nil
}

var generateCommand = &cobra.Command{
	Use: "generate",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) > 1 {
			return errors.New("requires at most one arg")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		var gradleProjectRootPath string

		if len(args) == 0 {
			// Print the current working directory
			dir, err := os.Getwd()
			if err != nil {
				fmt.Fprintf(os.Stderr, "%v\n", err)
				return
			}
			gradleProjectRootPath = dir
		} else {
			gradleProjectRootPath = args[0]
		}
		fmt.Sprintf("gradleProjectRootPath %s", gradleProjectRootPath)

		gradleRoot, err := os.Open(gradleProjectRootPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			return
		}

		catalogFile, err := openVersionCatalogFile(*gradleRoot)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			return
		}
		defer catalogFile.Close()

		// List up files where extension is .gradle and .gradle.kts
		foundFiles, err := findBuildGradle(gradleRoot.Name(), 3, 0)
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
			fmt.Printf("No build.gradle files found within %s (up to depth %d).\n", gradleRoot.Name(), 3)
		}
	},
	Short: "Generate libs.version.toml",
}

func init() {
	rootCmd.AddCommand(generateCommand)
}
