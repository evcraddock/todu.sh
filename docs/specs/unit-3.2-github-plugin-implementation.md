# Unit 3.2: GitHub Plugin Implementation

**Status**: ✅ COMPLETE

**Goal**: Implement all GitHub plugin interface methods

**Prerequisites**: Unit 3.1 complete

**Estimated time**: 45 minutes

---

## Requirements

### 1. GitHub Client Wrapper

Create `plugins/github/client.go`:

- Wrap `github.com/google/go-github/v56/github` client
- Handle authentication with token
- Provide methods for:
  - Listing repositories
  - Getting repository details
  - Listing issues
  - Getting issue details
  - Creating issues
  - Updating issues
  - Listing comments
  - Creating comments

### 2. Type Mappings

Create `plugins/github/mapper.go`:

**Repository → Project:**
- external_id = "owner/repo" (e.g., "octocat/hello-world")
- name = repository name
- description = repository description
- status = "active" (repositories don't have status)

**Issue → Task:**
- external_id = issue number as string (e.g., "123")
- title = issue title
- description = issue body
- status mapping:
  - "open" → "open"
  - "closed" → "done"
- priority from labels:
  - "priority:high" → "high"
  - "priority:medium" → "medium"
  - "priority:low" → "low"
  - no priority label → null
- labels = issue labels (excluding priority labels)
- assignees = issue assignees
- source_url = issue HTML URL
- created_at/updated_at = issue timestamps
- due_date = milestone due date (if present)

**Comment → Comment:**
- content = comment body
- author = comment author login
- created_at/updated_at = comment timestamps

### 3. Implement FetchProjects

- List all repositories accessible to the authenticated user
- Convert each repository to Project
- Handle pagination
- Return empty list if no repositories

### 4. Implement FetchProject

- Get repository by external_id ("owner/repo")
- Parse external_id to extract owner and repo
- Return error if repository not found
- Convert to Project

### 5. Implement FetchTasks

- List issues for repository specified by projectExternalID
- If `since` provided, filter issues updated after that time
- Convert each issue to Task
- Handle pagination
- Return empty list if no issues

### 6. Implement FetchTask

- Get issue by projectExternalID and taskExternalID
- Parse projectExternalID to get owner/repo
- Parse taskExternalID to get issue number
- Return error if issue not found
- Convert to Task

### 7. Implement CreateTask

- Create new issue in repository
- Convert Task to GitHub issue format
- Set title and body
- Set labels (including priority label if specified)
- Set assignees if specified
- Return created task

### 8. Implement UpdateTask

- Update existing issue
- Support updating:
  - Title
  - Description (body)
  - Status (state: open/closed)
  - Priority (via labels)
  - Labels
  - Assignees
- Return updated task

### 9. Implement FetchComments

- List comments for issue
- Convert each comment to Comment
- Handle pagination
- Return empty list if no comments

### 10. Implement CreateComment

- Create comment on issue
- Set comment body
- Return created comment

### 11. Error Handling

- Return `plugin.ErrNotFound` for 404 responses
- Return `plugin.ErrUnauthorized` for 401/403 responses
- Return descriptive errors for other failures
- Include resource context in errors (repo, issue number)

### 12. Testing

Create `plugins/github/plugin_test.go`:

- Test type mappings (mapper_test.go)
- Mock GitHub API responses
- Test each plugin method
- Test error cases
- Test pagination handling

---

## Success Criteria

- ✅ All Plugin interface methods implemented
- ✅ Type mappings work correctly
- ✅ Can list repositories
- ✅ Can fetch issues
- ✅ Can create/update issues
- ✅ Can fetch/create comments
- ✅ Error handling is consistent
- ✅ Tests pass: `go test ./plugins/github`
- ✅ Integration test works with real GitHub token

---

## Verification

- `go test ./plugins/github -v` - all tests pass
- Manual test with real repository
- Can sync issues to/from GitHub

---

## Commit Message

```text
feat: implement GitHub plugin methods
```
