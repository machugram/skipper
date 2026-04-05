package connect

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"

	"github.com/jerryagbesi/skipper/internal/sshconfig"
)

type Commander func(name string, args ...string) *exec.Cmd //allows for mocking exec.Command

func Connect(host *sshconfig.Host, commander Commander) error {
	target := host.Alias

	var cmd *exec.Cmd

	if target != "" {
		cmd = commander("ssh", host.Alias)
	} else if host.Hostname != "" {
		target = host.Hostname
		cmd = commander("ssh", host.User+"@"+host.Hostname, "-p", strconv.Itoa(host.Port))
	} else {
		return fmt.Errorf("no target specified")
	}

	fmt.Printf("Connecting to %s\n", target)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
