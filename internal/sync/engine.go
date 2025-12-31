package sync

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/rs/zerolog"

	"github.com/evcraddock/todu.sh/internal/api"
	"github.com/evcraddock/todu.sh/internal/registry"
	"github.com/evcraddock/todu.sh/pkg/plugin"
	"github.com/evcraddock/todu.sh/pkg/types"
)

// Engine orchestrates bidirectional synchronization between external systems and the Todu API.
type Engine struct {
	apiClient *api.Client
	registry  *registry.Registry
	logger    zerolog.Logger
}

// NewEngine creates a new sync engine with the given API client and plugin registry.
func NewEngine(apiClient *api.Client, registry *registry.Registry) *Engine {
	// Default to a disabled logger (no output)
	return &Engine{
		apiClient: apiClient,
		registry:  registry,
		logger:    zerolog.New(io.Discard),
	}
}

// WithLogger sets a logger for the sync engine
func (e *Engine) WithLogger(logger zerolog.Logger) *Engine {
	e.logger = logger
	return e
}

// Sync performs synchronization based on the provided options.
// Returns a Result summarizing what was synced and any errors encountered.
func (e *Engine) Sync(ctx context.Context, options Options) (*Result, error) {
	startTime := time.Now()
	result := &Result{}

	// Get projects to sync
	projects, err := e.getProjectsToSync(ctx, options)
	if err != nil {
		return nil, fmt.Errorf("failed to get projects: %w", err)
	}

	if len(projects) == 0 {
		e.logger.Debug().Msg("No projects to sync")
		return result, nil
	}

	e.logger.Debug().Int("count", len(projects)).Msg("Syncing projects...")

	// Sync each project
	for _, project := range projects {
		pr := e.syncProject(ctx, project, options)
		result.AddProjectResult(pr)
	}

	result.Duration = time.Since(startTime)
	return result, nil
}

// getProjectsToSync retrieves the list of projects to sync based on options.
func (e *Engine) getProjectsToSync(ctx context.Context, options Options) ([]*types.Project, error) {
	// If specific project IDs are provided, fetch those
	if len(options.ProjectIDs) > 0 {
		projects := make([]*types.Project, 0, len(options.ProjectIDs))
		for _, id := range options.ProjectIDs {
			project, err := e.apiClient.GetProject(ctx, id)
			if err != nil {
				return nil, fmt.Errorf("failed to get project %d: %w", id, err)
			}
			projects = append(projects, project)
		}
		return projects, nil
	}

	// Otherwise, list projects (optionally filtered by system)
	opts := &api.ProjectListOptions{SystemID: options.SystemID}
	return e.apiClient.ListProjects(ctx, opts)
}

// syncProject synchronizes a single project.
func (e *Engine) syncProject(ctx context.Context, project *types.Project, options Options) ProjectResult {
	pr := ProjectResult{
		ProjectID:   project.ID,
		ProjectName: project.Name,
		Errors:      []error{},
	}

	e.logger.Debug().Str("project", project.Name).Int("id", project.ID).Msg("Syncing project")

	// Determine sync strategy
	strategy := e.determineStrategy(project, options)
	e.logger.Debug().Str("project", project.Name).Str("strategy", string(strategy)).Msg("Using strategy")

	// Get system for project
	system, err := e.apiClient.GetSystem(ctx, project.SystemID)
	if err != nil {
		pr.Errors = append(pr.Errors, fmt.Errorf("failed to get system: %w", err))
		return pr
	}

	// Create plugin instance
	pluginConfig, err := registry.LoadPluginConfig(system.Identifier)
	if err != nil {
		pr.Errors = append(pr.Errors, fmt.Errorf("failed to load plugin config: %w", err))
		return pr
	}
	p, err := e.registry.Create(system.Identifier, pluginConfig)
	if err != nil {
		pr.Errors = append(pr.Errors, fmt.Errorf("failed to create plugin: %w", err))
		return pr
	}

	// Validate plugin configuration
	if err := p.ValidateConfig(); err != nil {
		pr.Errors = append(pr.Errors, fmt.Errorf("plugin not configured: %w", err))
		return pr
	}

	// Perform sync based on strategy
	switch strategy {
	case StrategyPull:
		e.syncPull(ctx, project, p, options.DryRun, &pr)
	case StrategyPush:
		e.syncPush(ctx, project, p, options, &pr)
	case StrategyBidirectional:
		e.syncPull(ctx, project, p, options.DryRun, &pr)
		e.syncPush(ctx, project, p, options, &pr)
	default:
		pr.Errors = append(pr.Errors, fmt.Errorf("unknown strategy: %s", strategy))
	}

	if len(pr.Errors) == 0 {
		e.logger.Debug().
			Str("project", project.Name).
			Int("created", pr.Created).
			Int("updated", pr.Updated).
			Int("skipped", pr.Skipped).
			Msg("Project synced")
		// Update last_synced_at timestamp on successful sync
		if !options.DryRun {
			now := time.Now()
			projectUpdate := &types.ProjectUpdate{
				LastSyncedAt: &now,
			}
			_, err := e.apiClient.UpdateProject(ctx, project.ID, projectUpdate)
			if err != nil {
				e.logger.Warn().Err(err).Str("project", project.Name).Msg("Failed to update last_synced_at")
			}
		}
	} else {
		e.logger.Debug().
			Str("project", project.Name).
			Int("error_count", len(pr.Errors)).
			Msg("Project sync completed with errors")
		for _, err := range pr.Errors {
			e.logger.Debug().Str("project", project.Name).Err(err).Msg("Project sync error detail")
		}
	}

	return pr
}

// determineStrategy determines which sync strategy to use for a project.
func (e *Engine) determineStrategy(project *types.Project, options Options) Strategy {
	// Use override if provided
	if options.StrategyOverride != nil {
		return *options.StrategyOverride
	}

	// Otherwise use project's configured strategy
	return Strategy(project.SyncStrategy)
}

// syncPull pulls tasks from external system to Todu.
func (e *Engine) syncPull(ctx context.Context, project *types.Project, p plugin.Plugin, dryRun bool, pr *ProjectResult) {
	// Fetch tasks from external system (use LastSyncedAt for incremental sync)
	externalTasks, err := p.FetchTasks(ctx, &project.ExternalID, project.LastSyncedAt)
	if err != nil {
		pr.Errors = append(pr.Errors, fmt.Errorf("failed to fetch external tasks: %w", err))
		return
	}

	// Fetch existing tasks from Todu API
	toduTasks, err := e.apiClient.ListTasks(ctx, &api.TaskListOptions{ProjectID: &project.ID})
	if err != nil {
		pr.Errors = append(pr.Errors, fmt.Errorf("failed to fetch Todu tasks: %w", err))
		return
	}

	// Build map of Todu tasks by external_id for quick lookup
	toduTaskMap := make(map[string]*types.Task)
	for _, task := range toduTasks {
		if task.ExternalID != "" {
			toduTaskMap[task.ExternalID] = task
		}
	}

	// Process each external task
	for _, externalTask := range externalTasks {
		if externalTask.ExternalID == "" {
			e.logger.Debug().Msg("External task has no external_id, skipping")
			pr.Skipped++
			continue
		}

		toduTask, exists := toduTaskMap[externalTask.ExternalID]

		if !exists {
			// Task doesn't exist in Todu, create it
			if !dryRun {
				taskCreate := &types.TaskCreate{
					ExternalID:  externalTask.ExternalID,
					SourceURL:   externalTask.SourceURL,
					Title:       externalTask.Title,
					Description: externalTask.Description,
					ProjectID:   project.ID,
					Status:      externalTask.Status,
					Priority:    externalTask.Priority,
					DueDate:     externalTask.DueDate,
					Labels:      extractLabelNames(externalTask.Labels),
					Assignees:   extractAssigneeNames(externalTask.Assignees),
				}
				createdTask, err := e.apiClient.CreateTask(ctx, taskCreate)
				if err != nil {
					pr.Errors = append(pr.Errors, fmt.Errorf("failed to create task %q: %w", externalTask.Title, err))
					continue
				}
				// Sync comments for newly created task
				e.syncPullComments(ctx, project, p, createdTask, dryRun, pr)
			}
			e.logger.Debug().Str("task", externalTask.Title).Msg("Created task")
			pr.Created++
		} else if NeedsUpdate(externalTask, toduTask) {
			// External task is newer, update Todu task
			if !dryRun {
				taskUpdate := &types.TaskUpdate{
					Title:       &externalTask.Title,
					Description: externalTask.Description,
					Status:      &externalTask.Status,
					Priority:    externalTask.Priority,
					DueDate:     externalTask.DueDate,
					Labels:      extractLabelNames(externalTask.Labels),
					Assignees:   extractAssigneeNames(externalTask.Assignees),
				}
				_, err := e.apiClient.UpdateTask(ctx, toduTask.ID, taskUpdate)
				if err != nil {
					pr.Errors = append(pr.Errors, fmt.Errorf("failed to update task %q: %w", externalTask.Title, err))
					continue
				}
			}
			e.logger.Debug().Str("task", externalTask.Title).Msg("Updated task")
			pr.Updated++
		} else {
			// Todu task is up to date
			pr.Skipped++
		}

		// Sync comments for existing tasks
		if exists && !dryRun {
			e.syncPullComments(ctx, project, p, toduTask, dryRun, pr)
		}
	}
}

// syncPush pushes tasks from Todu to external system.
func (e *Engine) syncPush(ctx context.Context, project *types.Project, p plugin.Plugin, options Options, pr *ProjectResult) {
	// Fetch tasks from Todu API
	toduTasks, err := e.apiClient.ListTasks(ctx, &api.TaskListOptions{ProjectID: &project.ID})
	if err != nil {
		pr.Errors = append(pr.Errors, fmt.Errorf("failed to fetch Todu tasks: %w", err))
		return
	}

	// Process all tasks: create new ones and update existing ones
	for _, toduTask := range toduTasks {
		// Skip tasks that haven't been modified since last successful push (optimization)
		// Uses per-task last_pushed_at instead of project-level last_synced_at for accurate tracking
		// Force flag bypasses this check to allow re-pushing all tasks
		if !options.Force && toduTask.ExternalID != "" && toduTask.LastPushedAt != nil && !toduTask.UpdatedAt.After(*toduTask.LastPushedAt) {
			pr.Skipped++
			continue
		}

		if toduTask.ExternalID == "" {
			// Task doesn't have external_id, create it in external system
			if !options.DryRun {
				// Fetch full task details to get description (not included in list response)
				fullTask, err := e.apiClient.GetTask(ctx, toduTask.ID)
				if err != nil {
					pr.Errors = append(pr.Errors, fmt.Errorf("failed to fetch full task %q: %w", toduTask.Title, err))
					continue
				}
				taskCreate := &types.TaskCreate{
					Title:       fullTask.Title,
					Description: fullTask.Description,
					Status:      fullTask.Status,
					Priority:    fullTask.Priority,
					DueDate:     fullTask.DueDate,
					Labels:      extractLabelNames(fullTask.Labels),
					Assignees:   extractAssigneeNames(fullTask.Assignees),
				}
				createdTask, err := p.CreateTask(ctx, &project.ExternalID, taskCreate)
				if err != nil {
					if err == plugin.ErrNotSupported {
						pr.Skipped++
						continue
					}
					pr.Errors = append(pr.Errors, fmt.Errorf("failed to create external task %q: %w", toduTask.Title, err))
					continue
				}

				// If task is already done/canceled in Todu, close it in external system
				// (GitHub API doesn't support creating issues in closed state)
				if fullTask.Status == "done" || fullTask.Status == "canceled" {
					statusUpdate := &types.TaskUpdate{
						Status: &fullTask.Status,
					}
					_, err = p.UpdateTask(ctx, &project.ExternalID, createdTask.ExternalID, statusUpdate)
					if err != nil && err != plugin.ErrNotSupported {
						pr.Errors = append(pr.Errors, fmt.Errorf("failed to close external task %q: %w", toduTask.Title, err))
						// Continue anyway - task was created, just not closed
					}
				}

				// Update Todu task with external_id, source_url, and last_pushed_at
				now := time.Now()
				taskUpdate := &types.TaskUpdate{
					ExternalID:   &createdTask.ExternalID,
					SourceURL:    createdTask.SourceURL,
					LastPushedAt: &now,
				}
				_, err = e.apiClient.UpdateTask(ctx, toduTask.ID, taskUpdate)
				if err != nil {
					pr.Errors = append(pr.Errors, fmt.Errorf("failed to update task with external_id: %w", err))
					continue
				}
				e.logger.Debug().Str("task", toduTask.Title).Str("external_id", createdTask.ExternalID).Msg("Created external task")
			} else {
				e.logger.Debug().Str("task", toduTask.Title).Msg("Would create external task (dry run)")
			}
			pr.Created++
			continue
		}

		// Fetch current state from external system
		externalTask, err := p.FetchTask(ctx, &project.ExternalID, toduTask.ExternalID)
		if err != nil {
			// If task not found in external system, skip (may have been deleted)
			if err == plugin.ErrNotSupported {
				pr.Skipped++
				continue
			}
			// If task was deleted externally (not found), try to handle completed tasks
			if errors.Is(err, plugin.ErrNotFound) {
				// If local task is done/canceled, it may already be completed externally
				// Try to close it anyway in case it exists but is just not in active list
				if toduTask.Status == "done" || toduTask.Status == "canceled" {
					if !options.DryRun {
						taskUpdate := &types.TaskUpdate{
							Status: &toduTask.Status,
						}
						_, closeErr := p.UpdateTask(ctx, &project.ExternalID, toduTask.ExternalID, taskUpdate)
						if closeErr == nil {
							e.logger.Debug().Str("task", toduTask.Title).Msg("Closed task externally")
							pr.Updated++
							continue
						}
						// If close failed, task is truly gone
					}
				}
				e.logger.Debug().Str("task", toduTask.Title).Msg("Task no longer exists externally, skipping")
				pr.Skipped++
				continue
			}
			pr.Errors = append(pr.Errors, fmt.Errorf("failed to fetch external task %q: %w", toduTask.ExternalID, err))
			continue
		}

		// Check if Todu task is newer (or force flag is set)
		if options.Force || NeedsUpdate(toduTask, externalTask) {
			// Todu task is newer, push to external system
			if !options.DryRun {
				// Fetch full task details to get description (not included in list response)
				fullTask, err := e.apiClient.GetTask(ctx, toduTask.ID)
				if err != nil {
					pr.Errors = append(pr.Errors, fmt.Errorf("failed to fetch full task %q: %w", toduTask.Title, err))
					continue
				}
				taskUpdate := &types.TaskUpdate{
					Title:       &fullTask.Title,
					Description: fullTask.Description,
					Status:      &fullTask.Status,
					Priority:    fullTask.Priority,
					DueDate:     fullTask.DueDate,
					Labels:      extractLabelNames(fullTask.Labels),
					Assignees:   extractAssigneeNames(fullTask.Assignees),
				}
				_, err = p.UpdateTask(ctx, &project.ExternalID, toduTask.ExternalID, taskUpdate)
				if err != nil {
					if err == plugin.ErrNotSupported {
						pr.Skipped++
						continue
					}
					pr.Errors = append(pr.Errors, fmt.Errorf("failed to push task %q: %w", toduTask.Title, err))
					continue
				}
				// Update last_pushed_at after successful push (skip if force to preserve timestamps)
				if !options.Force {
					now := time.Now()
					lastPushedUpdate := &types.TaskUpdate{
						LastPushedAt: &now,
					}
					_, err = e.apiClient.UpdateTask(ctx, toduTask.ID, lastPushedUpdate)
					if err != nil {
						e.logger.Warn().Err(err).Str("task", toduTask.Title).Msg("Failed to update last_pushed_at")
						// Don't fail the sync, just log the warning
					}
				}
			}
			e.logger.Debug().Str("task", toduTask.Title).Msg("Pushed task")
			pr.Updated++
		} else {
			// External task is up to date
			pr.Skipped++
		}

		// Sync comments for tasks being pushed
		if toduTask.ExternalID != "" && !options.DryRun {
			e.syncPushComments(ctx, project, p, toduTask, options.DryRun, pr)
		}
	}
}

// syncPullComments pulls comments from external system to Todu for a specific task.
func (e *Engine) syncPullComments(ctx context.Context, project *types.Project, p plugin.Plugin, toduTask *types.Task, dryRun bool, pr *ProjectResult) {
	// Skip if task has no external_id
	if toduTask.ExternalID == "" {
		return
	}

	// Fetch comments from external system
	externalComments, err := p.FetchComments(ctx, &project.ExternalID, toduTask.ExternalID)
	if err != nil {
		// If plugin doesn't support comments, silently skip
		if err == plugin.ErrNotSupported {
			return
		}
		pr.Errors = append(pr.Errors, fmt.Errorf("failed to fetch comments for task %s: %w", toduTask.Title, err))
		return
	}

	// Fetch existing comments from Todu API
	toduComments, err := e.apiClient.ListComments(ctx, toduTask.ID)
	if err != nil {
		pr.Errors = append(pr.Errors, fmt.Errorf("failed to fetch Todu comments for task %s: %w", toduTask.Title, err))
		return
	}

	// Build map of Todu comments by external_id for deduplication
	toduCommentMap := make(map[string]*types.Comment)
	for _, comment := range toduComments {
		if comment.ExternalID != "" {
			toduCommentMap[comment.ExternalID] = comment
		}
	}

	// Process each external comment
	for _, externalComment := range externalComments {
		if externalComment.ExternalID == "" {
			e.logger.Debug().Msg("External comment has no external_id, skipping")
			continue
		}

		_, exists := toduCommentMap[externalComment.ExternalID]
		if !exists {
			// Comment doesn't exist in Todu, create it
			if !dryRun {
				taskID := toduTask.ID
				commentCreate := &types.CommentCreate{
					TaskID:     &taskID,
					ExternalID: externalComment.ExternalID,
					Content:    externalComment.Content,
					Author:     externalComment.Author,
				}
				_, err := e.apiClient.CreateComment(ctx, commentCreate)
				if err != nil {
					pr.Errors = append(pr.Errors, fmt.Errorf("failed to create comment on task %s: %w", toduTask.Title, err))
					continue
				}
			}
			e.logger.Debug().Str("author", externalComment.Author).Msg("Synced comment")
		}
		// Note: Comments are typically immutable after creation, so we don't update them
	}
}

// syncPushComments pushes comments from Todu to external system for a specific task.
func (e *Engine) syncPushComments(ctx context.Context, project *types.Project, p plugin.Plugin, toduTask *types.Task, dryRun bool, pr *ProjectResult) {
	// Skip if task has no external_id
	if toduTask.ExternalID == "" {
		return
	}

	// Fetch comments from Todu API
	toduComments, err := e.apiClient.ListComments(ctx, toduTask.ID)
	if err != nil {
		pr.Errors = append(pr.Errors, fmt.Errorf("failed to fetch Todu comments for task %s: %w", toduTask.Title, err))
		return
	}

	// Fetch external comments to check what's already synced
	externalComments, err := p.FetchComments(ctx, &project.ExternalID, toduTask.ExternalID)
	if err != nil {
		// If plugin doesn't support comments, silently skip
		if err == plugin.ErrNotSupported {
			return
		}
		pr.Errors = append(pr.Errors, fmt.Errorf("failed to fetch external comments for task %s: %w", toduTask.Title, err))
		return
	}

	// Build map of external comments by external_id for deduplication
	externalCommentMap := make(map[string]*types.Comment)
	for _, comment := range externalComments {
		if comment.ExternalID != "" {
			externalCommentMap[comment.ExternalID] = comment
		}
	}

	// Process each Todu comment
	for _, toduComment := range toduComments {
		// Skip comments that already have external_id (already synced)
		if toduComment.ExternalID != "" {
			continue
		}

		// Create comment in external system
		if !dryRun {
			commentCreate := &types.CommentCreate{
				Content: toduComment.Content,
				Author:  toduComment.Author,
			}
			createdComment, err := p.CreateComment(ctx, &project.ExternalID, toduTask.ExternalID, commentCreate)
			if err != nil {
				if err == plugin.ErrNotSupported {
					return
				}
				pr.Errors = append(pr.Errors, fmt.Errorf("failed to push comment to task %s: %w", toduTask.Title, err))
				continue
			}

			// Update the Todu comment with the external_id
			if createdComment.ExternalID != "" {
				commentUpdate := &types.CommentUpdate{
					Content:    &toduComment.Content,
					ExternalID: &createdComment.ExternalID,
				}
				_, err = e.apiClient.UpdateComment(ctx, toduComment.ID, commentUpdate)
				if err != nil {
					pr.Errors = append(pr.Errors, fmt.Errorf("failed to update comment external_id for task %s: %w", toduTask.Title, err))
					// Continue anyway - comment was created in external system
				}
			}
		}
		e.logger.Debug().Str("author", toduComment.Author).Msg("Pushed comment")
	}
}

// extractLabelNames extracts just the label names as strings from a slice of Label structs.
// This is needed because the API expects label names as strings, not Label objects.
func extractLabelNames(labels []types.Label) []string {
	if labels == nil {
		return nil
	}
	result := make([]string, 0, len(labels))
	for _, label := range labels {
		if label.Name != "" {
			result = append(result, label.Name)
		}
	}
	return result
}

// extractAssigneeNames extracts just the assignee names as strings from a slice of Assignee structs.
// This is needed because the API expects assignee names as strings, not Assignee objects.
func extractAssigneeNames(assignees []types.Assignee) []string {
	if assignees == nil {
		return nil
	}
	result := make([]string, 0, len(assignees))
	for _, assignee := range assignees {
		if assignee.Name != "" {
			result = append(result, assignee.Name)
		}
	}
	return result
}
