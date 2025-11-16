package todoist

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	defaultBaseURL = "https://api.todoist.com/rest/v2"
	userAgent      = "todu.sh-todoist-plugin/1.0.0"
)

// Todoist API response types

// Project represents a Todoist project.
type Project struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Color          string `json:"color"`
	ParentID       string `json:"parent_id,omitempty"`
	Order          int    `json:"order"`
	CommentCount   int    `json:"comment_count"`
	IsShared       bool   `json:"is_shared"`
	IsFavorite     bool   `json:"is_favorite"`
	IsInboxProject bool   `json:"is_inbox_project"`
	IsTeamInbox    bool   `json:"is_team_inbox"`
	ViewStyle      string `json:"view_style"`
	URL            string `json:"url"`
}

// Due represents a due date in Todoist.
type Due struct {
	Date        string `json:"date"`
	String      string `json:"string"`
	Lang        string `json:"lang"`
	IsRecurring bool   `json:"is_recurring"`
	Datetime    string `json:"datetime,omitempty"`
	Timezone    string `json:"timezone,omitempty"`
}

// Task represents a Todoist task.
type Task struct {
	ID           string   `json:"id"`
	ProjectID    string   `json:"project_id"`
	SectionID    string   `json:"section_id,omitempty"`
	Content      string   `json:"content"`
	Description  string   `json:"description"`
	IsCompleted  bool     `json:"is_completed"`
	Labels       []string `json:"labels"`
	Priority     int      `json:"priority"`
	CommentCount int      `json:"comment_count"`
	CreatorID    string   `json:"creator_id"`
	CreatedAt    string   `json:"created_at"`
	Due          *Due     `json:"due,omitempty"`
	URL          string   `json:"url"`
	Duration     *struct {
		Amount int    `json:"amount"`
		Unit   string `json:"unit"`
	} `json:"duration,omitempty"`
}

// Comment represents a Todoist comment.
type Comment struct {
	ID         string `json:"id"`
	TaskID     string `json:"task_id,omitempty"`
	ProjectID  string `json:"project_id,omitempty"`
	PostedAt   string `json:"posted_at"`
	Content    string `json:"content"`
	Attachment *struct {
		FileName    string `json:"file_name"`
		FileType    string `json:"file_type"`
		FileURL     string `json:"file_url"`
		ResourceType string `json:"resource_type"`
	} `json:"attachment,omitempty"`
}

// Request types

// TaskCreateRequest represents a request to create a task.
type TaskCreateRequest struct {
	Content     string   `json:"content"`
	Description string   `json:"description,omitempty"`
	ProjectID   string   `json:"project_id,omitempty"`
	SectionID   string   `json:"section_id,omitempty"`
	ParentID    string   `json:"parent_id,omitempty"`
	Order       int      `json:"order,omitempty"`
	Labels      []string `json:"labels,omitempty"`
	Priority    int      `json:"priority,omitempty"`
	DueString   string   `json:"due_string,omitempty"`
	DueDate     string   `json:"due_date,omitempty"`
	DueDatetime string   `json:"due_datetime,omitempty"`
	DueLang     string   `json:"due_lang,omitempty"`
	Assignee    string   `json:"assignee_id,omitempty"`
}

// TaskUpdateRequest represents a request to update a task.
type TaskUpdateRequest struct {
	Content     *string  `json:"content,omitempty"`
	Description *string  `json:"description,omitempty"`
	Labels      []string `json:"labels,omitempty"`
	Priority    *int     `json:"priority,omitempty"`
	DueString   *string  `json:"due_string,omitempty"`
	DueDate     *string  `json:"due_date,omitempty"`
	DueDatetime *string  `json:"due_datetime,omitempty"`
	Assignee    *string  `json:"assignee_id,omitempty"`
}

// CommentCreateRequest represents a request to create a comment.
type CommentCreateRequest struct {
	TaskID    string `json:"task_id,omitempty"`
	ProjectID string `json:"project_id,omitempty"`
	Content   string `json:"content"`
}

// client wraps the Todoist REST API.
type client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

// newClient creates a new Todoist API client.
func newClient(config map[string]string) (*client, error) {
	token := strings.TrimSpace(config["token"])
	if token == "" {
		return nil, fmt.Errorf("token is required")
	}

	baseURL := defaultBaseURL
	if url := config["url"]; url != "" {
		baseURL = strings.TrimSpace(url)
	}

	return &client{
		baseURL: baseURL,
		token:   token,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

// doRequest performs an HTTP request with authentication.
func (c *client) doRequest(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
	url := c.baseURL + path

	var bodyReader io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonData)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("User-Agent", userAgent)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	return resp, nil
}

// decodeResponse decodes a JSON response body.
func decodeResponse(resp *http.Response, v interface{}) error {
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	if v == nil {
		return nil
	}

	return json.NewDecoder(resp.Body).Decode(v)
}

// Project operations

// listProjects retrieves all projects.
func (c *client) listProjects(ctx context.Context) ([]*Project, error) {
	resp, err := c.doRequest(ctx, "GET", "/projects", nil)
	if err != nil {
		return nil, err
	}

	var projects []*Project
	if err := decodeResponse(resp, &projects); err != nil {
		return nil, err
	}

	return projects, nil
}

// getProject retrieves a single project by ID.
func (c *client) getProject(ctx context.Context, id string) (*Project, error) {
	resp, err := c.doRequest(ctx, "GET", "/projects/"+id, nil)
	if err != nil {
		return nil, err
	}

	var project Project
	if err := decodeResponse(resp, &project); err != nil {
		return nil, err
	}

	return &project, nil
}

// Task operations

// listTasks retrieves tasks with optional filters.
func (c *client) listTasks(ctx context.Context, projectID *string, since *time.Time) ([]*Task, error) {
	path := "/tasks"
	params := url.Values{}

	if projectID != nil && *projectID != "" {
		params.Set("project_id", *projectID)
	}

	// Note: Todoist REST API doesn't support "since" filtering directly
	// We fetch all tasks and filter client-side if needed
	// The sync API supports this but REST doesn't

	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	resp, err := c.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var tasks []*Task
	if err := decodeResponse(resp, &tasks); err != nil {
		return nil, err
	}

	// Client-side filtering for "since" if provided
	// Note: Todoist doesn't provide updated_at in REST API v2
	// so we can't filter by modification time
	// This is a limitation of the Todoist REST API

	return tasks, nil
}

// getTask retrieves a single task by ID.
func (c *client) getTask(ctx context.Context, id string) (*Task, error) {
	resp, err := c.doRequest(ctx, "GET", "/tasks/"+id, nil)
	if err != nil {
		return nil, err
	}

	var task Task
	if err := decodeResponse(resp, &task); err != nil {
		return nil, err
	}

	return &task, nil
}

// createTask creates a new task.
func (c *client) createTask(ctx context.Context, req *TaskCreateRequest) (*Task, error) {
	resp, err := c.doRequest(ctx, "POST", "/tasks", req)
	if err != nil {
		return nil, err
	}

	var task Task
	if err := decodeResponse(resp, &task); err != nil {
		return nil, err
	}

	return &task, nil
}

// updateTask updates an existing task.
func (c *client) updateTask(ctx context.Context, id string, req *TaskUpdateRequest) (*Task, error) {
	resp, err := c.doRequest(ctx, "POST", "/tasks/"+id, req)
	if err != nil {
		return nil, err
	}

	var task Task
	if err := decodeResponse(resp, &task); err != nil {
		return nil, err
	}

	return &task, nil
}

// closeTask marks a task as completed.
func (c *client) closeTask(ctx context.Context, id string) error {
	resp, err := c.doRequest(ctx, "POST", "/tasks/"+id+"/close", nil)
	if err != nil {
		return err
	}

	return decodeResponse(resp, nil)
}

// reopenTask reopens a completed task.
func (c *client) reopenTask(ctx context.Context, id string) error {
	resp, err := c.doRequest(ctx, "POST", "/tasks/"+id+"/reopen", nil)
	if err != nil {
		return err
	}

	return decodeResponse(resp, nil)
}

// Comment operations

// listComments retrieves comments for a task.
func (c *client) listComments(ctx context.Context, taskID string) ([]*Comment, error) {
	path := fmt.Sprintf("/comments?task_id=%s", taskID)

	resp, err := c.doRequest(ctx, "GET", path, nil)
	if err != nil {
		return nil, err
	}

	var comments []*Comment
	if err := decodeResponse(resp, &comments); err != nil {
		return nil, err
	}

	return comments, nil
}

// createComment creates a new comment on a task.
func (c *client) createComment(ctx context.Context, taskID string, content string) (*Comment, error) {
	req := &CommentCreateRequest{
		TaskID:  taskID,
		Content: content,
	}

	resp, err := c.doRequest(ctx, "POST", "/comments", req)
	if err != nil {
		return nil, err
	}

	var comment Comment
	if err := decodeResponse(resp, &comment); err != nil {
		return nil, err
	}

	return &comment, nil
}
