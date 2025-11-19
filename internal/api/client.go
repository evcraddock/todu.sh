package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/evcraddock/todu.sh/pkg/types"
)

// Client is an HTTP client for the Todu API
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a new API client with the given base URL
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// doRequest executes an HTTP request to the API
func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
	url := c.baseURL + path

	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	return resp, nil
}

// parseResponse parses an HTTP response into the destination interface
func parseResponse(resp *http.Response, dest interface{}) error {
	defer resp.Body.Close()

	// Check for HTTP error status codes
	if resp.StatusCode >= 400 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Handle nil destination (e.g., for DELETE requests)
	if dest == nil {
		return nil
	}

	// Decode JSON response
	if err := json.NewDecoder(resp.Body).Decode(dest); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	return nil
}

// System Methods

// ListSystems retrieves all systems
func (c *Client) ListSystems(ctx context.Context) ([]*types.System, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, "/api/v1/systems/", nil)
	if err != nil {
		return nil, err
	}

	var systems []*types.System
	if err := parseResponse(resp, &systems); err != nil {
		return nil, err
	}

	return systems, nil
}

// GetSystem retrieves a specific system by ID
func (c *Client) GetSystem(ctx context.Context, id int) (*types.System, error) {
	path := fmt.Sprintf("/api/v1/systems/%d", id)
	resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var system types.System
	if err := parseResponse(resp, &system); err != nil {
		return nil, err
	}

	return &system, nil
}

// CreateSystem creates a new system
func (c *Client) CreateSystem(ctx context.Context, system *types.SystemCreate) (*types.System, error) {
	resp, err := c.doRequest(ctx, http.MethodPost, "/api/v1/systems/", system)
	if err != nil {
		return nil, err
	}

	var created types.System
	if err := parseResponse(resp, &created); err != nil {
		return nil, err
	}

	return &created, nil
}

// UpdateSystem updates an existing system
func (c *Client) UpdateSystem(ctx context.Context, id int, system *types.SystemUpdate) (*types.System, error) {
	path := fmt.Sprintf("/api/v1/systems/%d", id)
	resp, err := c.doRequest(ctx, http.MethodPut, path, system)
	if err != nil {
		return nil, err
	}

	var updated types.System
	if err := parseResponse(resp, &updated); err != nil {
		return nil, err
	}

	return &updated, nil
}

// DeleteSystem deletes a system
func (c *Client) DeleteSystem(ctx context.Context, id int) error {
	path := fmt.Sprintf("/api/v1/systems/%d", id)
	resp, err := c.doRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return err
	}

	return parseResponse(resp, nil)
}

// Project Methods

// ListProjects retrieves all projects, optionally filtered by system ID
func (c *Client) ListProjects(ctx context.Context, systemID *int) ([]*types.Project, error) {
	path := "/api/v1/projects/"
	if systemID != nil {
		path = fmt.Sprintf("/api/v1/projects/?system_id=%d", *systemID)
	}

	resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var projects []*types.Project
	if err := parseResponse(resp, &projects); err != nil {
		return nil, err
	}

	return projects, nil
}

// GetProject retrieves a specific project by ID
func (c *Client) GetProject(ctx context.Context, id int) (*types.Project, error) {
	path := fmt.Sprintf("/api/v1/projects/%d", id)
	resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var project types.Project
	if err := parseResponse(resp, &project); err != nil {
		return nil, err
	}

	return &project, nil
}

// CreateProject creates a new project
func (c *Client) CreateProject(ctx context.Context, project *types.ProjectCreate) (*types.Project, error) {
	resp, err := c.doRequest(ctx, http.MethodPost, "/api/v1/projects/", project)
	if err != nil {
		return nil, err
	}

	var created types.Project
	if err := parseResponse(resp, &created); err != nil {
		return nil, err
	}

	return &created, nil
}

// UpdateProject updates an existing project
func (c *Client) UpdateProject(ctx context.Context, id int, project *types.ProjectUpdate) (*types.Project, error) {
	path := fmt.Sprintf("/api/v1/projects/%d", id)
	resp, err := c.doRequest(ctx, http.MethodPut, path, project)
	if err != nil {
		return nil, err
	}

	var updated types.Project
	if err := parseResponse(resp, &updated); err != nil {
		return nil, err
	}

	return &updated, nil
}

// DeleteProject deletes a project
func (c *Client) DeleteProject(ctx context.Context, id int, cascade bool) error {
	path := fmt.Sprintf("/api/v1/projects/%d", id)
	if cascade {
		path += "?cascade=true"
	}
	resp, err := c.doRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return err
	}

	return parseResponse(resp, nil)
}

// Task Methods

// TasksResponse represents the paginated tasks response from the API
type TasksResponse struct {
	Items []*types.Task `json:"items"`
	Total int           `json:"total"`
	Skip  int           `json:"skip"`
	Limit int           `json:"limit"`
}

// TaskListOptions contains optional filters for listing tasks
type TaskListOptions struct {
	ProjectID *int
	Status    string
	Priority  string
	Limit     int
}

// ListTasks retrieves tasks with optional filters
func (c *Client) ListTasks(ctx context.Context, opts *TaskListOptions) ([]*types.Task, error) {
	path := "/api/v1/tasks/?"

	if opts != nil {
		if opts.ProjectID != nil {
			path += fmt.Sprintf("project_id=%d&", *opts.ProjectID)
		}
		if opts.Status != "" {
			path += fmt.Sprintf("status=%s&", opts.Status)
		}
		if opts.Priority != "" {
			path += fmt.Sprintf("priority=%s&", opts.Priority)
		}
		if opts.Limit > 0 {
			path += fmt.Sprintf("limit=%d&", opts.Limit)
		} else {
			path += "limit=500&"
		}
	} else {
		path += "limit=500&"
	}

	// Remove trailing & or ?
	path = path[:len(path)-1]

	resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var tasksResp TasksResponse
	if err := parseResponse(resp, &tasksResp); err != nil {
		return nil, err
	}

	return tasksResp.Items, nil
}

// GetTask retrieves a specific task by ID
func (c *Client) GetTask(ctx context.Context, id int) (*types.Task, error) {
	path := fmt.Sprintf("/api/v1/tasks/%d", id)
	resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var task types.Task
	if err := parseResponse(resp, &task); err != nil {
		return nil, err
	}

	return &task, nil
}

// CreateTask creates a new task
func (c *Client) CreateTask(ctx context.Context, task *types.TaskCreate) (*types.Task, error) {
	resp, err := c.doRequest(ctx, http.MethodPost, "/api/v1/tasks/", task)
	if err != nil {
		return nil, err
	}

	var created types.Task
	if err := parseResponse(resp, &created); err != nil {
		return nil, err
	}

	return &created, nil
}

// UpdateTask updates an existing task
func (c *Client) UpdateTask(ctx context.Context, id int, task *types.TaskUpdate) (*types.Task, error) {
	path := fmt.Sprintf("/api/v1/tasks/%d", id)
	resp, err := c.doRequest(ctx, http.MethodPut, path, task)
	if err != nil {
		return nil, err
	}

	var updated types.Task
	if err := parseResponse(resp, &updated); err != nil {
		return nil, err
	}

	return &updated, nil
}

// DeleteTask deletes a task
func (c *Client) DeleteTask(ctx context.Context, id int) error {
	path := fmt.Sprintf("/api/v1/tasks/%d", id)
	resp, err := c.doRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return err
	}

	return parseResponse(resp, nil)
}

// Comment Methods

// ListComments retrieves all comments for a task
func (c *Client) ListComments(ctx context.Context, taskID int) ([]*types.Comment, error) {
	path := fmt.Sprintf("/api/v1/tasks/%d/comments", taskID)
	resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var comments []*types.Comment
	if err := parseResponse(resp, &comments); err != nil {
		return nil, err
	}

	return comments, nil
}

// GetComment retrieves a specific comment by ID
func (c *Client) GetComment(ctx context.Context, id int) (*types.Comment, error) {
	path := fmt.Sprintf("/api/v1/comments/%d", id)
	resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var comment types.Comment
	if err := parseResponse(resp, &comment); err != nil {
		return nil, err
	}

	return &comment, nil
}

// CreateComment creates a new comment
func (c *Client) CreateComment(ctx context.Context, comment *types.CommentCreate) (*types.Comment, error) {
	path := fmt.Sprintf("/api/v1/tasks/%d/comments", comment.TaskID)
	resp, err := c.doRequest(ctx, http.MethodPost, path, comment)
	if err != nil {
		return nil, err
	}

	var created types.Comment
	if err := parseResponse(resp, &created); err != nil {
		return nil, err
	}

	return &created, nil
}

// UpdateComment updates an existing comment
func (c *Client) UpdateComment(ctx context.Context, id int, comment *types.CommentUpdate) (*types.Comment, error) {
	path := fmt.Sprintf("/api/v1/comments/%d", id)
	resp, err := c.doRequest(ctx, http.MethodPut, path, comment)
	if err != nil {
		return nil, err
	}

	var updated types.Comment
	if err := parseResponse(resp, &updated); err != nil {
		return nil, err
	}

	return &updated, nil
}

// DeleteComment deletes a comment
func (c *Client) DeleteComment(ctx context.Context, id int) error {
	path := fmt.Sprintf("/api/v1/comments/%d", id)
	resp, err := c.doRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return err
	}

	return parseResponse(resp, nil)
}
