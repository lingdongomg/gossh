# GoSSH

A TUI (Terminal User Interface) SSH connection manager built with Go and [Bubbletea](https://github.com/charmbracelet/bubbletea).

[中文文档](README_CN.md)

## Features

### Core Features
- **Connection Management** - Add, edit, delete SSH connections with TUI
- **Secure Storage** - Master password protection with AES-256-GCM encryption
- **Group Organization** - Organize connections into groups
- **Search** - Real-time search and filter connections
- **Import/Export** - YAML-based backup and restore

### Advanced Features (v2.0)
- **SFTP File Transfer** - Interactive SFTP shell for file operations
- **Port Forwarding** - Local (-L) and remote (-R) port forwarding
- **Batch Execution** - Execute commands on multiple servers simultaneously

## Installation

### From Source

```bash
git clone https://github.com/yourusername/gossh.git
cd gossh
go build -o gossh .
```

### Using Go Install

```bash
go install github.com/yourusername/gossh@latest
```

## Quick Start

```bash
# Start the TUI application
./gossh

# Show help
./gossh help

# List all connections
./gossh list

# Connect to a server by name
./gossh connect myserver
```

## Usage

### TUI Mode

Launch the interactive TUI:

```bash
./gossh
```

#### Key Bindings

| Key | Action |
|-----|--------|
| `↑/k` | Move up |
| `↓/j` | Move down |
| `g/G` | Jump to top/bottom |
| `/` | Search connections |
| `Enter` | Connect to selected server |
| `a` | Add new connection |
| `e` | Edit selected connection |
| `d` | Delete selected connection |
| `?` | Show help |
| `q` | Quit |

### CLI Mode

#### Basic Commands

```bash
# Show version
gossh version

# List all connections
gossh list

# Connect by name
gossh connect <name>

# Export connections to file
gossh export [filename]

# Import connections from file
gossh import <filename>
```

#### SFTP Session

```bash
gossh sftp <connection-name>
```

SFTP shell commands:
- `ls [path]` - List directory contents
- `cd <path>` - Change directory
- `pwd` - Print working directory
- `get <remote> [local]` - Download file
- `put <local> [remote]` - Upload file
- `mkdir <path>` - Create directory
- `rm <path>` - Remove file
- `rmdir <path>` - Remove directory recursively
- `exit/quit` - Exit SFTP session

#### Port Forwarding

```bash
# Local port forwarding (-L)
# Forward local port 3306 to remote localhost:3306
gossh forward <name> -L 3306:localhost:3306

# Remote port forwarding (-R)
# Forward remote port 8080 to local localhost:80
gossh forward <name> -R 8080:localhost:80
```

#### Batch Execution

Execute commands on multiple servers:

```bash
# Execute on all servers in a group
gossh exec "uptime" --group=Production

# Execute on servers with specific tags
gossh exec "df -h" --tags=web,nginx

# Execute on specific servers by name
gossh exec "hostname" --names=server1,server2

# Set custom timeout (default: 30s)
gossh exec "long-running-command" --group=All --timeout=120
```

## Configuration

Configuration is stored in YAML format:

- **Linux/macOS**: `~/.config/gossh/config.yaml`
- **Windows**: `%APPDATA%\gossh\config.yaml`

### Connection Fields

| Field | Description |
|-------|-------------|
| `name` | Connection name (unique identifier) |
| `host` | Server hostname or IP |
| `port` | SSH port (default: 22) |
| `user` | Username |
| `password` | Password (encrypted) |
| `key_path` | Path to SSH private key |
| `key_passphrase` | Passphrase for key (encrypted) |
| `group` | Group name for organization |
| `tags` | List of tags for filtering |
| `startup_command` | Command to run after connection |

## Security

- **Master Password**: Required on first run, uses Argon2id key derivation
- **Encryption**: AES-256-GCM for storing sensitive data (passwords, key passphrases)
- **Host Key Verification**: Uses system known_hosts (configurable)

## Dependencies

- [Bubbletea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Bubbles](https://github.com/charmbracelet/bubbles) - TUI components
- [Lipgloss](https://github.com/charmbracelet/lipgloss) - Styling
- [golang.org/x/crypto](https://pkg.go.dev/golang.org/x/crypto) - SSH and cryptography
- [pkg/sftp](https://github.com/pkg/sftp) - SFTP support
- [gopkg.in/yaml.v3](https://github.com/go-yaml/yaml) - Configuration

## License

MIT License
