package connect

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"testing"

	"github.com/jerryagbesi/skipper/internal/sshconfig"
)

func fakeComander(exitCode int) Commander {
	return func(name string, args ...string) *exec.Cmd {
		cmd := exec.Command(os.Args[0], "-test.run=TestHelperProcess")
		cmd.Env = append(os.Environ(), "TEST_SSH_STUB=1",
			fmt.Sprintf("TEST_SSH_EXIT_CODE=%d", exitCode))

		return cmd
	}
}

func TestHelperProcess(t *testing.T) {
	if os.Getenv("TEST_SSH_STUB") != "1" {
		return
	}

	exitCode, _ := strconv.Atoi(os.Getenv("TEST_SSH_EXIT_CODE"))
	os.Exit(exitCode)
}

func TestConnect_WithAlias_BuildsCorrectCommand(t *testing.T) {
	host := &sshconfig.Host{
		Alias:    "bastion",
		Hostname: "10.0.0.1",
	}

	var capturedName string
	var capturedArgs []string

	fakeCommander := func(name string, args ...string) *exec.Cmd {
		capturedName = name
		capturedArgs = args
		// return a no-op command
		return exec.Command("true")
	}

	Connect(host, fakeCommander)

	if capturedName != "ssh" {
		t.Errorf("expected ssh, got %s", capturedName)
	}
	if capturedArgs[0] != "bastion" {
		t.Errorf("expected bastion, got %s", capturedArgs[0])
	}
}

func TestConnect_NoAlias_BuildsCorrectCommand(t *testing.T) {

	host := &sshconfig.Host{
		Alias:    "",
		Hostname: "10.0.0.1",
		User:     "ubuntu",
		Port:     2222,
	}

	var capturedArgs []string

	fakeCommander := func(name string, args ...string) *exec.Cmd {
		capturedArgs = args
		return exec.Command("true")
	}

	Connect(host, fakeCommander)

	if capturedArgs[0] != "ubuntu@10.0.0.1" {
		t.Errorf("expected ubuntu@10.0.0.1, got %s", capturedArgs[0])
	}
	if capturedArgs[2] != "2222" {
		t.Errorf("expected port 2222, got %s", capturedArgs[2])
	}
}

func TestConnect_EmptyHost_ReturnsError(t *testing.T) {
	host := &sshconfig.Host{}

	fakeCommander := func(name string, args ...string) *exec.Cmd {
		t.Error("commander should not be called for empty host")
		return nil
	}

	err := Connect(host, fakeCommander)
	if err == nil {
		t.Error("expected error for empty host, got nil")
	}
}
