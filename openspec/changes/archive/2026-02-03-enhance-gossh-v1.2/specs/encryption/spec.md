# Encryption Spec Delta

## ADDED Requirements

### Requirement: Machine-Based Key Derivation
The system SHALL derive encryption keys from machine-specific characteristics when password protection is disabled, instead of using fixed strings.

#### Scenario: Key derivation on Linux
- **WHEN** the system needs encryption key in no-password mode on Linux
- **THEN** it combines hostname, username, and `/etc/machine-id` content
- **AND** derives the key using Argon2id with stored salt

#### Scenario: Key derivation on macOS
- **WHEN** the system needs encryption key in no-password mode on macOS
- **THEN** it combines hostname, username, and IOPlatformUUID
- **AND** derives the key using Argon2id with stored salt

#### Scenario: Key derivation on Windows
- **WHEN** the system needs encryption key in no-password mode on Windows
- **THEN** it combines hostname, username, and MachineGuid from registry
- **AND** derives the key using Argon2id with stored salt

#### Scenario: Machine ID unavailable
- **WHEN** machine ID cannot be retrieved
- **THEN** the system falls back to hostname and username combination
- **AND** logs a warning about reduced security

### Requirement: Legacy Configuration Migration
The system SHALL detect and handle configuration files encrypted with the legacy fixed-key method.

#### Scenario: Detecting legacy encryption
- **WHEN** loading a configuration file
- **AND** the encryption uses the legacy fixed-key format
- **THEN** the system attempts decryption with the legacy key
- **AND** prompts user to re-encrypt with the new method

#### Scenario: Successful migration
- **WHEN** user confirms configuration migration
- **THEN** the system decrypts with the legacy key
- **AND** re-encrypts with machine-derived key
- **AND** saves the updated configuration
