# settings-ui Specification

## Purpose
TBD - created by archiving change enhance-gossh-v1.2. Update Purpose after archive.
## Requirements
### Requirement: Settings View
The system SHALL provide a dedicated settings view accessible from the main interface.

#### Scenario: Accessing settings
- **WHEN** user presses `s` key from the connection list
- **THEN** the system displays the settings view
- **AND** shows current configuration options

#### Scenario: Exiting settings
- **WHEN** user presses `Escape` or `q` in settings view
- **THEN** the system returns to the connection list
- **AND** discards unsaved changes

### Requirement: Language Selection
The system SHALL allow users to switch between supported languages.

#### Scenario: Changing language
- **WHEN** user selects a different language in settings
- **AND** saves the settings
- **THEN** the UI immediately updates to the selected language
- **AND** the preference persists across sessions

#### Scenario: Supported languages
- **WHEN** viewing language options
- **THEN** the system offers English and Chinese (中文)

### Requirement: Password Protection Management
The system SHALL allow users to enable, disable, or change master password from settings.

#### Scenario: Enabling password protection
- **WHEN** user enables password protection in settings
- **THEN** the system prompts for new master password
- **AND** requires confirmation
- **AND** re-encrypts all sensitive data with the new key

#### Scenario: Disabling password protection
- **WHEN** user disables password protection in settings
- **AND** password protection is currently enabled
- **THEN** the system prompts for current password verification
- **AND** re-encrypts data with machine-derived key

#### Scenario: Changing master password
- **WHEN** user selects change password option
- **THEN** the system prompts for current password
- **AND** prompts for new password with confirmation
- **AND** re-encrypts all sensitive data

### Requirement: Internationalization Support
The system SHALL support multiple languages throughout the user interface.

#### Scenario: Text translation
- **WHEN** the UI displays any text
- **THEN** it retrieves the text from the i18n module
- **AND** uses the currently selected language

#### Scenario: Translation fallback
- **WHEN** a translation key is missing for the current language
- **THEN** the system falls back to English
- **AND** logs a warning for developers

#### Scenario: CLI output language
- **WHEN** running CLI commands
- **THEN** output messages use the configured language preference

### Requirement: Version Display
The system SHALL display the current version consistently throughout the application.

#### Scenario: Help view version
- **WHEN** viewing the help page
- **THEN** the system displays the build version (e.g., v1.2.0)

#### Scenario: CLI version flag
- **WHEN** user runs `gossh --version` or `gossh version`
- **THEN** the system prints the build version

#### Scenario: Settings version info
- **WHEN** viewing settings page
- **THEN** the About section displays the current version

