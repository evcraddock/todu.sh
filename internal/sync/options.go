package sync

// Options configures a sync operation.
type Options struct {
	// ProjectIDs specifies which projects to sync.
	// If empty, all projects will be synced.
	ProjectIDs []int

	// SystemID filters projects by system.
	// If nil, projects from all systems will be synced.
	SystemID *int

	// StrategyOverride overrides the project's configured sync strategy.
	// If nil, each project's sync_strategy field will be used.
	// This allows command-line overrides of per-project defaults.
	StrategyOverride *Strategy

	// DryRun when true will simulate the sync without making any changes.
	// The result will show what would have been created/updated.
	DryRun bool

	// Force when true will push all tasks regardless of last_pushed_at.
	// Use this to re-sync tasks that may have been pushed with missing data.
	Force bool
}
