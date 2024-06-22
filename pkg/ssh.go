package pkg

import (
	"sync"

	"golang.org/x/crypto/ssh"
)

var (
	clientMutex = &sync.Mutex{}
)

func NewSession(config *SSHConfig) (*ssh.Session, error) {
	client, err := NewOrReusableClient(config)
	if err != nil {
		return nil, err
	}
	session, err := client.NewSession()
	if err != nil {
		return nil, err
	}
	return session, nil
}

func NewOrReusableClient(config *SSHConfig) (*ssh.Client, error) {
	key := config.Host + ":" + config.Port

	clientMutex.Lock()
	client := GlobalSSHClients[key]
	clientMutex.Unlock()

	if client == nil {
		c, err := sshConnect(config)
		if err != nil {
			return nil, err
		}
		clientMutex.Lock()
		GlobalSSHClients[key] = c
		clientMutex.Unlock()
		client = c
	}

	// check if client is still connected
	_, _, err := client.SendRequest("", true, nil)
	if err != nil {
		clientMutex.Lock()
		delete(GlobalSSHClients, key)
		clientMutex.Unlock()
		return NewOrReusableClient(config)
	}

	return client, nil
}
