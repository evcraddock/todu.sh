# Todu Work Unit Specifications

This directory contains individual specification files for each unit of work. Each unit is designed to be:

- **Small**: 15-30 minutes of work
- **Compilable**: Results in working code
- **Testable**: Includes verification steps
- **Committable**: Has a clear commit message

## Unit Index

### Phase 0: Bootstrap (15-30 min total)

- [Unit 0.1: Initialize Git Repository and GitHub Project](./unit-0.1-init-git.md) (5 min)
- [Unit 0.2: Initialize Go Module](./unit-0.2-init-go-module.md) (10 min)
- [Unit 0.3: Add Version Command with Cobra](./unit-0.3-add-version-command.md) (15 min)

### Phase 1: Core Foundation (70-90 min total)

- [Unit 1.1: Define Core Types](./unit-1.1-define-core-types.md) (20 min)
- [Unit 1.2: Configuration Loading](./unit-1.2-config-loading.md) (20 min)
- [Unit 1.3: Config Show Command](./unit-1.3-config-show-command.md) (15 min)
- [Unit 1.4: Basic API Client Structure](./unit-1.4-api-client-structure.md) (15 min)

### Phase 2: Plugin System (TBD)

Additional units will be added as implementation progresses.

## How to Use

1. Read the spec for the current unit
2. Implement the code as specified
3. Run the verification steps
4. Commit with the provided commit message
5. Move to the next unit

## Naming Convention

- `unit-X.Y-short-description.md`
- X = Phase number
- Y = Unit number within phase
- short-description = Brief description with hyphens

## Dependencies

Each unit clearly states its prerequisites. Complete units in order unless you have a good reason not to.
