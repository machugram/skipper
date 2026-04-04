package connect

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"

	"github.com/jerryagbesi/skipper/internal/sshconfig"
)

func Connect(host *sshconfig.Host) error {
	target := host.Alias

	var cmd *exec.Cmd

	if target != "" {
		cmd = exec.Command("ssh", host.Alias)
	} else {
		target = host.Hostname
		cmd = exec.Command("ssh", host.User+"@"+host.Hostname, "-p", strconv.Itoa(host.Port))
	}

	fmt.Printf("Connecting to %s\n", target)

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}
