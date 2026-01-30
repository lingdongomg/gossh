# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is an **OpenSpec project** - a spec-driven development framework for managing change proposals and specifications. Despite the repository name "gossh", this contains no Go source code; it's a scaffolding project for the OpenSpec methodology.

## OpenSpec Workflow

### Three Stages

1. **Proposal** (`/proposal`) - Create change proposals in `openspec/changes/<change-id>/`
2. **Apply** (`/apply`) - Implement approved changes
3. **Archive** (`/archive`) - Move completed changes to `changes/archive/` and update specs

### Key Commands

```bash
openspec list                    # List active changes
openspec list --specs            # List specifications
openspec show <item>             # Display change or spec
openspec validate <id> --strict --no-interactive  # Validate changes
openspec archive <id> --yes      # Archive after deployment
```

### Directory Structure

```
openspec/
├── project.md              # Project conventions
├── specs/                  # Current truth - what IS built
│   └── <capability>/spec.md
├── changes/                # Proposals - what SHOULD change
│   ├── <change-id>/
│   │   ├── proposal.md     # Why, what, impact
│   │   ├── tasks.md        # Implementation checklist
│   │   ├── design.md       # Technical decisions (optional)
│   │   └── specs/<capability>/spec.md  # Delta changes
│   └── archive/            # Completed changes
```

## Critical Formatting Rules

### Scenario Headers (Must Use)

```markdown
#### Scenario: User login success
- **WHEN** valid credentials provided
- **THEN** return JWT token
```

### Delta Operations

Use these headers in change spec files:
- `## ADDED Requirements` - New capabilities
- `## MODIFIED Requirements` - Changed behavior (include full requirement text)
- `## REMOVED Requirements` - Deprecated features
- `## RENAMED Requirements` - Name changes

Every requirement MUST have at least one `#### Scenario:`.

## Naming Conventions

- **Capability names**: verb-noun format (`user-auth`, `payment-capture`)
- **Change IDs**: kebab-case, verb-led (`add-two-factor-auth`, `update-config`)

## When to Create Proposals

**Create proposal for:**
- New features or functionality
- Breaking changes (API, schema)
- Architecture or pattern changes
- Security pattern updates

**Skip proposal for:**
- Bug fixes restoring intended behavior
- Typos, formatting, comments
- Non-breaking dependency updates
- Configuration changes
