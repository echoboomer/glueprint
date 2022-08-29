package configmanage

import (
	"io/fs"
	"os"

	"github.com/kyokomi/emoji/v2"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

// The file that the application looks for to establish whether or
// not there are resources to manage
var watchedFileName string = "glue.yaml"

// listDirectoryContents returns objects in the directory in which the
// utility is called from
func listDirectoryContents() ([]fs.DirEntry, error) {
	// Context is always current directory
	path, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	// Get all files in current directory
	files, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	return files, nil
}

// parseConfigurationFile returns the contents of discovered configuration
// files
func parseConfigurationFile() ([]map[string]ManagedResource, error) {
	var parsedFileContents = []map[string]ManagedResource{}

	currentDirectoryFiles, err := listDirectoryContents()
	if err != nil {
		log.Errorf("Error listing directory contents: %s", err)
		return nil, err
	}

	results := traverseFiles(currentDirectoryFiles)
	if len(results) < 1 {
		log.Warnf("No files matching %s were found in current directory.\n"+
			"To use glueprint, create a file called %s and specify a configuration.\n"+
			"More details are available in the docs.", watchedFileName, watchedFileName)
	}
	for _, f := range results {
		out := make(map[string]ManagedResource)
		data, err := os.ReadFile(f.Name())
		if err != nil {
			log.Errorf("Error reading configuration file: %s", err)
			return nil, err
		}
		err = yaml.Unmarshal([]byte(string(data)), &out)
		if err != nil {
			log.Errorf("Error parsing configuration file: %s", err)
		}
		parsedFileContents = append(parsedFileContents, out)
	}

	return parsedFileContents, nil
}

// traverseFiles loops over objects passed in to determine if a target
// set of files exists
func traverseFiles(files []fs.DirEntry) []fs.DirEntry {
	foundFiles := []fs.DirEntry{}
	for _, file := range files {
		if file.IsDir() {
			// ToDo: handle recursive file lookups
			continue
		}
		if file.Name() == watchedFileName {
			_, err := emoji.Printf(":white_check_mark: Found configuration file %s\n\n", file.Name())
			if err != nil {
				log.Fatal(err)
			}
			foundFiles = append(foundFiles, file)
		}
	}
	return foundFiles
}
