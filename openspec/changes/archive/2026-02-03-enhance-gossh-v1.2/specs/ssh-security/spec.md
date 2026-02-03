# SSH Security Spec Delta

## ADDED Requirements

### Requirement: SSH HostKey Verification
The system SHALL verify SSH host keys against a known_hosts file before establishing connections to prevent man-in-the-middle attacks.

#### Scenario: First connection to unknown host
- **WHEN** user connects to a host not in known_hosts
- **THEN** the system displays the host's fingerprint
- **AND** prompts user to confirm trust decision
- **AND** saves the key to known_hosts upon confirmation

#### Scenario: Connection to known host with matching key
- **WHEN** user connects to a host in known_hosts
- **AND** the host key matches the stored key
- **THEN** the connection proceeds without prompting

#### Scenario: Connection to host with changed key
- **WHEN** user connects to a host in known_hosts
- **AND** the host key differs from the stored key
- **THEN** the system displays a security warning
- **AND** requires explicit user confirmation to proceed
- **AND** provides option to update the stored key

#### Scenario: Known hosts file location
- **WHEN** the system needs to access known_hosts
- **THEN** it uses `~/.config/gossh/known_hosts` on Linux/macOS
- **AND** `%APPDATA%\gossh\known_hosts` on Windows

### Requirement: SSH Connection Factory
The system SHALL provide a unified SSH connection factory to eliminate duplicate connection logic across modules.

#### Scenario: Creating SSH connection
- **WHEN** any module needs an SSH connection
- **THEN** it uses the centralized `Connect()` factory function
- **AND** provides connection options including host, port, user, auth method, and timeout

#### Scenario: Connection with custom host key callback
- **WHEN** creating a connection with custom host key verification
- **THEN** the factory accepts a `HostKeyCallback` parameter
- **AND** uses the provided callback during SSH handshake

### Requirement: StartupCommand Execution
The system SHALL execute a configured startup command after successfully establishing an SSH terminal session.

#### Scenario: Connection with startup command
- **WHEN** user connects to a host with `StartupCommand` configured
- **AND** the SSH session is established
- **THEN** the system sends the startup command to the remote shell
- **AND** waits for command completion before interactive mode

#### Scenario: Connection without startup command
- **WHEN** user connects to a host without `StartupCommand` configured
- **THEN** the system enters interactive mode immediately

#### Scenario: Startup command timeout
- **WHEN** startup command execution exceeds timeout
- **THEN** the system logs a warning
- **AND** proceeds to interactive mode
