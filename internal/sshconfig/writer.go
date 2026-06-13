package sshconfig

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

func AddHost(path string, host Host) (addedHost *Host, created bool, err error) {
	if strings.TrimSpace(host.Hostname) == "" {
		return nil, false, fmt.Errorf("host name is required")
	}

	if strings.TrimSpace(host.User) == "" {
		return nil, false, fmt.Errorf("user is required")
	}

	host.Alias = resolveAlias(host)
	if err := validateHostFields(host); err != nil {
		return nil, false, err
	}

	// Read the file once; use the bytes for both the duplicate check and the
	// trailing-newline check below, avoiding a second open/read.
	currentContent, err := os.ReadFile(path)
	fileExists := err == nil
	if err != nil && !os.IsNotExist(err) {
		return nil, false, fmt.Errorf("failed to read config file %q: %w", path, err)
	}

	if fileExists && len(currentContent) > 0 {
		existingHosts, err := parseHostsBytes(currentContent)
		if err != nil {
			return nil, false, fmt.Errorf("failed to parse config file %q: %w", path, err)
		}
		for _, existingHost := range existingHosts {
			if existingHost.Alias != host.Alias {
				continue
			}
			if sameHostSettings(existingHost, host) {
				return &existingHost, false, nil
			}
			return nil, false, fmt.Errorf("host %q already exists in %s with different settings", host.Alias, path)
		}
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, false, fmt.Errorf("failed to create config directory: %w", err)
	}

	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return nil, false, fmt.Errorf("failed to open config file %q: %w", path, err)
	}
	defer func() {
		if cerr := file.Close(); cerr != nil {
			if err == nil {
				err = fmt.Errorf("failed to close config file %q: %w", path, cerr)
			} else {
				err = fmt.Errorf("%w; close error for %q: %w", err, path, cerr)
			}
		}
	}()

	if len(currentContent) > 0 && !bytes.HasSuffix(currentContent, []byte("\n")) {
		if _, err := file.WriteString("\n"); err != nil {
			return nil, false, fmt.Errorf("failed to prepare config file %q: %w", path, err)
		}
	}

	if len(bytes.TrimSpace(currentContent)) > 0 {
		if _, err := file.WriteString("\n"); err != nil {
			return nil, false, fmt.Errorf("failed to separate config entries in %q: %w", path, err)
		}
	}

	if _, err := file.WriteString(formatHostEntry(host)); err != nil {
		return nil, false, fmt.Errorf("failed to write host %q to %s: %w", host.Alias, path, err)
	}

	addedHost = &host
	created = true
	return addedHost, created, nil
}

// RemoveHost removes the Host block matching alias from the config at path.
// Returns (true, nil) when the host was found and removed, (false, nil) when
// no such alias exists, or (false, err) on I/O failure.
func RemoveHost(path, alias string) (removed bool, err error) {
	alias = strings.TrimSpace(alias)
	if alias == "" {
		return false, fmt.Errorf("alias is required")
	}

	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to read config file %q: %w", path, err)
	}

	out, found := removeHostBlock(content, alias)
	if !found {
		return false, nil
	}

	if err := os.WriteFile(path, out, 0o600); err != nil {
		return false, fmt.Errorf("failed to write config file %q: %w", path, err)
	}
	return true, nil
}

// removeHostBlock strips the Host block whose first token matches alias from
// src and returns the rewritten content.  A "block" is the Host line plus all
// indented/blank lines that follow it, up to (but not including) the next
// non-indented, non-blank line.
func removeHostBlock(src []byte, alias string) ([]byte, bool) {
	scanner := bufio.NewScanner(bytes.NewReader(src))

	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	targetLine := -1
	for i, l := range lines {
		if isHostLine(l, alias) {
			targetLine = i
			break
		}
	}
	if targetLine == -1 {
		return src, false
	}

	// End of block: next Host line, next non-blank non-indented line, or EOF.
	blockEnd := len(lines)
	for i := targetLine + 1; i < len(lines); i++ {
		t := lines[i]
		if isAnyHostLine(t) {
			blockEnd = i
			break
		}
		if t != "" && t[0] != ' ' && t[0] != '\t' {
			blockEnd = i
			break
		}
	}

	// Drop blank lines immediately before the block so removal doesn't stack blanks.
	start := targetLine
	for start > 0 && strings.TrimSpace(lines[start-1]) == "" {
		start--
	}

	var buf bytes.Buffer
	for i, l := range lines {
		if i >= start && i < blockEnd {
			continue
		}
		buf.WriteString(l)
		buf.WriteByte('\n')
	}
	return bytes.TrimRight(buf.Bytes(), "\n"), true
}

// isHostLine reports whether line declares alias as one of its patterns.
// Uses index-based scanning on the keyword to avoid a strings.Fields allocation.
func isHostLine(line, alias string) bool {
	line = strings.TrimLeft(line, " \t")
	i := strings.IndexAny(line, " \t")
	if i <= 0 || !strings.EqualFold(line[:i], "Host") {
		return false
	}
	for _, p := range strings.Fields(line[i:]) {
		if p == alias {
			return true
		}
	}
	return false
}

// isAnyHostLine reports whether line is any Host declaration.
func isAnyHostLine(line string) bool {
	line = strings.TrimLeft(line, " \t")
	i := strings.IndexAny(line, " \t")
	return i > 0 && strings.EqualFold(line[:i], "Host")
}

func formatHostEntry(host Host) string {
	var b strings.Builder
	b.WriteString("Host ")
	b.WriteString(host.Alias)
	b.WriteString("\n  HostName ")
	b.WriteString(host.Hostname)
	b.WriteString("\n  User ")
	b.WriteString(host.User)
	b.WriteByte('\n')
	if host.Port > 0 {
		fmt.Fprintf(&b, "  Port %d\n", host.Port)
	}
	if host.IdentityFile != "" {
		b.WriteString("  IdentityFile ")
		b.WriteString(host.IdentityFile)
		b.WriteByte('\n')
	}
	return b.String()
}

func resolveAlias(host Host) string {
	if alias := strings.TrimSpace(host.Alias); alias != "" {
		return alias
	}

	if host.Port > 0 {
		return fmt.Sprintf("%s-%d", host.Hostname, host.Port)
	}

	return host.Hostname
}

func sameHostSettings(existingHost, requestedHost Host) bool {
	return existingHost.Hostname == requestedHost.Hostname &&
		existingHost.User == requestedHost.User &&
		existingHost.Port == requestedHost.Port &&
		normalizeIdentityFile(existingHost.IdentityFile) == normalizeIdentityFile(requestedHost.IdentityFile)
}

func normalizeIdentityFile(identityFile string) string {
	return strings.TrimSpace(identityFile)
}

func validateHostFields(host Host) error {
	// Stack-allocated array: no heap allocation unlike a []struct{} literal.
	for _, f := range [4]struct {
		name     string
		value    string
		required bool
	}{
		{"alias", host.Alias, true},
		{"host name", host.Hostname, true},
		{"user", host.User, true},
		{"identity file", host.IdentityFile, false},
	} {
		if f.required && strings.TrimSpace(f.value) == "" {
			return fmt.Errorf("%s is required", f.name)
		}
		if containsUnsafeWhitespace(f.value) {
			return fmt.Errorf("%s cannot contain whitespace", f.name)
		}
	}
	return nil
}

func containsUnsafeWhitespace(value string) bool {
	return strings.ContainsFunc(value, unicode.IsSpace)
}
