# Unit 1.5: API Client Methods

**Status**: ✅ COMPLETE

**Goal**: Implement all API client methods for Todu API endpoints

**Prerequisites**: Unit 1.4 complete

**Estimated time**: 30 minutes

---

## Requirements

### 1. System Methods

Implement methods in `internal/api/client.go` for system management:

- `ListSystems(ctx context.Context) ([]*types.System, error)`
- `GetSystem(ctx context.Context, id int) (*types.System, error)`
- `CreateSystem(ctx context.Context, system *types.SystemCreate) (*types.System, error)`
- `UpdateSystem(ctx context.Context, id int, system *types.SystemUpdate) (*types.System, error)`
- `DeleteSystem(ctx context.Context, id int) error`

### 2. Project Methods

Implement methods for project management:

- `ListProjects(ctx context.Context, systemID *int) ([]*types.Project, error)`
- `GetProject(ctx context.Context, id int) (*types.Project, error)`
- `CreateProject(ctx context.Context, project *types.ProjectCreate) (*types.Project, error)`
- `UpdateProject(ctx context.Context, id int, project *types.ProjectUpdate) (*types.Project, error)`
- `DeleteProject(ctx context.Context, id int) error`

### 3. Task Methods

Implement methods for task management:

- `ListTasks(ctx context.Context, projectID *int) ([]*types.Task, error)`
- `GetTask(ctx context.Context, id int) (*types.Task, error)`
- `CreateTask(ctx context.Context, task *types.TaskCreate) (*types.Task, error)`
- `UpdateTask(ctx context.Context, id int, task *types.TaskUpdate) (*types.Task, error)`
- `DeleteTask(ctx context.Context, id int) error`

### 4. Comment Methods

Implement methods for comment management:

- `ListComments(ctx context.Context, taskID int) ([]*types.Comment, error)`
- `GetComment(ctx context.Context, id int) (*types.Comment, error)`
- `CreateComment(ctx context.Context, comment *types.CommentCreate) (*types.Comment, error)`
- `DeleteComment(ctx context.Context, id int) error`

### 5. Error Handling

- Return clear errors for HTTP failures
- Parse API error responses if available
- Handle network timeouts appropriately
- Include request context in errors

### 6. URL Construction

Each method must:

- Use `doRequest` helper from Unit 1.4
- Construct proper API paths (e.g., `/api/systems`, `/api/projects/{id}`)
- Pass request body for CREATE/UPDATE operations
- Use `parseResponse` to decode responses

### 7. Testing

Create `internal/api/methods_test.go` with:

- Mock HTTP server tests for each method
- Test successful responses
- Test error responses (404, 500, etc.)
- Test request body marshaling
- Test response parsing

---

## Success Criteria

- ✅ All API methods implemented
- ✅ Methods use correct HTTP verbs (GET/POST/PUT/DELETE)
- ✅ Request bodies properly marshaled to JSON
- ✅ Responses properly parsed from JSON
- ✅ Tests pass: `go test ./internal/api`
- ✅ Error handling is consistent and clear
- ✅ No external API calls during tests (use mock server)

---

## Verification

- `go test ./internal/api -v` - all tests must pass
- Each CRUD operation has corresponding test
- Methods can be called with proper context

---

## Commit Message

```text
feat: implement all API client methods
```
