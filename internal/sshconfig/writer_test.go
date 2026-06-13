package sshconfig

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAddHostRejectsUnsafeWhitespaceInWrittenFields(t *testing.T) {
	tests := []struct {
		name        string
		host        Host
		wantMessage string
	}{
		{
			name: "alias with space",
			host: Host{
				Alias:    "jump box",
				Hostname: "example.com",
				User:     "alice",
			},
			wantMessage: "alias cannot contain whitespace",
		},
		{
			name: "hostname with newline",
			host: Host{
				Alias:    "jump-box",
				Hostname: "example.com\nProxyCommand yes",
				User:     "alice",
			},
			wantMessage: "host name cannot contain whitespace",
		},
		{
			name: "user with tab",
			host: Host{
				Alias:    "jump-box",
				Hostname: "example.com",
				User:     "alice\tadmin",
			},
			wantMessage: "user cannot contain whitespace",
		},
		{
			name: "identity file with space",
			host: Host{
				Alias:        "jump-box",
				Hostname:     "example.com",
				User:         "alice",
				IdentityFile: "/Users/test/.ssh/my key",
			},
			wantMessage: "identity file cannot contain whitespace",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			configPath := filepath.Join(t.TempDir(), "config")

			_, _, err := AddHost(configPath, test.host)
			if err == nil {
				t.Fatal("expected error, got nil")
			}

			if !strings.Contains(err.Error(), test.wantMessage) {
				t.Fatalf("expected error containing %q, got %v", test.wantMessage, err)
			}

			content, readErr := os.ReadFile(configPath)
			if !os.IsNotExist(readErr) {
				t.Fatalf("expected config file to not be created, readErr=%v content=%q", readErr, string(content))
			}
		})
	}
}

func TestAddHostRejectsUnsafeWhitespaceInResolvedAlias(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config")

	_, _, err := AddHost(configPath, Host{
		Hostname: "example host",
		User:     "alice",
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "alias cannot contain whitespace") {
		t.Fatalf("expected resolved alias validation error, got %v", err)
	}
}

func TestAddHostWritesSafeIdentityFile(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config")

	_, _, err := AddHost(configPath, Host{
		Alias:        "jump-box",
		Hostname:     "example.com",
		User:         "alice",
		IdentityFile: "/Users/test/.ssh/id_ed25519",
	})
	if err != nil {
		t.Fatalf("expected host to be added, got %v", err)
	}

	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("expected config to be readable, got %v", err)
	}

	if !strings.Contains(string(content), "  IdentityFile /Users/test/.ssh/id_ed25519\n") {
		t.Fatalf("expected identity file to be written, got content:\n%s", string(content))
	}
}

func TestAddHostRejectsDuplicateAliasWithDifferentIdentityFile(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config")

	_, _, err := AddHost(configPath, Host{
		Alias:        "jump-box",
		Hostname:     "example.com",
		User:         "alice",
		IdentityFile: "/Users/test/.ssh/id_ed25519",
	})
	if err != nil {
		t.Fatalf("expected first host to be added, got %v", err)
	}

	_, _, err = AddHost(configPath, Host{
		Alias:        "jump-box",
		Hostname:     "example.com",
		User:         "alice",
		IdentityFile: "/Users/test/.ssh/id_rsa",
	})
	if err == nil {
		t.Fatal("expected duplicate alias with different identity file to fail")
	}

	if !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("expected duplicate alias error, got %v", err)
	}
}

func TestRemoveHostRemovesMatchingBlock(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config")

	for _, h := range []Host{
		{Alias: "bastion", Hostname: "10.0.0.1", User: "admin"},
		{Alias: "worker", Hostname: "10.0.0.2", User: "ubuntu"},
	} {
		if _, _, err := AddHost(configPath, h); err != nil {
			t.Fatalf("setup: %v", err)
		}
	}

	removed, err := RemoveHost(configPath, "bastion")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !removed {
		t.Fatal("expected removed=true")
	}

	hosts, err := ParseHosts(configPath)
	if err != nil {
		t.Fatalf("expected config to parse, got %v", err)
	}
	if len(hosts) != 1 || hosts[0].Alias != "worker" {
		t.Fatalf("expected only 'worker' remaining, got %+v", hosts)
	}
}

func TestRemoveHostReturnsFalseForMissingAlias(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config")

	if _, _, err := AddHost(configPath, Host{Alias: "worker", Hostname: "10.0.0.2", User: "ubuntu"}); err != nil {
		t.Fatalf("setup: %v", err)
	}

	removed, err := RemoveHost(configPath, "nonexistent")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if removed {
		t.Fatal("expected removed=false for nonexistent alias")
	}
}

func TestRemoveHostReturnsFalseWhenFileAbsent(t *testing.T) {
	removed, err := RemoveHost(filepath.Join(t.TempDir(), "config"), "anything")
	if err != nil {
		t.Fatalf("expected no error for missing file, got %v", err)
	}
	if removed {
		t.Fatal("expected removed=false when file does not exist")
	}
}

func TestRemoveHostRequiresAlias(t *testing.T) {
	_, err := RemoveHost(filepath.Join(t.TempDir(), "config"), "   ")
	if err == nil {
		t.Fatal("expected error for blank alias")
	}
}
