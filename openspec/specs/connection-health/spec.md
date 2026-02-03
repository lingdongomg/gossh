# connection-health Specification

## Purpose
TBD - created by archiving change enhance-gossh-v1.2. Update Purpose after archive.
## Requirements
### Requirement: Quick Health Check
The system SHALL provide a quick TCP-based health check to verify host reachability.

#### Scenario: Host reachable
- **WHEN** user requests quick health check for a connection
- **AND** TCP connection to host:port succeeds within timeout
- **THEN** the system reports the connection as reachable
- **AND** records the response time

#### Scenario: Host unreachable
- **WHEN** user requests quick health check for a connection
- **AND** TCP connection fails or times out
- **THEN** the system reports the connection as unreachable
- **AND** provides the error reason

### Requirement: Full Health Check
The system SHALL provide a comprehensive SSH-based health check to verify authentication capability.

#### Scenario: Full check success
- **WHEN** user requests full health check for a connection
- **AND** SSH handshake completes successfully
- **THEN** the system reports the connection as healthy
- **AND** closes the connection immediately after verification

#### Scenario: Authentication failure
- **WHEN** user requests full health check for a connection
- **AND** SSH handshake fails due to authentication error
- **THEN** the system reports authentication failure
- **AND** suggests checking credentials

### Requirement: Status Indicator Display
The system SHALL display connection status indicators in the connection list view.

#### Scenario: Displaying status
- **WHEN** viewing the connection list
- **THEN** each connection shows a status indicator
- **AND** uses ✓ for reachable, ✗ for unreachable, ? for unknown

#### Scenario: Manual status refresh
- **WHEN** user presses `t` key on a selected connection
- **THEN** the system performs a quick health check
- **AND** updates the status indicator

### Requirement: Batch Health Check
The system SHALL support checking multiple connections simultaneously via CLI.

#### Scenario: Batch check via CLI
- **WHEN** user runs `gossh check [--all|--group NAME]`
- **THEN** the system checks all matching connections in parallel
- **AND** displays a summary of results

