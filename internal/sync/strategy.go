package sync

// Strategy defines the synchronization strategy for a project.
type Strategy string

const (
	// StrategyPull synchronizes from external system to Todu only.
	// Changes in Todu will not be pushed to the external system.
	StrategyPull Strategy = "pull"

	// StrategyPush synchronizes from Todu to external system only.
	// Changes in the external system will not be pulled to Todu.
	StrategyPush Strategy = "push"

	// StrategyBidirectional synchronizes in both directions.
	// Changes in either system will be synced to the other.
	// Uses last-write-wins for conflict resolution.
	StrategyBidirectional Strategy = "bidirectional"
)

// IsValid checks if the strategy is a valid value.
func (s Strategy) IsValid() bool {
	switch s {
	case StrategyPull, StrategyPush, StrategyBidirectional:
		return true
	default:
		return false
	}
}
