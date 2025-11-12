package sync

import "time"

// Result represents the outcome of a sync operation.
type Result struct {
	// ProjectResults contains results for each project that was synced.
	ProjectResults []ProjectResult

	// TotalCreated is the total number of tasks created across all projects.
	TotalCreated int

	// TotalUpdated is the total number of tasks updated across all projects.
	TotalUpdated int

	// TotalSkipped is the total number of tasks skipped across all projects.
	TotalSkipped int

	// TotalErrors is the total number of errors across all projects.
	TotalErrors int

	// Duration is the total time taken for the sync operation.
	Duration time.Duration
}

// ProjectResult represents the outcome of syncing a single project.
type ProjectResult struct {
	// ProjectID is the Todu API project ID.
	ProjectID int

	// ProjectName is the project name for display.
	ProjectName string

	// Created is the number of tasks created in this project.
	Created int

	// Updated is the number of tasks updated in this project.
	Updated int

	// Skipped is the number of tasks skipped in this project.
	Skipped int

	// Errors contains any errors that occurred during sync.
	Errors []error
}

// HasErrors returns true if any errors occurred during sync.
func (r *Result) HasErrors() bool {
	return r.TotalErrors > 0
}

// AddProjectResult adds a project result and updates the totals.
func (r *Result) AddProjectResult(pr ProjectResult) {
	r.ProjectResults = append(r.ProjectResults, pr)
	r.TotalCreated += pr.Created
	r.TotalUpdated += pr.Updated
	r.TotalSkipped += pr.Skipped
	r.TotalErrors += len(pr.Errors)
}
