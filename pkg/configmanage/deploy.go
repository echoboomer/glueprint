package configmanage

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/kyokomi/emoji/v2"
	log "github.com/sirupsen/logrus"
)

func Deploy() {
	parsedFileContents, err := parseConfigurationFile()
	if err != nil {
		log.Error(err)
	}

	// Validate discovered components
	var validates bool
	for _, obj := range parsedFileContents {
		validates = validateManagedResources(obj, true)
	}

	// Iterate over discovered components
	if validates {
		for _, obj := range parsedFileContents {
			for k, v := range obj {
				credentials := Credentials{
					Hostname: v.Host,
					Username: "root",
					Password: v.Password,
				}

				_, err := emoji.Printf(":package: Applying configuration for %s\n", k)
				if err != nil {
					log.Fatal(err)
				}

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
					// Establish diffs
					// Packages
					packageDiffs := GetPackageDiffs(credentials, v.Packages, fromState[k])
					if len(packageDiffs) != 0 {
						for _, diff := range packageDiffs {
							switch diff.Operation {
							case "INSTALL":
								InstallPackage(credentials, diff.PackageResource)
							case "REMOVE":
								RemovePackage(credentials, diff.PackageResource)
							}
						}
					} else {
						log.Info("All packages are up to date\n")
					}
					// Files
					fileDiffs := GetFileDiffs(credentials, v.Files, fromState[k])
					if len(fileDiffs) != 0 {
						for _, diff := range fileDiffs {
							switch diff.Operation {
							case "CREATE":
								err := UploadFileViaSFTP(credentials, diff.FileResource)
								if err != nil {
									log.Errorf("Error copying file to host: %s", err)
								}
							case "REPLACE":
								DeleteFile(credentials, diff.FileResource)
								err := UploadFileViaSFTP(credentials, diff.FileResource)
								if err != nil {
									log.Errorf("Error copying file to host: %s", err)
								}
							case "UPDATE":
								UpdateFileMode(credentials, diff.FileResource)
							case "DELETE":
								DeleteFile(credentials, diff.FileResource)
							}
						}
					} else {
						log.Info("All files are up to date\n")
					}
					fmt.Println()
				} else {
					// Instantiate a new resource
					// Packages
					for _, pkg := range v.Packages {
						InstallPackage(credentials, pkg)
					}
					// Files
					for _, file := range v.Files {
						err := UploadFileViaSFTP(credentials, file)
						if err != nil {
							log.Errorf("Error copying file to host: %s", err)
						}
					}
				}
				if len(v.Command) != 0 {
					// Run any commands
					command := strings.Join(v.Command, " ")
					result, err := RunOnRemoteHost(credentials, command)
					if err != nil {
						log.Errorf("Error executing command: %s", err)
					}
					log.Infof("Running command on host...")
					fmt.Println(result)
				}
			}
			// Write to state
			WriteToState(obj)
			color.Green("Deploy complete!")
		}
	}
}
