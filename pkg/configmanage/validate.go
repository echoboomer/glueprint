package configmanage

import (
	"fmt"

	"github.com/kyokomi/emoji/v2"
	log "github.com/sirupsen/logrus"
)

// Validate parses fields in a configuration file and returns
// whether or not the file structure is valid
func Validate() {
	// Parse contents of the configuration file
	parsedFileContents, err := parseConfigurationFile()
	if err != nil {
		log.Error(err)
	}

	// Validate discovered components
	for _, obj := range parsedFileContents {
		validateManagedResources(obj, false)
	}
}

// validateManagedResources parses resources specified in the
// configuration file and validates them
func validateManagedResources(resource map[string]ManagedResource, silent bool) bool {
	var validates bool
	for k, v := range resource {
		if !silent {
			_, err := emoji.Printf(":package: %s\n", k)
			if err != nil {
				log.Fatal(err)
			}
		}
		if len(v.Files) >= 1 && len(v.Packages) >= 1 {
			_, err := emoji.Printf(":white_check_mark: %s %s\n", k, "Passes Validation")
			if err != nil {
				log.Fatal(err)
			}
			validates = true
		} else {
			_, err := emoji.Printf(":x: %s %s\n", k, "Fails Validation")
			if err != nil {
				log.Fatal(err)
			}
			validates = false
		}

		if !silent {
			fmt.Println("----------------------------------------")
			_, err := emoji.Printf(":file_folder: %s\n", "Files")
			if err != nil {
				log.Fatal(err)
			}
			if len(v.Files) >= 1 {
				for _, v := range v.Files {
					fmt.Printf("	Filename: %s\n", v.Name)
					fmt.Printf("	Path: %s\n", v.Path)
					fmt.Printf("	Mode: %v\n", v.Mode)
				}
			} else {
				fmt.Printf("This resource has not declared any files. You may add them with the files section.\n\n")
			}

			_, err = emoji.Printf(":wrench: %s\n", "Packages")
			if err != nil {
				log.Fatal(err)
			}
			if len(v.Packages) >= 1 {
				for _, v := range v.Packages {
					fmt.Printf("	Package Name: %s\n", v.Package)
					fmt.Printf("	Version: %s\n\n", v.Version)
				}
			} else {
				fmt.Printf("This resource has not declared any packages. You may add them with the packages section.\n\n")
			}

			fmt.Printf("Command: %s\n", v.Command)

			fmt.Printf("----------\n\n")
		}
	}
	return validates
}
