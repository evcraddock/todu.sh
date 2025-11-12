# Unit 1.1: Define Core Types

**Goal**: Create shared type definitions matching Todu API schema

**Prerequisites**: Unit 0.3 complete

**Estimated time**: 20 minutes

---

## Requirements

### 1. Type Definitions

Create Go type definitions in `pkg/types/` that exactly match the Todu API schema.

### 2. Task Types

Define types in `pkg/types/task.go`:

- `Task` - Full task representation with all fields
- `TaskCreate` - Data for creating a new task
- `TaskUpdate` - Data for updating an existing task

Must include fields:

- ID, ExternalID, SourceURL
- Title, Description
- ProjectID, Status, Priority, DueDate
- CreatedAt, UpdatedAt
- Labels, Assignees

### 3. Project Types

Define types in `pkg/types/project.go`:

- `Project` - Full project representation
- `ProjectCreate` - Data for creating a project
- `ProjectUpdate` - Data for updating a project

Must include fields:

- ID, Name, Description
- SystemID, ExternalID, Status
- CreatedAt, UpdatedAt

### 4. System Types

Define types in `pkg/types/system.go`:

- `System` - External task management system
- `SystemCreate` - Data for creating a system
- `SystemUpdate` - Data for updating a system

Must include fields:

- ID, Identifier, Name, URL
- Metadata (map of strings)
- CreatedAt, UpdatedAt

### 5. Supporting Types

Define in separate files:

- `Label` - Task label (id, name)
- `Assignee` - Task assignee (id, name)
- `Comment` - Task comment with full fields
- `CommentCreate` - Data for creating comments

### 6. JSON Serialization

- All types must have appropriate JSON struct tags
- Use `omitempty` for optional fields
- Use pointers for nullable fields
- Follow Go naming conventions while mapping to API snake_case

### 7. Testing

Create `pkg/types/types_test.go` with:

- Tests for JSON marshaling/unmarshaling
- Tests for at least Task and Project types
- Verify JSON tag correctness

---

## Success Criteria

- ✅ All type files compile without errors
- ✅ Types exactly match Todu API schema
- ✅ JSON tags are correctly applied
- ✅ Tests pass: `go test ./pkg/types`
- ✅ Can successfully serialize and deserialize types

---

## Verification

- `go build ./pkg/types` - must succeed
- `go test ./pkg/types -v` - all tests pass
- Types can be imported by other packages

---

## Commit Message

```text
feat: add core type definitions
```
