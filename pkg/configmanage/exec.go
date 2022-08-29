package configmanage

import (
	"bytes"
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/fatih/color"
	"github.com/kyokomi/emoji/v2"
	"github.com/r3labs/diff/v3"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

var sshPort string = "22"

type Credentials struct {
	Hostname string
	Username string
	Password string
}

// RunOnRemoteHost allows execution of a command on a host via an ssh
// shell - useful for managing resources on remote hosts
func RunOnRemoteHost(credentials Credentials, command string) (string, error) {
	config := &ssh.ClientConfig{
		User: credentials.Username,
		Auth: []ssh.AuthMethod{
			ssh.Password(credentials.Password),
		},
		// This is never acceptable in production, but is suitable for a demo
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	hostAddress := strings.Join([]string{credentials.Hostname, sshPort}, ":")
	client, err := ssh.Dial("tcp", hostAddress, config)
	if err != nil {
		log.Errorf("Error executing command on host: %s", err)
		return "", err
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		log.Errorf("Error executing command on host: %s", err)
		return "", err
	}
	defer session.Close()

	var outb, errb bytes.Buffer
	session.Stdout = &outb
	session.Stderr = &errb
	fmt.Println(errb.String())
	if err := session.Run(command); err != nil {
		log.Errorf("Error executing command on host: %s", err)
		return "", err
	}

	return outb.String(), nil
}

// GetFileDiffs processes a slice of FileSpecification and determines changes for
// resources that already have state entries
func GetFileDiffs(credentials Credentials, files []FileSpecification, fromState ManagedResource) []FileResourceDiff {
	_, err := emoji.Printf(":file_folder: %s\n", "Files")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("-----------------------------------")
	var diffs []FileResourceDiff
	changelog, err := diff.Diff(fromState.Files, files, diff.DisableStructValues())
	if err != nil {
		log.Errorf("Error comparing filesets: %s", err)
	}
	if len(changelog) != 0 {
		for _, ch := range changelog {
			switch ch.Type {
			case "delete":
				idx, _ := strconv.Atoi(ch.Path[0])
				fileToDelete := fromState.Files[idx]
				color.Yellow("File %s will be deleted", strings.Join([]string{fileToDelete.Path, fileToDelete.Name}, "/"))
				diffs = append(diffs, FileResourceDiff{Operation: strings.ToUpper(ch.Type), Target: "", FileResource: fileToDelete})
			case "update":
				idx, _ := strconv.Atoi(ch.Path[0])
				field := ch.Path[1]
				fileToUpdate := fromState.Files[idx]
				switch field {
				case "Mode":
					color.Yellow("File %s will be updated in place:", strings.Join([]string{fileToUpdate.Path, fileToUpdate.Name}, "/"))
					color.Yellow("%s: %s -> %s", field, ch.From, ch.To)
					diffs = append(diffs, FileResourceDiff{Operation: strings.ToUpper(ch.Type), Target: field, UpdateValue: ch.To, FileResource: fileToUpdate})
				case "Name", "Path":
					color.Yellow("File %s will be updated as its configuration has changed:", strings.Join([]string{fileToUpdate.Path, fileToUpdate.Name}, "/"))
					color.Yellow("%s: %s -> %s", field, ch.From, ch.To)
					diffs = append(diffs, FileResourceDiff{Operation: "REPLACE", FileResource: fileToUpdate})
				}
			case "create":
				conv := ch.To.(FileSpecification)
				color.Yellow("File %s will be created", strings.Join([]string{conv.Path, conv.Name}, "/"))
				diffs = append(diffs, FileResourceDiff{Operation: strings.ToUpper(ch.Type), FileResource: ch.To.(FileSpecification)})
			}
		}
	} else {
		// If there were no diffs on file config, check for diff in file content
		for _, file := range files {
			localFileHash, err := exec.Command("sha1sum", file.Name).Output()
			if err != nil {
				log.Errorf("Error getting file hash: %s", err)
			}
			remoteFileHash, err := RunOnRemoteHost(credentials, fmt.Sprintf("sha1sum %s", strings.Join([]string{file.Path, file.Name}, "/")))
			if err != nil {
				log.Errorf("Error executing command: %s", err)
			}
			// Compare hash values for files to determine if there is a diff
			if strings.Split(string(localFileHash), " ")[0] == strings.Split(remoteFileHash, " ")[0] {
				color.Green("File %s unchanged", strings.Join([]string{file.Path, file.Name}, "/"))
			} else {
				color.Yellow("File %s will be updated in place as its contents has changed", strings.Join([]string{file.Path, file.Name}, "/"))
				diffs = append(diffs, FileResourceDiff{Operation: "REPLACE", FileResource: file})
			}
		}
	}
	return diffs
}

// GetPackageDiffs iterates through requested packages on a managed
// resource and shows and returns any diffs
func GetPackageDiffs(credentials Credentials, pkgs []PackageSpecification, fromState ManagedResource) []PackageResourceDiff {
	_, err := emoji.Printf(":wrench: %s\n", "Packages")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("-----------------------------------")
	var diffs []PackageResourceDiff
	for _, p := range pkgs {
		log.Infof("Determining state of package %s on host...", p.Package)
		command := fmt.Sprintf("dpkg-query --show %s", p.Package)
		result, err := RunOnRemoteHost(credentials, command)
		if err != nil {
			log.Errorf("Error executing command: %s", err)
		}
		if strings.Contains(result, fmt.Sprintf("dpkg-query: no packages found matching %s", p.Package)) {
			if p.Version != "" {
				color.Red("Package %s is not installed and will be installed using version %s", p.Package, p.Version)
				diffs = append(diffs, PackageResourceDiff{
					Operation:       "INSTALL",
					PackageResource: p,
				})
			} else {
				color.Red("Package %s is not installed and will be installed using latest version", p.Package)
				diffs = append(diffs, PackageResourceDiff{
					Operation:       "INSTALL",
					PackageResource: p,
				})
			}
		} else {
			command := fmt.Sprintf("dpkg-query --showformat='${Version}' --show %s", p.Package)
			result, err := RunOnRemoteHost(credentials, command)
			if err != nil {
				log.Errorf("Error executing command: %s", err)
			}
			if p.Version != "" && result == p.Version {
				color.Green("Package %s is installed and matches specified version %s", p.Package, p.Version)
				return []PackageResourceDiff{}
			} else if p.Version != "" && result != p.Version {
				color.Red("Package %s is installed at version %s and will be upgraded to %s", p.Package, result, p.Version)
				diffs = append(diffs, PackageResourceDiff{
					Operation:       "INSTALL",
					PackageResource: p,
				})
			} else if p.Version == "" {
				color.Green("Package %s is installed at version %s", p.Package, result)
			}
		}
	}
	return diffs
}

// Package Management

// InstallPackage installs a package on a managed resource
func InstallPackage(credentials Credentials, pkg PackageSpecification) {
	var command string
	if pkg.Version == "" || pkg.Version == "latest" {
		color.Green("Installing package %s with latest version...", pkg.Package)
		command = fmt.Sprintf("apt update && apt install -y %s", pkg.Package)
	} else {
		color.Green("Installing package %s with version %s...", pkg.Package, pkg.Version)
		command = fmt.Sprintf("apt update && apt install -y %s=%s", pkg.Package, pkg.Version)
	}
	result, err := RunOnRemoteHost(credentials, command)
	if err != nil {
		log.Errorf("Error executing command: %s", err)
	}
	fmt.Println(result)
}

// RemovePackage removes a package from a managed resource
func RemovePackage(credentials Credentials, pkg PackageSpecification) {
	command := fmt.Sprintf("apt remove -y %s", pkg.Package)
	result, err := RunOnRemoteHost(credentials, command)
	if err != nil {
		log.Errorf("Error executing command: %s", err)
	}
	fmt.Println(result)
}

// File management

// DeleteFile removes a file from a managed resource
func DeleteFile(credentials Credentials, file FileSpecification) {
	fileName := strings.Join([]string{file.Path, file.Name}, "/")
	command := fmt.Sprintf("rm %s && echo true || echo false", fileName)
	result, err := RunOnRemoteHost(credentials, command)
	if err != nil {
		log.Errorf("Error executing command: %s", err)
	}
	fmt.Println(result)
	if strings.Contains(result, "true") {
		color.Green("File %s removed successfully", fileName)
	} else {
		color.Red("Failed to delete file %s", fileName)
	}
}

// UpdateFileMode updates a file's permissions
func UpdateFileMode(credentials Credentials, file FileSpecification) {
	//target will either be content or mode
	fileName := strings.Join([]string{file.Path, file.Name}, "/")
	// Set file mode
	command := fmt.Sprintf("chmod -v %s %s && echo true || echo false", file.Mode, fileName)
	result, err := RunOnRemoteHost(credentials, command)
	if err != nil {
		log.Errorf("Error setting file mode: %s", err)
	}
	if strings.Contains(result, "true") {
		color.Green("File %s updated successfully", fileName)
	} else {
		color.Red("Failed to update file %s", fileName)
	}
}
