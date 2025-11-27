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

// ProjectListOptions contains optional filters for listing projects
type ProjectListOptions struct {
	SystemID *int
	Priority []string
}

// ListProjects retrieves all projects, optionally filtered
func (c *Client) ListProjects(ctx context.Context, opts *ProjectListOptions) ([]*types.Project, error) {
	path := "/api/v1/projects/?"

	if opts != nil {
		if opts.SystemID != nil {
			path += fmt.Sprintf("system_id=%d&", *opts.SystemID)
		}
		if len(opts.Priority) > 0 {
			for _, p := range opts.Priority {
				path += fmt.Sprintf("priority=%s&", p)
			}
		}
	}

	// Remove trailing & or ?
	if path[len(path)-1] == '&' || path[len(path)-1] == '?' {
		path = path[:len(path)-1]
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
	ProjectID       *int
	Status          string
	Priority        string
	ProjectStatus   []string
	ProjectPriority []string
	Limit           int
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
		if len(opts.ProjectStatus) > 0 {
			for _, ps := range opts.ProjectStatus {
				path += fmt.Sprintf("project_status=%s&", ps)
			}
		}
		if len(opts.ProjectPriority) > 0 {
			for _, pp := range opts.ProjectPriority {
				path += fmt.Sprintf("project_priority=%s&", pp)
			}
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

// CreateComment creates a new comment or journal entry
func (c *Client) CreateComment(ctx context.Context, comment *types.CommentCreate) (*types.Comment, error) {
	var path string
	if comment.TaskID != nil {
		// Task comment: use task-specific endpoint
		path = fmt.Sprintf("/api/v1/tasks/%d/comments", *comment.TaskID)
	} else {
		// Journal entry: use general comments endpoint
		path = "/api/v1/comments"
	}

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
	resp, err := c.doRequest(ctx, http.MethodPatch, path, comment)
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

// ListJournals retrieves all journal entries (comments without task_id)
func (c *Client) ListJournals(ctx context.Context, skip, limit int) ([]*types.Comment, error) {
	path := fmt.Sprintf("/api/v1/comments?type=journal&skip=%d&limit=%d", skip, limit)
	resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var journals []*types.Comment
	if err := parseResponse(resp, &journals); err != nil {
		return nil, err
	}

	return journals, nil
}

// ListAllComments retrieves all comments and journal entries with optional type filter
func (c *Client) ListAllComments(ctx context.Context, commentType string, skip, limit int) ([]*types.Comment, error) {
	path := fmt.Sprintf("/api/v1/comments?type=%s&skip=%d&limit=%d", commentType, skip, limit)
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

// Recurring Task Template Methods

// TemplateListOptions contains optional filters for listing recurring task templates
type TemplateListOptions struct {
	ProjectID    *int
	Active       *bool
	TemplateType string
	Skip         int
	Limit        int
}

// ListTemplates retrieves recurring task templates with optional filters
func (c *Client) ListTemplates(ctx context.Context, opts *TemplateListOptions) ([]*types.RecurringTaskTemplate, error) {
	path := "/api/v1/recurring-templates/?"

	if opts != nil {
		if opts.ProjectID != nil {
			path += fmt.Sprintf("project_id=%d&", *opts.ProjectID)
		}
		if opts.Active != nil {
			path += fmt.Sprintf("is_active=%t&", *opts.Active)
		}
		if opts.TemplateType != "" {
			path += fmt.Sprintf("template_type=%s&", opts.TemplateType)
		}
		if opts.Skip > 0 {
			path += fmt.Sprintf("skip=%d&", opts.Skip)
		}
		if opts.Limit > 0 {
			path += fmt.Sprintf("limit=%d&", opts.Limit)
		} else {
			path += "limit=100&"
		}
	} else {
		path += "limit=100&"
	}

	// Remove trailing & or ?
	path = path[:len(path)-1]

	resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var templates []*types.RecurringTaskTemplate
	if err := parseResponse(resp, &templates); err != nil {
		return nil, err
	}

	return templates, nil
}

// GetTemplate retrieves a specific recurring task template by ID
func (c *Client) GetTemplate(ctx context.Context, id int) (*types.RecurringTaskTemplate, error) {
	path := fmt.Sprintf("/api/v1/recurring-templates/%d", id)
	resp, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	var template types.RecurringTaskTemplate
	if err := parseResponse(resp, &template); err != nil {
		return nil, err
	}

	return &template, nil
}

// CreateTemplate creates a new recurring task template
func (c *Client) CreateTemplate(ctx context.Context, template *types.RecurringTaskTemplateCreate) (*types.RecurringTaskTemplate, error) {
	resp, err := c.doRequest(ctx, http.MethodPost, "/api/v1/recurring-templates/", template)
	if err != nil {
		return nil, err
	}

	var created types.RecurringTaskTemplate
	if err := parseResponse(resp, &created); err != nil {
		return nil, err
	}

	return &created, nil
}

// UpdateTemplate updates an existing recurring task template
func (c *Client) UpdateTemplate(ctx context.Context, id int, template *types.RecurringTaskTemplateUpdate) (*types.RecurringTaskTemplate, error) {
	path := fmt.Sprintf("/api/v1/recurring-templates/%d", id)
	resp, err := c.doRequest(ctx, http.MethodPatch, path, template)
	if err != nil {
		return nil, err
	}

	var updated types.RecurringTaskTemplate
	if err := parseResponse(resp, &updated); err != nil {
		return nil, err
	}

	return &updated, nil
}

// DeleteTemplate deletes a recurring task template
func (c *Client) DeleteTemplate(ctx context.Context, id int) error {
	path := fmt.Sprintf("/api/v1/recurring-templates/%d", id)
	resp, err := c.doRequest(ctx, http.MethodDelete, path, nil)
	if err != nil {
		return err
	}

	return parseResponse(resp, nil)
}

// TemplateProcessDetail represents details about a single template processing result
type TemplateProcessDetail struct {
	TemplateID int    `json:"template_id"`
	Action     string `json:"action"`   // "created", "skipped", or "failed"
	TaskID     *int   `json:"task_id"`  // only if action == "created"
	Reason     string `json:"reason"`   // only if action == "skipped"
	Error      string `json:"error"`    // only if action == "failed"
}

// ProcessDueTemplatesResponse represents the response from processing due templates
type ProcessDueTemplatesResponse struct {
	Processed    int                     `json:"processed"`
	TasksCreated int                     `json:"tasks_created"`
	Skipped      int                     `json:"skipped"`
	Failed       int                     `json:"failed"`
	Details      []TemplateProcessDetail `json:"details"`
}

// ProcessDueTemplates processes all recurring task templates that are due
func (c *Client) ProcessDueTemplates(ctx context.Context) (*ProcessDueTemplatesResponse, error) {
	path := "/api/v1/recurring-templates/process-due"
	resp, err := c.doRequest(ctx, http.MethodPost, path, nil)
	if err != nil {
		return nil, err
	}

	var result ProcessDueTemplatesResponse
	if err := parseResponse(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}
