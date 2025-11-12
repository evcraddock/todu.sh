# Unit 1.4: Basic API Client Structure

**Goal**: Create HTTP client wrapper for Todu API

**Prerequisites**: Unit 1.3 complete

**Estimated time**: 15 minutes

---

## Requirements

### 1. API Client Package

Create `internal/api/client.go` with basic HTTP client infrastructure.

### 2. Client Structure

Implement `Client` struct that:

- Stores the base URL for the API
- Contains an HTTP client with appropriate timeout (30 seconds)
- Is created via `NewClient(baseURL string)` constructor
- Can be used for making HTTP requests

### 3. Request Method

Implement private `doRequest` method that:

- Takes context, HTTP method, path, and optional body
- Constructs full URL from base URL + path
- Marshals request body to JSON if provided
- Sets appropriate headers:
  - `Content-Type: application/json` (when body present)
  - `Accept: application/json`
- Creates HTTP request with context
- Executes the request
- Returns HTTP response and error

### 4. Response Parsing

Implement private `parseResponse` function that:

- Takes HTTP response and destination interface
- Checks for HTTP error status codes (>=400)
- Returns descriptive error for non-success responses
- Decodes JSON response body into destination
- Properly closes response body
- Handles nil destination (for DELETE requests, etc.)

### 5. Error Handling

- All errors must be wrapped with context
- HTTP errors must include status code
- JSON parsing errors must be clear
- Network errors must be propagated

### 6. Testing

Create `internal/api/client_test.go` with:

- Test that NewClient creates a client
- Test that base URL is stored correctly
- Test that HTTP client is initialized
- Test multiple base URL scenarios

---

## Success Criteria

- ✅ API client package compiles
- ✅ Can create new client with any base URL
- ✅ HTTP client has 30-second timeout
- ✅ Tests pass: `go test ./internal/api`
- ✅ Methods are properly encapsulated (private helpers)
- ✅ Error messages are clear and helpful

---

## Verification

- `go test ./internal/api -v` - all tests must pass
- Client can be imported by other packages
- No external API calls made during tests

---

## Commit Message

```text
feat: add basic API client structure
```
