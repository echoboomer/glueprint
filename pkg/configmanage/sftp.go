package configmanage

import (
	"io"
	"os"
	"strings"

	"github.com/fatih/color"
	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"

	"github.com/pkg/sftp"
)

// UploadFileViaSFTP leverages sftp to place a file onto a host
func UploadFileViaSFTP(credentials Credentials, file FileSpecification) error {
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
		return err
	}
	defer client.Close()

	// Instantiate SFTP Client
	sftpClient, err := sftp.NewClient(client)
	if err != nil {
		log.Fatal(err)
	}
	defer sftpClient.Close()

	fileName := strings.Join([]string{file.Path, file.Name}, "/")

	// Create destination file
	dstFile, err := sftpClient.Create(fileName)
	if err != nil {
		log.Fatal(err)
	}
	defer dstFile.Close()

	// Parse source file
	srcFile, err := os.Open(file.Name)
	if err != nil {
		log.Fatal(err)
	}

	// Copy source to destination
	bytes, err := io.Copy(dstFile, srcFile)
	if err != nil {
		log.Fatal(err)
	}
	color.Green("%d bytes copied to %s on host", bytes, strings.Join([]string{file.Path, file.Name}, "/"))

	return nil
}
