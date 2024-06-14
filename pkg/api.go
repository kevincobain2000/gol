package pkg

type API struct {
}

func NewAPI() *API {
	return &API{}
}

func (a *API) FindSSHConfig(host string) *SSHPathConfig {
	for _, sshConfig := range GlobalPathSSHConfig {
		if sshConfig.Host == host {
			return &sshConfig
		}
	}
	return nil
}
