# Todu Work Unit Specifications

This directory contains individual specification files for each unit of work. Each unit is designed to be:

- **Small**: 15-30 minutes of work
- **Compilable**: Results in working code
- **Testable**: Includes verification steps
- **Committable**: Has a clear commit message

## Unit Index

### Phase 0: Bootstrap (30 min total)

- ✅ [Unit 0.1: Initialize Git Repository and GitHub Project](./unit-0.1-init-git.md) (5 min)
- ✅ [Unit 0.2: Initialize Go Module](./unit-0.2-init-go-module.md) (10 min)
- ✅ [Unit 0.3: Add Version Command with Cobra](./unit-0.3-add-version-command.md) (15 min)

### Phase 1: Core Foundation (125 min total)

- ✅ [Unit 1.1: Define Core Types](./unit-1.1-define-core-types.md) (20 min)
- ✅ [Unit 1.2: Configuration Loading](./unit-1.2-config-loading.md) (20 min)
- ✅ [Unit 1.3: Config Show Command](./unit-1.3-config-show-command.md) (15 min)
- ✅ [Unit 1.4: Basic API Client Structure](./unit-1.4-api-client-structure.md) (15 min)
- ✅ [Unit 1.5: API Client Methods](./unit-1.5-api-client-methods.md) (30 min)

### Phase 2: Plugin Interface and Registry (75 min total)

- ✅ [Unit 2.1: Plugin Interface Definition](./unit-2.1-plugin-interface.md) (20 min)
- ✅ [Unit 2.2: Plugin Registry](./unit-2.2-plugin-registry.md) (25 min)
- ✅ [Unit 2.3: System Management Commands](./unit-2.3-system-commands.md) (30 min)

### Phase 3: GitHub Plugin (60 min total)

- ✅ [Unit 3.1: GitHub Plugin Structure](./unit-3.1-github-plugin-structure.md) (15 min)
- ✅ [Unit 3.2: GitHub Plugin Implementation](./unit-3.2-github-plugin-implementation.md) (45 min)

### Phase 4: Project Management (40 min total)

- ✅ [Unit 4.1: Project Management Commands](./unit-4.1-project-commands.md) (40 min)

### Phase 5: Sync Engine (90 min total)

- 🔲 [Unit 5.1: Sync Engine Core](./unit-5.1-sync-engine-core.md) (60 min)
- 🔲 [Unit 5.2: Sync Commands](./unit-5.2-sync-commands.md) (30 min)

### Phase 6: Task Management (50 min total)

- 🔲 [Unit 6.1: Task Management Commands](./unit-6.1-task-commands.md) (50 min)

### Phase 7: Daemon Mode (85 min total)

- 🔲 [Unit 7.1: Daemon Mode Implementation](./unit-7.1-daemon-mode.md) (45 min)
- 🔲 [Unit 7.2: Daemon Commands and Service Management](./unit-7.2-daemon-commands.md) (40 min)

### Phase 8: Polish and Documentation (60 min total)

- 🔲 [Unit 8.1: Polish and Documentation](./unit-8.1-polish-and-documentation.md) (60 min)

## Total Estimated Time

- ✅ Completed: ~305 minutes (~5.08 hours)
- 🔲 Remaining: ~310 minutes (~5.17 hours)
- **Total**: ~615 minutes (~10 hours)

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
