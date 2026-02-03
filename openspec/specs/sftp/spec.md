# sftp Specification

## Purpose
TBD - created by archiving change enhance-gossh-v1.2. Update Purpose after archive.
## Requirements
### Requirement: Working Directory Tracking
The system SHALL track the current working directory during SFTP sessions to support proper `cd` command functionality.

#### Scenario: Changing directory
- **WHEN** user executes `cd <path>` command
- **THEN** the system verifies the path exists and is a directory
- **AND** updates the internal current directory state
- **AND** subsequent commands use this directory as base

#### Scenario: Displaying current directory
- **WHEN** user executes `pwd` command
- **THEN** the system displays the tracked current directory

#### Scenario: Relative path resolution
- **WHEN** user provides a relative path in any SFTP command
- **THEN** the system resolves it against the current working directory

### Requirement: Transfer Progress Display
The system SHALL display transfer progress during file upload and download operations.

#### Scenario: Large file transfer
- **WHEN** transferring a file larger than 1MB
- **THEN** the system displays a progress indicator
- **AND** shows transferred bytes and percentage
- **AND** updates progress at least every second

#### Scenario: Small file transfer
- **WHEN** transferring a file smaller than 1MB
- **THEN** the system may skip progress display for performance
- **AND** still reports completion status

#### Scenario: Transfer cancellation
- **WHEN** user interrupts a transfer in progress
- **THEN** the system terminates the transfer gracefully
- **AND** cleans up any partial files

