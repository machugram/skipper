# Skipper

A CLI tool for managing SSH connections with an interactive terminal UI. Skipper reads your `~/.ssh/config` file and lets you browse, search, and connect to hosts without memorizing aliases.

## Features

- **Interactive host selection** -- filterable list of SSH hosts powered by [Bubbletea](https://github.com/charmbracelet/bubbletea)
- **SSH config parsing** -- reads host aliases, users, hostnames, ports, and identity files from your SSH config
- **Fuzzy search** -- quickly narrow down hosts by typing
- **Seamless connection** -- selects a host and drops you straight into an SSH session

## Prerequisites

- [Go](https://go.dev/) 1.25+
- An SSH config file (typically `~/.ssh/config`)

## Getting Started

```bash
# Clone the repo
git clone https://github.com/JerryAgbesi/skipper.git
cd skipper

# Build and run
make build
./skipper

# Or run directly
make run
```

## Usage

```
skipper [flags]
```

| Flag | Description |
|------|-------------|
| `-c, --config <path>` | Path to SSH config file (default: `~/.ssh/config`) |
| `-v, --version` | Print version |
| `-h, --help` | Show help |

### Keyboard Controls

| Key | Action |
|-----|--------|
| `Enter` | Connect to selected host |
| `Up/Down` or `j/k` | Navigate the list |
| / | Start filtering hosts |
| `Esc` / `Ctrl+C` / `Q` | Quit |

## Make Targets

| Target | Description |
|--------|-------------|
| `make build` | Compile the `skipper` binary |
| `make run` | Build and run |
| `make lint` | Run golangci-lint |
| `make fmt` | Format code |
| `make all` | Format + Build + Run |

## Project Structure

```
skipper/
├── main.go                  # Entry point
├── cmd/root.go              # CLI command definition
├── internal/
│   ├── connect/connect.go   # SSH session execution
│   ├── sshconfig/parser.go  # SSH config file parser
│   └── ui/model.go          # Interactive terminal UI
├── Makefile
└── go.mod
```
