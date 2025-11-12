package github

// mapper.go contains functions for converting between GitHub API types and Todu types.
//
// Mappings:
//   - GitHub Repository → Todu Project (external_id = "owner/repo")
//   - GitHub Issue → Todu Task (external_id = issue number as string)
//   - GitHub Issue State (open/closed) → Todu Status (active/done)
//   - GitHub Labels → Todu Labels (priority extracted from priority:* labels)
//   - GitHub Issue Comments → Todu Comments (1:1 mapping)
//
// Priority Mapping:
//   - Labels matching "priority:high" → high priority
//   - Labels matching "priority:medium" → medium priority
//   - Labels matching "priority:low" → low priority
//   - No priority label → no priority set

// This file is a placeholder for type conversion functions.
// Implementation will be added in Unit 3.2.
