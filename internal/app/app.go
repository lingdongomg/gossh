package app

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"golang.org/x/term"
	tea "github.com/charmbracelet/bubbletea"
	"gopkg.in/yaml.v3"
	"gossh/internal/config"
	"gossh/internal/model"
	"gossh/internal/sftp"
	"gossh/internal/ssh"
	"gossh/internal/ui"
)

const version = "2.0.0"

// Run starts the application
func Run() error {
	// Initialize config manager
	cfg, err := config.NewManager()
	if err != nil {
		return fmt.Errorf("failed to initialize config: %w", err)
	}

	// Create the app model
	appModel := ui.NewModel(cfg)

	// Create and run the Bubbletea program
	p := tea.NewProgram(appModel, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("failed to run program: %w", err)
	}

	return nil
}

// RunWithArgs runs the app with command line arguments
func RunWithArgs(args []string) error {
	if len(args) > 1 {
		switch args[1] {
		case "version", "-v", "--version":
			fmt.Printf("gossh v%s\n", version)
			return nil
		case "help", "-h", "--help":
			printHelp()
			return nil
		case "export":
			return runExport(args[2:])
		case "import":
			return runImport(args[2:])
		case "list":
			return runList()
		case "connect":
			if len(args) < 3 {
				return fmt.Errorf("usage: gossh connect <name>")
			}
			return runConnect(args[2])
		case "sftp":
			if len(args) < 3 {
				return fmt.Errorf("usage: gossh sftp <name>")
			}
			return runSFTP(args[2])
		case "forward":
			return runForward(args[2:])
		case "exec":
			return runExec(args[2:])
		}
	}

	return Run()
}

func printHelp() {
	help := `GoSSH - TUI SSH Connection Manager v%s

Usage:
  gossh                              Start the TUI application
  gossh help                         Show this help message
  gossh version                      Show version information
  gossh list                         List all connections
  gossh connect <name>               Connect to a server by name
  gossh export [file]                Export connections (default: connections.yaml)
  gossh import <file>                Import connections from file

Advanced Commands (v2.0):
  gossh sftp <name>                  Start SFTP session with a server
  gossh forward <name> -L/-R <spec>  Port forwarding (-L local, -R remote)
  gossh exec <command> [options]     Execute command on multiple servers
    --group=<group>                  Filter by group
    --tags=<tag1,tag2>               Filter by tags
    --names=<n1,n2>                  Filter by names
    --timeout=<seconds>              Command timeout (default: 30)

Examples:
  gossh sftp prod-web-01
  gossh forward prod-db -L 3306:localhost:3306
  gossh forward prod-web -R 8080:localhost:80
  gossh exec "uptime" --group=Production
  gossh exec "df -h" --tags=web,nginx

TUI Navigation:
  up/k               Move up
  down/j             Move down
  g/G                Jump to top/bottom
  /                  Search connections
  enter              Connect to selected server
  a                  Add new connection
  e                  Edit selected connection
  d                  Delete selected connection
  ?                  Show help
  q                  Quit

Config location:
  Linux/macOS: ~/.config/gossh/config.yaml
  Windows:     %%APPDATA%%\gossh\config.yaml
`
	fmt.Fprintf(os.Stdout, help, version)
}

// runExport exports connections to a file
func runExport(args []string) error {
	filename := "connections.yaml"
	if len(args) > 0 {
		filename = args[0]
	}

	cfg, err := config.NewManager()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if err := unlockIfNeeded(cfg); err != nil {
		return err
	}

	connections := cfg.Connections()

	// Create export structure (without encrypted fields)
	exportData := struct {
		Version     string             `yaml:"version"`
		Connections []model.Connection `yaml:"connections"`
	}{
		Version:     version,
		Connections: make([]model.Connection, len(connections)),
	}

	for i, conn := range connections {
		exportData.Connections[i] = conn
		exportData.Connections[i].EncryptedPassword = ""
		exportData.Connections[i].EncryptedKeyPassphrase = ""
	}

	data, err := yaml.Marshal(&exportData)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(filename, data, 0600); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	fmt.Printf("Exported %d connections to %s\n", len(connections), filename)
	return nil
}

// runImport imports connections from a file
func runImport(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: gossh import <file>")
	}

	filename := args[0]

	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var importData struct {
		Version     string             `yaml:"version"`
		Connections []model.Connection `yaml:"connections"`
	}

	if err := yaml.Unmarshal(data, &importData); err != nil {
		return fmt.Errorf("failed to parse file: %w", err)
	}

	cfg, err := config.NewManager()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if err := unlockIfNeeded(cfg); err != nil {
		return err
	}

	fmt.Print("Overwrite existing connections with same name? [y/N]: ")
	var answer string
	fmt.Scanln(&answer)
	overwrite := answer == "y" || answer == "Y"

	imported, err := cfg.ImportConnections(importData.Connections, overwrite)
	if err != nil {
		return fmt.Errorf("failed to import: %w", err)
	}

	fmt.Printf("Imported %d connections from %s\n", imported, filename)
	return nil
}

// runList lists all connections
func runList() error {
	cfg, err := config.NewManager()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if err := unlockIfNeeded(cfg); err != nil {
		return err
	}

	connections := cfg.Connections()

	if len(connections) == 0 {
		fmt.Println("No connections found.")
		return nil
	}

	fmt.Printf("%-20s %-30s %-10s %s\n", "NAME", "HOST", "PORT", "GROUP")
	fmt.Println("-------------------------------------------------------------------------------")
	for _, conn := range connections {
		group := conn.Group
		if group == "" {
			group = "Ungrouped"
		}
		fmt.Printf("%-20s %-30s %-10d %s\n", conn.Name, conn.User+"@"+conn.Host, conn.Port, group)
	}

	fmt.Printf("\nTotal: %d connections\n", len(connections))
	return nil
}

// runConnect connects to a server by name
func runConnect(name string) error {
	cfg, err := config.NewManager()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if err := unlockIfNeeded(cfg); err != nil {
		return err
	}

	conn := findConnection(cfg.Connections(), name)
	if conn == nil {
		return fmt.Errorf("connection '%s' not found", name)
	}

	fmt.Printf("Connecting to %s (%s@%s:%d)...\n", conn.Name, conn.User, conn.Host, conn.Port)

	terminal := ssh.NewTerminal(*conn)
	err = terminal.Run()

	if err != nil {
		cfg.UpdateConnectionStatus(conn.ID, model.ConnStatusFailed)
		return fmt.Errorf("connection failed: %w", err)
	}

	cfg.UpdateConnectionStatus(conn.ID, model.ConnStatusSuccess)
	return nil
}

// runSFTP starts an SFTP session
func runSFTP(name string) error {
	cfg, err := config.NewManager()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if err := unlockIfNeeded(cfg); err != nil {
		return err
	}

	conn := findConnection(cfg.Connections(), name)
	if conn == nil {
		return fmt.Errorf("connection '%s' not found", name)
	}

	fmt.Printf("Starting SFTP session to %s (%s@%s:%d)...\n", conn.Name, conn.User, conn.Host, conn.Port)

	client := sftp.NewClient(*conn)
	if err := client.Connect(); err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer client.Close()

	fmt.Println("Connected. Type 'help' for available commands.")

	// Simple SFTP shell
	scanner := bufio.NewScanner(os.Stdin)
	for {
		pwd, _ := client.Pwd()
		fmt.Printf("sftp:%s> ", pwd)

		if !scanner.Scan() {
			break
		}

		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		cmd := parts[0]
		args := parts[1:]

		switch cmd {
		case "help":
			fmt.Println("Commands:")
			fmt.Println("  ls [path]           List directory")
			fmt.Println("  cd <path>           Change directory")
			fmt.Println("  pwd                 Print working directory")
			fmt.Println("  get <remote> [local] Download file")
			fmt.Println("  put <local> [remote] Upload file")
			fmt.Println("  mkdir <path>        Create directory")
			fmt.Println("  rm <path>           Remove file")
			fmt.Println("  rmdir <path>        Remove directory")
			fmt.Println("  exit/quit           Exit SFTP")

		case "ls":
			path := "."
			if len(args) > 0 {
				path = args[0]
			}
			files, err := client.List(path)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				continue
			}
			for _, f := range files {
				fmt.Println(f.String())
			}

		case "cd":
			if len(args) == 0 {
				fmt.Println("Usage: cd <path>")
				continue
			}
			// Note: SFTP doesn't have cd, we'd need to track cwd ourselves
			fmt.Println("Note: cd not fully supported, use absolute paths")

		case "pwd":
			pwd, err := client.Pwd()
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				continue
			}
			fmt.Println(pwd)

		case "get":
			if len(args) == 0 {
				fmt.Println("Usage: get <remote> [local]")
				continue
			}
			remote := args[0]
			local := remote
			if len(args) > 1 {
				local = args[1]
			}
			if err := client.Download(remote, local); err != nil {
				fmt.Printf("Error: %v\n", err)
				continue
			}
			fmt.Printf("Downloaded %s -> %s\n", remote, local)

		case "put":
			if len(args) == 0 {
				fmt.Println("Usage: put <local> [remote]")
				continue
			}
			local := args[0]
			remote := local
			if len(args) > 1 {
				remote = args[1]
			}
			if err := client.Upload(local, remote); err != nil {
				fmt.Printf("Error: %v\n", err)
				continue
			}
			fmt.Printf("Uploaded %s -> %s\n", local, remote)

		case "mkdir":
			if len(args) == 0 {
				fmt.Println("Usage: mkdir <path>")
				continue
			}
			if err := client.Mkdir(args[0]); err != nil {
				fmt.Printf("Error: %v\n", err)
				continue
			}
			fmt.Printf("Created directory %s\n", args[0])

		case "rm":
			if len(args) == 0 {
				fmt.Println("Usage: rm <path>")
				continue
			}
			if err := client.Remove(args[0]); err != nil {
				fmt.Printf("Error: %v\n", err)
				continue
			}
			fmt.Printf("Removed %s\n", args[0])

		case "rmdir":
			if len(args) == 0 {
				fmt.Println("Usage: rmdir <path>")
				continue
			}
			if err := client.RemoveAll(args[0]); err != nil {
				fmt.Printf("Error: %v\n", err)
				continue
			}
			fmt.Printf("Removed directory %s\n", args[0])

		case "exit", "quit":
			fmt.Println("Goodbye!")
			return nil

		default:
			fmt.Printf("Unknown command: %s. Type 'help' for available commands.\n", cmd)
		}
	}

	return nil
}

// runForward starts port forwarding
func runForward(args []string) error {
	if len(args) < 3 {
		return fmt.Errorf("usage: gossh forward <name> -L/-R <spec>\nExample: gossh forward myserver -L 8080:localhost:80")
	}

	name := args[0]
	fwdFlag := args[1]
	spec := args[2]

	var fwdType ssh.ForwardType
	switch fwdFlag {
	case "-L":
		fwdType = ssh.ForwardLocal
	case "-R":
		fwdType = ssh.ForwardRemote
	default:
		return fmt.Errorf("invalid forward type: %s (use -L or -R)", fwdFlag)
	}

	cfg, err := config.NewManager()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if err := unlockIfNeeded(cfg); err != nil {
		return err
	}

	conn := findConnection(cfg.Connections(), name)
	if conn == nil {
		return fmt.Errorf("connection '%s' not found", name)
	}

	pf, err := ssh.ParsePortForward(fwdType, spec)
	if err != nil {
		return err
	}

	fmt.Printf("Setting up port forwarding to %s (%s@%s:%d)...\n",
		conn.Name, conn.User, conn.Host, conn.Port)

	forwarder := ssh.NewForwarder(*conn)
	forwarder.AddForward(pf)

	if err := forwarder.Connect(); err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	if err := forwarder.Start(); err != nil {
		forwarder.Stop()
		return fmt.Errorf("failed to start forwarding: %w", err)
	}

	fmt.Println("Port forwarding active. Press Ctrl+C to stop.")

	// Wait for interrupt
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	fmt.Println("\nStopping port forwarding...")
	forwarder.Stop()

	return nil
}

// runExec executes a command on multiple servers
func runExec(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: gossh exec <command> [--group=<group>] [--tags=<tags>] [--names=<names>]")
	}

	// Parse arguments
	var command string
	var group string
	var tags []string
	var names []string
	timeout := 30 * time.Second

	for _, arg := range args {
		if strings.HasPrefix(arg, "--group=") {
			group = strings.TrimPrefix(arg, "--group=")
		} else if strings.HasPrefix(arg, "--tags=") {
			tags = strings.Split(strings.TrimPrefix(arg, "--tags="), ",")
		} else if strings.HasPrefix(arg, "--names=") {
			names = strings.Split(strings.TrimPrefix(arg, "--names="), ",")
		} else if strings.HasPrefix(arg, "--timeout=") {
			var secs int
			fmt.Sscanf(strings.TrimPrefix(arg, "--timeout="), "%d", &secs)
			if secs > 0 {
				timeout = time.Duration(secs) * time.Second
			}
		} else if command == "" {
			command = arg
		}
	}

	if command == "" {
		return fmt.Errorf("no command specified")
	}

	cfg, err := config.NewManager()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if err := unlockIfNeeded(cfg); err != nil {
		return err
	}

	connections := cfg.Connections()

	// Filter connections
	if group != "" {
		connections = ssh.FilterByGroup(connections, group)
	}
	if len(tags) > 0 {
		connections = ssh.FilterByTags(connections, tags)
	}
	if len(names) > 0 {
		connections = ssh.FilterByNames(connections, names)
	}

	if len(connections) == 0 {
		return fmt.Errorf("no matching connections found")
	}

	fmt.Printf("Executing command on %d server(s):\n", len(connections))
	for _, c := range connections {
		fmt.Printf("  - %s (%s@%s)\n", c.Name, c.User, c.Host)
	}
	fmt.Printf("\nCommand: %s\n", command)
	fmt.Printf("Timeout: %v\n\n", timeout)

	// Confirm execution
	fmt.Print("Continue? [y/N]: ")
	var answer string
	fmt.Scanln(&answer)
	if answer != "y" && answer != "Y" {
		fmt.Println("Aborted.")
		return nil
	}

	// Execute
	ctx, cancel := context.WithTimeout(context.Background(), timeout*time.Duration(len(connections)))
	defer cancel()

	executor := ssh.NewBatchExecutor(connections)
	executor.SetTimeout(timeout)

	results := executor.Execute(ctx, command)
	ssh.PrintResults(results)

	return nil
}

// Helper functions

func unlockIfNeeded(cfg *config.Manager) error {
	if !cfg.IsFirstRun() && !cfg.IsUnlocked() {
		password, err := readPassword("Enter master password: ")
		if err != nil {
			return err
		}
		if err := cfg.Unlock(password); err != nil {
			return fmt.Errorf("failed to unlock: %w", err)
		}
	}
	return nil
}

// readPassword reads a password from stdin without echoing it
func readPassword(prompt string) (string, error) {
	fmt.Print(prompt)
	bytePassword, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		return "", fmt.Errorf("failed to read password: %w", err)
	}
	fmt.Println() // Print newline after password input
	return string(bytePassword), nil
}

func findConnection(connections []model.Connection, name string) *model.Connection {
	for i := range connections {
		if connections[i].Name == name {
			return &connections[i]
		}
	}
	return nil
}
