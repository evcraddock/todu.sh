package forgejo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Forgejo API response types

// Repository represents a Forgejo repository.
type Repository struct {
	ID          int64     `json:"id"`
	Owner       *User     `json:"owner"`
	Name        string    `json:"name"`
	FullName    string    `json:"full_name"`
	Description string    `json:"description"`
	HTMLURL     string    `json:"html_url"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// User represents a Forgejo user.
type User struct {
	ID       int64  `json:"id"`
	Login    string `json:"login"`
	FullName string `json:"full_name"`
}

// Issue represents a Forgejo issue.
type Issue struct {
	ID          int64      `json:"id"`
	Number      int        `json:"number"`
	Title       string     `json:"title"`
	Body        string     `json:"body"`
	State       string     `json:"state"`
	StateReason string     `json:"state_reason"`
	HTMLURL     string     `json:"html_url"`
	Labels      []*Label   `json:"labels"`
	Assignees   []*User    `json:"assignees"`
	PullRequest *struct{}  `json:"pull_request"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	ClosedAt    *time.Time `json:"closed_at"`
}

// Label represents a Forgejo label.
type Label struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
}

// Comment represents a Forgejo issue comment.
type Comment struct {
	ID        int64     `json:"id"`
	Body      string    `json:"body"`
	User      *User     `json:"user"`
	HTMLURL   string    `json:"html_url"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// API request types

// CreateIssueRequest represents the request body for creating an issue.
type CreateIssueRequest struct {
	Title    string  `json:"title"`
	Body     string  `json:"body,omitempty"`
	Labels   []int64 `json:"labels,omitempty"`
	Assignee string  `json:"assignee,omitempty"`
}

// UpdateIssueRequest represents the request body for updating an issue.
type UpdateIssueRequest struct {
	Title *string `json:"title,omitempty"`
	Body  *string `json:"body,omitempty"`
	State *string `json:"state,omitempty"`
}

// UpdateLabelsRequest represents the request body for updating issue labels.
type UpdateLabelsRequest struct {
	Labels []int64 `json:"labels"`
}

// CreateCommentRequest represents the request body for creating a comment.
type CreateCommentRequest struct {
	Body string `json:"body"`
}

// CreateLabelRequest represents the request body for creating a label.
type CreateLabelRequest struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

// client wraps the Forgejo API with HTTP client and label caching.
type client struct {
	baseURL    string
	token      string
	httpClient *http.Client

	// Label cache: map[owner/repo]map[labelName]labelID
	labelCache map[string]map[string]int64
	labelMu    sync.RWMutex
}

// newClient creates a new Forgejo API client.
func newClient(config map[string]string) (*client, error) {
	token := strings.TrimSpace(config["token"])
	if token == "" {
		return nil, fmt.Errorf("token is required")
	}

	baseURL := strings.TrimSpace(config["url"])
	if baseURL == "" {
		return nil, fmt.Errorf("url is required")
	}

	// Normalize base URL (remove trailing slash)
	baseURL = strings.TrimSuffix(baseURL, "/")

	return &client{
		baseURL:    baseURL,
		token:      token,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		labelCache: make(map[string]map[string]int64),
	}, nil
}

// doRequest performs an HTTP request with authentication.
func (c *client) doRequest(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
	fullURL := c.baseURL + "/api/v1" + path

	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	req, err := http.NewRequestWithContext(ctx, method, fullURL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "token "+c.token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	// Check for errors
	if resp.StatusCode >= 400 {
		defer resp.Body.Close()
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return resp, nil
}

// listRepositories retrieves all repositories accessible to the authenticated user.
func (c *client) listRepositories(ctx context.Context) ([]*Repository, error) {
	var allRepos []*Repository
	page := 1
	limit := 100

	for {
		path := fmt.Sprintf("/user/repos?page=%d&limit=%d", page, limit)
		resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
		if err != nil {
			return nil, err
		}

		var repos []*Repository
		if err := json.NewDecoder(resp.Body).Decode(&repos); err != nil {
			resp.Body.Close()
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}
		resp.Body.Close()

		allRepos = append(allRepos, repos...)

		if len(repos) < limit {
			break
		}
		page++
	}

	return allRepos, nil
}

// getRepository retrieves a single repository by owner and name.
func (c *client) getRepository(ctx context.Context, owner, repo string) (*Repository, error) {
	path := fmt.Sprintf("/repos/%s/%s", owner, repo)
	resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var repository Repository
	if err := json.NewDecoder(resp.Body).Decode(&repository); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &repository, nil
}

// listIssues retrieves issues for a repository.
func (c *client) listIssues(ctx context.Context, owner, repo string, since *time.Time) ([]*Issue, error) {
	var allIssues []*Issue
	page := 1
	limit := 100

	for {
		path := fmt.Sprintf("/repos/%s/%s/issues?state=all&page=%d&limit=%d", owner, repo, page, limit)
		if since != nil {
			path += "&since=" + url.QueryEscape(since.Format(time.RFC3339))
		}

		resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
		if err != nil {
			return nil, err
		}

		var issues []*Issue
		if err := json.NewDecoder(resp.Body).Decode(&issues); err != nil {
			resp.Body.Close()
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}
		resp.Body.Close()

		// Filter out pull requests (they come back in the issues endpoint)
		for _, issue := range issues {
			if issue.PullRequest == nil {
				allIssues = append(allIssues, issue)
			}
		}

		if len(issues) < limit {
			break
		}
		page++
	}

	return allIssues, nil
}

// getIssue retrieves a single issue by number.
func (c *client) getIssue(ctx context.Context, owner, repo string, number int) (*Issue, error) {
	path := fmt.Sprintf("/repos/%s/%s/issues/%d", owner, repo, number)
	resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var issue Issue
	if err := json.NewDecoder(resp.Body).Decode(&issue); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &issue, nil
}

// createIssue creates a new issue in a repository.
func (c *client) createIssue(ctx context.Context, owner, repo string, req *CreateIssueRequest) (*Issue, error) {
	path := fmt.Sprintf("/repos/%s/%s/issues", owner, repo)
	resp, err := c.doRequest(ctx, http.MethodPost, path, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var issue Issue
	if err := json.NewDecoder(resp.Body).Decode(&issue); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &issue, nil
}

// updateIssue updates an existing issue.
func (c *client) updateIssue(ctx context.Context, owner, repo string, number int, req *UpdateIssueRequest) (*Issue, error) {
	path := fmt.Sprintf("/repos/%s/%s/issues/%d", owner, repo, number)
	resp, err := c.doRequest(ctx, http.MethodPatch, path, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var issue Issue
	if err := json.NewDecoder(resp.Body).Decode(&issue); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &issue, nil
}

// updateIssueLabels replaces all labels on an issue.
func (c *client) updateIssueLabels(ctx context.Context, owner, repo string, number int, labelIDs []int64) (*Issue, error) {
	path := fmt.Sprintf("/repos/%s/%s/issues/%d/labels", owner, repo, number)
	req := &UpdateLabelsRequest{Labels: labelIDs}

	resp, err := c.doRequest(ctx, http.MethodPut, path, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// The labels endpoint returns labels, not the full issue
	// We need to fetch the issue again to return consistent data
	return c.getIssue(ctx, owner, repo, number)
}

// listComments retrieves all comments for an issue.
func (c *client) listComments(ctx context.Context, owner, repo string, number int) ([]*Comment, error) {
	var allComments []*Comment
	page := 1
	limit := 100

	for {
		path := fmt.Sprintf("/repos/%s/%s/issues/%d/comments?page=%d&limit=%d", owner, repo, number, page, limit)
		resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
		if err != nil {
			return nil, err
		}

		var comments []*Comment
		if err := json.NewDecoder(resp.Body).Decode(&comments); err != nil {
			resp.Body.Close()
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}
		resp.Body.Close()

		allComments = append(allComments, comments...)

		if len(comments) < limit {
			break
		}
		page++
	}

	return allComments, nil
}

// createComment creates a new comment on an issue.
func (c *client) createComment(ctx context.Context, owner, repo string, number int, body string) (*Comment, error) {
	path := fmt.Sprintf("/repos/%s/%s/issues/%d/comments", owner, repo, number)
	req := &CreateCommentRequest{Body: body}

	resp, err := c.doRequest(ctx, http.MethodPost, path, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var comment Comment
	if err := json.NewDecoder(resp.Body).Decode(&comment); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &comment, nil
}

// Label management methods

// listLabels retrieves all labels for a repository.
func (c *client) listLabels(ctx context.Context, owner, repo string) ([]*Label, error) {
	var allLabels []*Label
	page := 1
	limit := 100

	for {
		path := fmt.Sprintf("/repos/%s/%s/labels?page=%d&limit=%d", owner, repo, page, limit)
		resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
		if err != nil {
			return nil, err
		}

		var labels []*Label
		if err := json.NewDecoder(resp.Body).Decode(&labels); err != nil {
			resp.Body.Close()
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}
		resp.Body.Close()

		allLabels = append(allLabels, labels...)

		if len(labels) < limit {
			break
		}
		page++
	}

	return allLabels, nil
}

// createLabel creates a new label in a repository.
func (c *client) createLabel(ctx context.Context, owner, repo, name string) (*Label, error) {
	path := fmt.Sprintf("/repos/%s/%s/labels", owner, repo)
	req := &CreateLabelRequest{
		Name:  name,
		Color: generateLabelColor(name),
	}

	resp, err := c.doRequest(ctx, http.MethodPost, path, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var label Label
	if err := json.NewDecoder(resp.Body).Decode(&label); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &label, nil
}

// resolveLabelIDs converts label names to IDs, creating labels if they don't exist.
func (c *client) resolveLabelIDs(ctx context.Context, owner, repo string, labelNames []string) ([]int64, error) {
	if len(labelNames) == 0 {
		return []int64{}, nil
	}

	repoKey := owner + "/" + repo

	// Ensure cache is populated
	if err := c.populateLabelCache(ctx, owner, repo); err != nil {
		return nil, err
	}

	c.labelMu.RLock()
	cache := c.labelCache[repoKey]
	c.labelMu.RUnlock()

	labelIDs := make([]int64, 0, len(labelNames))

	for _, name := range labelNames {
		if id, ok := cache[name]; ok {
			labelIDs = append(labelIDs, id)
		} else {
			// Create the label
			label, err := c.createLabel(ctx, owner, repo, name)
			if err != nil {
				return nil, fmt.Errorf("failed to create label %q: %w", name, err)
			}

			// Update cache
			c.labelMu.Lock()
			if c.labelCache[repoKey] == nil {
				c.labelCache[repoKey] = make(map[string]int64)
			}
			c.labelCache[repoKey][name] = label.ID
			c.labelMu.Unlock()

			labelIDs = append(labelIDs, label.ID)
		}
	}

	return labelIDs, nil
}

// populateLabelCache fetches all labels for a repository and caches them.
func (c *client) populateLabelCache(ctx context.Context, owner, repo string) error {
	repoKey := owner + "/" + repo

	c.labelMu.RLock()
	_, exists := c.labelCache[repoKey]
	c.labelMu.RUnlock()

	if exists {
		return nil
	}

	labels, err := c.listLabels(ctx, owner, repo)
	if err != nil {
		return err
	}

	c.labelMu.Lock()
	c.labelCache[repoKey] = make(map[string]int64)
	for _, label := range labels {
		c.labelCache[repoKey][label.Name] = label.ID
	}
	c.labelMu.Unlock()

	return nil
}

// generateLabelColor generates a color for a label based on its name.
func generateLabelColor(name string) string {
	// Use a simple hash to generate consistent colors
	var hash uint32
	for _, c := range name {
		hash = hash*31 + uint32(c)
	}

	// Generate a pastel color
	r := 128 + (hash&0xFF)%128
	g := 128 + ((hash>>8)&0xFF)%128
	b := 128 + ((hash>>16)&0xFF)%128

	return fmt.Sprintf("%02x%02x%02x", r, g, b)
}

// Helper to convert int64 to string
func int64ToString(i int64) string {
	return strconv.FormatInt(i, 10)
}
