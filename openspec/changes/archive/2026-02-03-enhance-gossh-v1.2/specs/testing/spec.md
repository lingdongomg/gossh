# Testing Spec Delta

## ADDED Requirements

### Requirement: Unit Test Coverage
The system SHALL have unit tests covering critical modules with a minimum target coverage of 60%.

#### Scenario: Crypto module tests
- **WHEN** running `go test ./internal/crypto/...`
- **THEN** tests verify encryption/decryption correctness
- **AND** tests verify password hashing and validation
- **AND** tests verify machine ID retrieval

#### Scenario: Config module tests
- **WHEN** running `go test ./internal/config/...`
- **THEN** tests verify configuration file read/write
- **AND** tests verify connection CRUD operations
- **AND** tests verify thread-safe access

#### Scenario: Model module tests
- **WHEN** running `go test ./internal/model/...`
- **THEN** tests verify connection validation
- **AND** tests verify filter matching

#### Scenario: SSH config parser tests
- **WHEN** running `go test ./internal/sshconfig/...`
- **THEN** tests verify parsing of valid SSH config files
- **AND** tests verify handling of invalid or missing files

#### Scenario: HostKey module tests
- **WHEN** running `go test ./internal/ssh/...`
- **THEN** tests verify known_hosts file operations
- **AND** tests verify fingerprint formatting

#### Scenario: i18n module tests
- **WHEN** running `go test ./internal/i18n/...`
- **THEN** tests verify translation retrieval
- **AND** tests verify fallback behavior

### Requirement: Test Infrastructure
The system SHALL provide convenient test execution via Makefile.

#### Scenario: Running all tests
- **WHEN** user runs `make test`
- **THEN** the system executes all unit tests
- **AND** reports pass/fail status

#### Scenario: Coverage report
- **WHEN** user runs `make coverage`
- **THEN** the system generates coverage report
- **AND** outputs HTML report to `coverage.html`

#### Scenario: Verbose test output
- **WHEN** user runs `make test-verbose`
- **THEN** the system shows detailed test output
