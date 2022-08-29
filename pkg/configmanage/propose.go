package configmanage

import (
	"fmt"
	"time"

	"github.com/fatih/color"
	"github.com/kyokomi/emoji/v2"
	log "github.com/sirupsen/logrus"
)

// Propose determines what changes need to be made and clearly describes
// them
func Propose() {
	// Parse contents of the configuration file
	parsedFileContents, err := parseConfigurationFile()
	if err != nil {
		log.Error(err)
	}

	// Validate discovered components
	var validates bool
	for _, obj := range parsedFileContents {
		validates = validateManagedResources(obj, false)
		fmt.Println()
	}

	// Iterate over discovered components
	if validates {
		var fileDiffs []FileResourceDiff
		var packageDiffs []PackageResourceDiff
		for _, obj := range parsedFileContents {
			for k, v := range obj {
				credentials := Credentials{
					Hostname: v.Host,
					Username: "root",
					Password: v.Password,
				}
				_, err := emoji.Printf(":package: Proposed configuration for %s\n", k)
				if err != nil {
					log.Fatal(err)
				}
				// Proposed Changes
				showProposedOutput(v)
				// If the resource exists in state, we must compare it
				// Otherwise, it doesn't exist and should be created
				fromState := ReadOneFromState(k)
				var resourceExistsInState bool
				if fromState[k].Host == v.Host {
					resourceExistsInState = true
				} else {
					resourceExistsInState = false
				}
				if resourceExistsInState {
					// Files
					fileDiffs = GetFileDiffs(credentials, v.Files, fromState[k])
					fmt.Println()
					// Packages
					packageDiffs = GetPackageDiffs(credentials, v.Packages, fromState[k])
				} else {
					// Files
					fileDiffs = GetFileDiffs(credentials, v.Files, ManagedResource{})
					fmt.Println()
					// Packages
					packageDiffs = GetPackageDiffs(credentials, v.Packages, ManagedResource{})
				}
				fmt.Println("----------------------------------------")
				fmt.Println()
				time.Sleep(2 * time.Second)
			}
		}
		if len(fileDiffs) == 0 && len(packageDiffs) == 0 {
			color.Green("No changes to apply, resource is up to date")
		} else {
			color.Green("To apply these changes, run: glueprint deploy")
		}
	}
}

// showProposedOutput displays proposed changes
func showProposedOutput(resource ManagedResource) {
	fmt.Println("----------")
	_, err := emoji.Printf(":file_folder: %s\n", "Files")
	if err != nil {
		log.Fatal(err)
	}
	if len(resource.Files) >= 1 {
		for _, v := range resource.Files {
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
	if len(resource.Packages) >= 1 {
		for _, v := range resource.Packages {
			fmt.Printf("	Package Name: %s\n", v.Package)
			fmt.Printf("	Version: %s\n\n", v.Version)
		}
	} else {
		fmt.Printf("This resource has not declared any packages. You may add them with the packages section.\n\n")
	}
	fmt.Printf("Command: %s\n", resource.Command)
	fmt.Printf("----------\n\n")
}
