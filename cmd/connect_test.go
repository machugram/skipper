package cmd

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/jerryagbesi/skipper/internal/sshconfig"
)

func TestFindHostByAliasReturnsMatchingHost(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config")

	for _, h := range []sshconfig.Host{
		{Alias: "bastion", Hostname: "10.0.0.1", User: "admin"},
		{Alias: "worker", Hostname: "10.0.0.2", User: "ubuntu"},
	} {
		if _, _, err := sshconfig.AddHost(path, h); err != nil {
			t.Fatalf("setup: %v", err)
		}
	}

	host, err := findHostByAlias(path, "bastion")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if host.Alias != "bastion" || host.Hostname != "10.0.0.1" || host.User != "admin" {
		t.Fatalf("unexpected host returned: %+v", host)
	}
}

func TestFindHostByAliasErrorsWhenNotFound(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config")

	if _, _, err := sshconfig.AddHost(path, sshconfig.Host{Alias: "worker", Hostname: "10.0.0.2", User: "ubuntu"}); err != nil {
		t.Fatalf("setup: %v", err)
	}

	_, err := findHostByAlias(path, "nonexistent")
	if err == nil {
		t.Fatal("expected error for missing alias")
	}

	if !strings.Contains(err.Error(), "nonexistent") {
		t.Fatalf("expected alias in error message, got %v", err)
	}
}
