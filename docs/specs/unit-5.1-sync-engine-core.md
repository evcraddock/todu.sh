# Unit 5.1: Sync Engine Core

**Status**: 🔲 TODO

**Goal**: Implement bidirectional sync engine

**Prerequisites**: Unit 4.1 complete

**Estimated time**: 60 minutes

---

## Requirements

### 1. Sync Engine Structure

Create `internal/sync/engine.go` with:

```go
type Engine struct {
    apiClient *api.Client
    registry  *registry.Registry
}

func NewEngine(apiClient *api.Client, registry *registry.Registry) *Engine
```

### 2. Sync Strategy

Define sync strategies in `internal/sync/strategy.go`:

```go
type Strategy string

const (
    StrategyPull           Strategy = "pull"  // External → Todu only
    StrategyPush           Strategy = "push"  // Todu → External only
    StrategyBidirectional  Strategy = "bidirectional"
)
```

### 3. Sync Options

Create `internal/sync/options.go`:

```go
type Options struct {
    ProjectIDs []int
    SystemID   *int
    Strategy   Strategy
    DryRun     bool
}
```

### 4. Sync Algorithm

Implement `Sync(ctx context.Context, options Options) (*Result, error)`:

**For each project:**

1. Get project details from Todu API
2. Get system for project
3. Create plugin instance for system
4. Get last sync time from project metadata
5. Fetch tasks from plugin (since last sync if pull/bidirectional)
6. Fetch tasks from Todu API for this project
7. Sync tasks based on strategy:

**Pull Strategy:**
- For each external task:
  - Find in Todu by external_id and project_id
  - If not found: Create in Todu
  - If found and external is newer: Update in Todu

**Push Strategy:**
- For each Todu task with external_id:
  - Fetch from external system
  - If Todu is newer: Push update to external

**Bidirectional:**
- Run both pull and push
- Use last-write-wins for conflicts (compare updated_at)

8. Update last sync time in project metadata
9. Collect and return sync results

### 5. Conflict Resolution

Implement in `internal/sync/conflict.go`:

- `ResolveConflict(toduTask, externalTask *types.Task) *types.Task`
- Use last-write-wins based on updated_at
- Log conflicts for user awareness

### 6. Sync Result

Create `internal/sync/result.go`:

```go
type Result struct {
    ProjectResults []ProjectResult
    TotalCreated   int
    TotalUpdated   int
    TotalSkipped   int
    TotalErrors    int
    Duration       time.Duration
}

type ProjectResult struct {
    ProjectID    int
    ProjectName  string
    Created      int
    Updated      int
    Skipped      int
    Errors       []error
}
```

### 7. Progress Reporting

Implement progress tracking:
- Log progress to stdout
- Show current project being synced
- Show tasks processed/total
- Show sync results summary

### 8. Error Handling

- Continue syncing other projects if one fails
- Collect all errors in results
- Don't crash on plugin errors
- Log errors with context (project, task)

### 9. Dry Run Mode

When `DryRun` is true:
- Don't make any changes
- Show what would be created/updated
- Return predicted results

### 10. Testing

Create `internal/sync/engine_test.go`:

- Test pull strategy
- Test push strategy
- Test bidirectional strategy
- Test conflict resolution
- Test dry run mode
- Test error handling
- Use mock plugin and mock API client

---

## Success Criteria

- ✅ Sync engine implements all strategies
- ✅ Can sync tasks from external to Todu
- ✅ Can push tasks from Todu to external
- ✅ Handles conflicts correctly
- ✅ Reports progress and results
- ✅ Tracks last sync time
- ✅ Dry run mode works
- ✅ Tests pass: `go test ./internal/sync`

---

## Verification

- `go test ./internal/sync -v` - all tests pass
- Sync works with GitHub plugin
- Conflicts are resolved correctly

---

## Commit Message

```text
feat: implement sync engine core
```
