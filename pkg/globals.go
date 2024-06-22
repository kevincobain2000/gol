package pkg

import (
	"golang.org/x/crypto/ssh"
)

var GlobalFilePaths []FileInfo
var GlobalPipeTmpFilePath string
var GlobalPathSSHConfig []SSHPathConfig
var GlobalSSHClients = make(map[string]*ssh.Client)
