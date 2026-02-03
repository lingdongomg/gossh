# ssh-config-import Specification

## Purpose
TBD - created by archiving change enhance-gossh-v1.2. Update Purpose after archive.
## Requirements
### Requirement: SSH Config File Parsing
The system SHALL parse OpenSSH configuration files and import connections.

#### Scenario: Basic import
- **WHEN** user runs `gossh import --ssh-config [path]`
- **THEN** the system parses the SSH config file (default `~/.ssh/config`)
- **AND** creates connections for each Host entry

#### Scenario: Supported directives
- **WHEN** parsing SSH config
- **THEN** the system extracts the following directives:
  - `Host` → connection name
  - `HostName` → server address
  - `User` → username
  - `Port` → port number (default 22)
  - `IdentityFile` → private key path

#### Scenario: Wildcard host patterns
- **WHEN** encountering Host with wildcards (e.g., `Host *`)
- **THEN** the system skips the entry
- **AND** logs an informational message

#### Scenario: Missing required fields
- **WHEN** a Host entry lacks HostName
- **AND** Host is not a valid hostname itself
- **THEN** the system skips the entry
- **AND** reports the reason

### Requirement: Import Conflict Resolution
The system SHALL handle conflicts when imported connections match existing ones.

#### Scenario: Duplicate connection name
- **WHEN** imported connection has same name as existing
- **THEN** the system prompts user to skip, rename, or overwrite

#### Scenario: Interactive mode
- **WHEN** running import in interactive mode
- **THEN** the system shows preview of connections to import
- **AND** allows user to select which to import

#### Scenario: Non-interactive mode
- **WHEN** running import with `--yes` flag
- **THEN** the system skips duplicates automatically
- **AND** reports skipped entries in summary

