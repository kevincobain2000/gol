package pkg

import "golang.org/x/crypto/ssh"

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
	client := GlobalSSHClients[key]
	if client == nil {
		c, err := sshConnect(config)
		if err != nil {
			return nil, err
		}
		GlobalSSHClients[key] = c
		client = c
	}
	return client, nil
}
