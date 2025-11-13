package sync

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/evcraddock/todu.sh/internal/api"
	"github.com/evcraddock/todu.sh/internal/registry"
	"github.com/evcraddock/todu.sh/pkg/plugin"
	"github.com/evcraddock/todu.sh/pkg/types"
)

// Engine orchestrates bidirectional synchronization between external systems and the Todu API.
type Engine struct {
	apiClient *api.Client
	registry  *registry.Registry
}

// NewEngine creates a new sync engine with the given API client and plugin registry.
func NewEngine(apiClient *api.Client, registry *registry.Registry) *Engine {
	return &Engine{
		apiClient: apiClient,
		registry:  registry,
	}
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
		log.Println("No projects to sync")
		return result, nil
	}

	log.Printf("Syncing %d project(s)...", len(projects))

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
	return e.apiClient.ListProjects(ctx, options.SystemID)
}

// syncProject synchronizes a single project.
func (e *Engine) syncProject(ctx context.Context, project *types.Project, options Options) ProjectResult {
	pr := ProjectResult{
		ProjectID:   project.ID,
		ProjectName: project.Name,
		Errors:      []error{},
	}

	log.Printf("Syncing project %q (ID: %d)...", project.Name, project.ID)

	// Determine sync strategy
	strategy := e.determineStrategy(project, options)
	log.Printf("  Using strategy: %s", strategy)

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
		e.syncPush(ctx, project, p, options.DryRun, &pr)
	case StrategyBidirectional:
		e.syncPull(ctx, project, p, options.DryRun, &pr)
		e.syncPush(ctx, project, p, options.DryRun, &pr)
	default:
		pr.Errors = append(pr.Errors, fmt.Errorf("unknown strategy: %s", strategy))
	}

	if len(pr.Errors) == 0 {
		log.Printf("  ✓ Project synced: %d created, %d updated, %d skipped", pr.Created, pr.Updated, pr.Skipped)
	} else {
		log.Printf("  ✗ Project sync completed with %d error(s)", len(pr.Errors))
		for i, err := range pr.Errors {
			log.Printf("    Error %d: %v", i+1, err)
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
	// Fetch tasks from external system
	externalTasks, err := p.FetchTasks(ctx, &project.ExternalID, nil)
	if err != nil {
		pr.Errors = append(pr.Errors, fmt.Errorf("failed to fetch external tasks: %w", err))
		return
	}

	// Fetch existing tasks from Todu API
	toduTasks, err := e.apiClient.ListTasks(ctx, &project.ID)
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
			log.Printf("  WARNING: External task has no external_id, skipping")
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
					Labels:      externalTask.Labels,
					Assignees:   externalTask.Assignees,
				}
				createdTask, err := e.apiClient.CreateTask(ctx, taskCreate)
				if err != nil {
					pr.Errors = append(pr.Errors, fmt.Errorf("failed to create task %q: %w", externalTask.Title, err))
					continue
				}
				// Sync comments for newly created task
				e.syncPullComments(ctx, project, p, createdTask, dryRun, pr)
			}
			log.Printf("  → Created task: %s", externalTask.Title)
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
					Labels:      externalTask.Labels,
					Assignees:   externalTask.Assignees,
				}
				_, err := e.apiClient.UpdateTask(ctx, toduTask.ID, taskUpdate)
				if err != nil {
					pr.Errors = append(pr.Errors, fmt.Errorf("failed to update task %q: %w", externalTask.Title, err))
					continue
				}
			}
			log.Printf("  ↻ Updated task: %s", externalTask.Title)
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
func (e *Engine) syncPush(ctx context.Context, project *types.Project, p plugin.Plugin, dryRun bool, pr *ProjectResult) {
	// Fetch tasks from Todu API
	toduTasks, err := e.apiClient.ListTasks(ctx, &project.ID)
	if err != nil {
		pr.Errors = append(pr.Errors, fmt.Errorf("failed to fetch Todu tasks: %w", err))
		return
	}

	// Process all tasks: create new ones and update existing ones
	for _, toduTask := range toduTasks {
		if toduTask.ExternalID == "" {
			// Task doesn't have external_id, create it in external system
			if !dryRun {
				taskCreate := &types.TaskCreate{
					Title:       toduTask.Title,
					Description: toduTask.Description,
					Status:      toduTask.Status,
					Priority:    toduTask.Priority,
					DueDate:     toduTask.DueDate,
					Labels:      toduTask.Labels,
					Assignees:   toduTask.Assignees,
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
				// Update Todu task with external_id and source_url
				taskUpdate := &types.TaskUpdate{
					ExternalID: &createdTask.ExternalID,
					SourceURL:  createdTask.SourceURL,
				}
				_, err = e.apiClient.UpdateTask(ctx, toduTask.ID, taskUpdate)
				if err != nil {
					pr.Errors = append(pr.Errors, fmt.Errorf("failed to update task with external_id: %w", err))
					continue
				}
				log.Printf("  → Created external task: %s (external_id: %s)", toduTask.Title, createdTask.ExternalID)
			} else {
				log.Printf("  → Would create external task: %s", toduTask.Title)
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
			pr.Errors = append(pr.Errors, fmt.Errorf("failed to fetch external task %q: %w", toduTask.ExternalID, err))
			continue
		}

		// Check if Todu task is newer
		if NeedsUpdate(toduTask, externalTask) {
			// Todu task is newer, push to external system
			if !dryRun {
				taskUpdate := &types.TaskUpdate{
					Title:       &toduTask.Title,
					Description: toduTask.Description,
					Status:      &toduTask.Status,
					Priority:    toduTask.Priority,
					DueDate:     toduTask.DueDate,
					Labels:      toduTask.Labels,
					Assignees:   toduTask.Assignees,
				}
				_, err := p.UpdateTask(ctx, &project.ExternalID, toduTask.ExternalID, taskUpdate)
				if err != nil {
					if err == plugin.ErrNotSupported {
						pr.Skipped++
						continue
					}
					pr.Errors = append(pr.Errors, fmt.Errorf("failed to push task %q: %w", toduTask.Title, err))
					continue
				}
			}
			log.Printf("  ← Pushed task: %s", toduTask.Title)
			pr.Updated++
		} else {
			// External task is up to date
			pr.Skipped++
		}

		// Sync comments for tasks being pushed
		if toduTask.ExternalID != "" && !dryRun {
			e.syncPushComments(ctx, project, p, toduTask, dryRun, pr)
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
			log.Printf("    WARNING: External comment has no external_id, skipping")
			continue
		}

		_, exists := toduCommentMap[externalComment.ExternalID]
		if !exists {
			// Comment doesn't exist in Todu, create it
			if !dryRun {
				commentCreate := &types.CommentCreate{
					TaskID:     toduTask.ID,
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
			log.Printf("    → Synced comment from %s", externalComment.Author)
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
		log.Printf("    ← Pushed comment from %s", toduComment.Author)
	}
}
