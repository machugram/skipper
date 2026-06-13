package cmd

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jerryagbesi/skipper/internal/sshconfig"
)



func TestRemoveCommandDeletesHost(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config")

	for _, h := range []sshconfig.Host{
		{Alias: "bastion", Hostname: "10.0.0.1", User: "admin"},
		{Alias: "worker", Hostname: "10.0.0.2", User: "ubuntu"},
	} {
		if _, _, err := sshconfig.AddHost(path, h); err != nil {
			t.Fatalf("setup: %v", err)
		}
	}

	out := new(bytes.Buffer)
	rootCmd.SetOut(out)
	rootCmd.SetErr(out)
	rootCmd.SetArgs([]string{"remove", "bastion", "-c", path})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("expected remove to succeed, got %v", err)
	}

	hosts, err := sshconfig.ParseHosts(path)
	if err != nil {
		t.Fatalf("expected config to parse, got %v", err)
	}

	if len(hosts) != 1 || hosts[0].Alias != "worker" {
		t.Fatalf("expected only 'worker' remaining, got %+v", hosts)
	}

}

func TestRemoveCommandAliasRm(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config")

	if _, _, err := sshconfig.AddHost(path, sshconfig.Host{Alias: "bastion", Hostname: "10.0.0.1", User: "admin"}); err != nil {
		t.Fatalf("setup: %v", err)
	}

	out := new(bytes.Buffer)
	rootCmd.SetOut(out)
	rootCmd.SetErr(out)
	rootCmd.SetArgs([]string{"rm", "bastion", "-c", path})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("expected rm alias to succeed, got %v", err)
	}
}

func TestRemoveCommandErrorsWhenNotFound(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config")

	if _, _, err := sshconfig.AddHost(path, sshconfig.Host{Alias: "worker", Hostname: "10.0.0.2", User: "ubuntu"}); err != nil {
		t.Fatalf("setup: %v", err)
	}

	out := new(bytes.Buffer)
	rootCmd.SetOut(out)
	rootCmd.SetErr(out)
	rootCmd.SetArgs([]string{"remove", "nonexistent", "-c", path})

	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error when alias not found")
	}

	if !strings.Contains(err.Error(), "not found") {
		t.Fatalf("expected 'not found' error, got %v", err)
	}
}

func TestRemoveCommandForceIgnoresMissingAlias(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config")

	if _, _, err := sshconfig.AddHost(path, sshconfig.Host{Alias: "worker", Hostname: "10.0.0.2", User: "ubuntu"}); err != nil {
		t.Fatalf("setup: %v", err)
	}

	out := new(bytes.Buffer)
	rootCmd.SetOut(out)
	rootCmd.SetErr(out)
	rootCmd.SetArgs([]string{"remove", "nonexistent", "--force", "-c", path})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("expected --force to suppress not-found error, got %v", err)
	}
}
