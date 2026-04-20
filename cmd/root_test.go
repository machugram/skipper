package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jerryagbesi/skipper/internal/sshconfig"
	"github.com/jerryagbesi/skipper/internal/ui"
	"github.com/spf13/cobra"
)

func TestFilterHostsReturnsOriginalListForBlankQuery(t *testing.T) {
	hosts := []sshconfig.Host{{Alias: "dev"}, {Alias: "prod"}}

	filtered := filterHosts(hosts, "   ")

	if len(filtered) != len(hosts) {
		t.Fatalf("expected %d hosts, got %d", len(hosts), len(filtered))
	}
}

func TestFilterHostsMatchesAliasHostnameUserAndPort(t *testing.T) {
	hosts := []sshconfig.Host{
		{Alias: "dev-api", Hostname: "10.0.0.4", User: "ubuntu", Port: 22},
		{Alias: "prod-db", Hostname: "db.internal", User: "postgres", Port: 5432},
	}

	tests := []struct {
		name     string
		query    string
		expected string
	}{
		{name: "matches alias", query: "DEV", expected: "dev-api"},
		{name: "matches hostname", query: "internal", expected: "prod-db"},
		{name: "matches user", query: "ubuntu", expected: "dev-api"},
		{name: "matches port", query: "5432", expected: "prod-db"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			filtered := filterHosts(hosts, test.query)
			if len(filtered) != 1 {
				t.Fatalf("expected 1 host, got %d", len(filtered))
			}

			if filtered[0].Alias != test.expected {
				t.Fatalf("expected %q, got %q", test.expected, filtered[0].Alias)
			}
		})
	}
}

func TestPrepareHostSelectionStartsFilteringWhenFindHasNoTerm(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().String("find", "", "")
	if err := cmd.Flags().Set("find", ""); err != nil {
		t.Fatalf("expected flag set to succeed, got %v", err)
	}

	findQuery = ""
	hosts := []sshconfig.Host{{Alias: "dev"}}

	options, filtered, err := prepareHostSelection(cmd, hosts)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !options.StartFiltering {
		t.Fatal("expected filtering mode to start")
	}

	if len(filtered) != 1 || filtered[0].Alias != "dev" {
		t.Fatalf("unexpected filtered hosts: %+v", filtered)
	}
}

func TestPrepareHostSelectionReturnsErrorWhenNoMatch(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().String("find", "", "")
	if err := cmd.Flags().Set("find", "prod"); err != nil {
		t.Fatalf("expected flag set to succeed, got %v", err)
	}

	findQuery = "staging"
	_, _, err := prepareHostSelection(cmd, []sshconfig.Host{{Alias: "dev"}})
	if err == nil {
		t.Fatal("expected error when no hosts match")
	}
}

func TestAddHostWritesAliasAndTarget(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config")

	host, err := addHost(configPath, "devone", []string{"user@10.0.0.8:9000"})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if host.Alias != "devone" {
		t.Fatalf("expected alias devone, got %q", host.Alias)
	}

	hosts, err := sshconfig.ParseHosts(configPath)
	if err != nil {
		t.Fatalf("expected config to parse, got %v", err)
	}

	if len(hosts) != 1 {
		t.Fatalf("expected 1 host, got %d", len(hosts))
	}

	if hosts[0].Alias != "devone" || hosts[0].User != "user" || hosts[0].Hostname != "10.0.0.8" || hosts[0].Port != 9000 {
		t.Fatalf("unexpected host written: %+v", hosts[0])
	}
}

func TestAddHostIsIdempotentForSameAliasAndTarget(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config")

	firstHost, err := addHost(configPath, "devone", []string{"user@10.0.0.8:9000"})
	if err != nil {
		t.Fatalf("expected first add to succeed, got %v", err)
	}

	secondHost, err := addHost(configPath, "devone", []string{"user@10.0.0.8:9000"})
	if err != nil {
		t.Fatalf("expected second add to succeed, got %v", err)
	}

	if secondHost.Alias != firstHost.Alias || secondHost.User != firstHost.User || secondHost.Hostname != firstHost.Hostname || secondHost.Port != firstHost.Port {
		t.Fatalf("expected same host back, got first=%+v second=%+v", firstHost, secondHost)
	}

	hosts, err := sshconfig.ParseHosts(configPath)
	if err != nil {
		t.Fatalf("expected config to parse, got %v", err)
	}

	if len(hosts) != 1 {
		t.Fatalf("expected 1 host after duplicate add, got %d", len(hosts))
	}

	content, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("expected config to be readable, got %v", err)
	}

	if strings.Count(string(content), "Host devone\n") != 1 {
		t.Fatalf("expected single host entry, got content:\n%s", string(content))
	}
}

func TestAddHostRejectsDuplicateAliasWithDifferentTarget(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config")

	if _, err := addHost(configPath, "devone", []string{"user@10.0.0.8:9000"}); err != nil {
		t.Fatalf("expected first add to succeed, got %v", err)
	}

	_, err := addHost(configPath, "devone", []string{"user@10.0.0.9:9000"})
	if err == nil {
		t.Fatal("expected duplicate alias with different target to fail")
	}

	if !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("expected duplicate alias error, got %v", err)
	}
}

func TestAddHostRequiresAlias(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config")

	_, err := addHost(configPath, "   ", []string{"user@10.0.0.8:9000"})
	if err == nil {
		t.Fatal("expected error for missing alias")
	}

	if !strings.Contains(err.Error(), "--add <alias> <user@host[:port]>") {
		t.Fatalf("expected usage hint in error, got %v", err)
	}
}

func TestAddHostRequiresExactlyOneTarget(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config")

	_, err := addHost(configPath, "devone", nil)
	if err == nil {
		t.Fatal("expected error for missing target")
	}

	if !strings.Contains(err.Error(), "--add <alias> <user@host[:port]>") {
		t.Fatalf("expected usage hint in error, got %v", err)
	}
}

func TestAddHostRejectsInvalidTarget(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config")

	_, err := addHost(configPath, "devone", []string{"invalid-target"})
	if err == nil {
		t.Fatal("expected error for invalid target")
	}
}

func TestResolveConfigPathReturnsExplicitPath(t *testing.T) {
	explicitPath := filepath.Join(t.TempDir(), "config")

	path, err := resolveConfigPath(explicitPath)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if path != explicitPath {
		t.Fatalf("expected %q, got %q", explicitPath, path)
	}
}

func TestGenerateManPagesWritesRootManPage(t *testing.T) {
	outDir := t.TempDir()

	if err := generateManPages(outDir); err != nil {
		t.Fatalf("expected man page generation to succeed, got %v", err)
	}

	content, err := os.ReadFile(filepath.Join(outDir, "skipper.1"))
	if err != nil {
		t.Fatalf("expected generated man page to exist, got %v", err)
	}

	text := string(content)
	if !strings.Contains(text, "skipper") || !strings.Contains(text, "SSH") {
		t.Fatalf("expected man page to describe skipper, got:\n%s", text)
	}
}

func TestFindHostReturnsMatchingHost(t *testing.T) {
	hosts := []sshconfig.Host{
		{Alias: "dev", Hostname: "10.0.0.1"},
		{Alias: "prod", Hostname: "10.0.0.2"},
	}

	host, err := findHost(hosts, "dev")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if host.Alias != "dev" {
		t.Fatalf("expected alias dev, got %q", host.Alias)
	}
}

func TestFindHostIsCaseInsensitive(t *testing.T) {
	hosts := []sshconfig.Host{{Alias: "DevBox", Hostname: "10.0.0.1"}}

	host, err := findHost(hosts, "devbox")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if host.Alias != "DevBox" {
		t.Fatalf("expected alias DevBox, got %q", host.Alias)
	}
}

func TestFindHostReturnsErrorWhenNotFound(t *testing.T) {
	hosts := []sshconfig.Host{{Alias: "dev"}}

	_, err := findHost(hosts, "staging")
	if err == nil {
		t.Fatal("expected error for unknown alias")
	}

	if !strings.Contains(err.Error(), "not found") {
		t.Fatalf("expected not-found error, got %v", err)
	}
}

func TestRunOptionsZeroValueDoesNotStartFiltering(t *testing.T) {
	options := ui.RunOptions{}
	if options.StartFiltering {
		t.Fatal("expected zero-value run options to keep filtering disabled")
	}
}
