package sshconfig

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/kevinburke/ssh_config"
)

type Host struct {
	Alias        string
	Hostname     string
	User         string
	Port         int
	IdentityFile string
}

func DefaultConfigPath() (string, error) {
	if home, err := os.UserHomeDir(); err != nil {
		return "", fmt.Errorf("error resolving home directory: %w", err)
	} else {
		return filepath.Join(home, ".ssh", "config"), nil
	}
}

func ParseHosts(path string) ([]Host, error) {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []Host{}, fmt.Errorf("error: %w", err)
		}
		return nil, fmt.Errorf("failed to open config file %q: %w", path, err)
	}

	defer f.Close()

	cfg, err := ssh_config.Decode(f)
	if err != nil {
		return nil, fmt.Errorf("failed to decode config file %q: %w", path, err)
	}

	var hosts []Host

	for _, host := range cfg.Hosts {
		for _, pattern := range host.Patterns { //incase a user has multiple patterns for a host eg. Host bastion jump-box *.staging
			alias := pattern.String()

			if alias == "*" {
				continue
			}

			hostname, _ := cfg.Get(alias, "Hostname")
			user, _ := cfg.Get(alias, "User")
			identityFile, _ := cfg.Get(alias, "IdentityFile")

			var port int
			if raw, _ := cfg.Get(alias, "Port"); raw != "" {
				if p, err := strconv.Atoi(raw); err == nil {
					port = p
				} else {
					return nil, fmt.Errorf("failed to parse port %q: %w", raw, err)
				}
			}

			hosts = append(hosts, Host{
				Alias:        alias,
				Hostname:     hostname,
				User:         user,
				Port:         port,
				IdentityFile: identityFile,
			})

		}
	}

	return hosts, nil

}
