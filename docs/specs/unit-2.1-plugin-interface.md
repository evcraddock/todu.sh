# Unit 2.1: Plugin Interface Definition

**Status**: ✅ COMPLETE

**Goal**: Define the plugin interface for external task management systems

**Prerequisites**: Unit 1.5 complete

**Estimated time**: 20 minutes

---

## Requirements

### 1. Plugin Interface

Create `pkg/plugin/interface.go` with the Plugin interface:

```go
type Plugin interface {
    // Metadata
    Name() string
    Version() string

    // Configuration
    Configure(config map[string]string) error
    ValidateConfig() error

    // Projects
    FetchProjects(ctx context.Context) ([]*types.Project, error)
    FetchProject(ctx context.Context, externalID string) (*types.Project, error)

    // Tasks
    FetchTasks(ctx context.Context, projectExternalID *string, since *time.Time) ([]*types.Task, error)
    FetchTask(ctx context.Context, projectExternalID *string, taskExternalID string) (*types.Task, error)
    CreateTask(ctx context.Context, projectExternalID *string, task *types.TaskCreate) (*types.Task, error)
    UpdateTask(ctx context.Context, projectExternalID *string, taskExternalID string, task *types.TaskUpdate) (*types.Task, error)

    // Comments
    FetchComments(ctx context.Context, projectExternalID *string, taskExternalID string) ([]*types.Comment, error)
    CreateComment(ctx context.Context, projectExternalID *string, taskExternalID string, comment *types.CommentCreate) (*types.Comment, error)
}
```

### 2. Standard Errors

Define standard errors in `pkg/plugin/errors.go`:

- `ErrNotSupported` - For optional operations not supported by the plugin
- `ErrNotConfigured` - For missing or invalid configuration
- `ErrNotFound` - For resources that don't exist
- `ErrUnauthorized` - For authentication/authorization failures

Each error must include helpful context about what's needed.

### 3. Documentation

Add package documentation explaining:

- Purpose of the plugin interface
- How plugins should implement each method
- What external_id represents for projects and tasks
- How to handle optional operations
- Configuration expectations

### 4. Testing Support

Create `pkg/plugin/mock.go` with a mock plugin for testing:

- Implements all Plugin interface methods
- Stores data in memory
- Returns configurable errors for testing
- Can be used by tests in other packages

---

## Success Criteria

- ✅ Plugin interface is well-defined
- ✅ All methods have clear documentation
- ✅ Standard errors are defined
- ✅ Package documentation explains usage
- ✅ Mock plugin implementation available
- ✅ Interface compiles without errors

---

## Verification

- `go build ./pkg/plugin` - must compile
- Interface can be imported by other packages
- Mock plugin implements full interface

---

## Commit Message

```text
feat: define plugin interface
```
