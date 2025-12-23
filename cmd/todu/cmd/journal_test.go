package cmd

import (
	"testing"
	"time"

	"github.com/evcraddock/todu.sh/pkg/types"
)

func TestFilterJournalsByDate_Until(t *testing.T) {
	// Reset flags after test
	defer func() {
		journalListToday = false
		journalListLast = 0
		journalListSince = ""
		journalListUntil = ""
	}()

	// Create test entries with different dates
	entries := []*types.Comment{
		{ID: 1, Content: "Entry 1", CreatedAt: time.Date(2025, 1, 10, 10, 0, 0, 0, time.Local)},
		{ID: 2, Content: "Entry 2", CreatedAt: time.Date(2025, 1, 15, 14, 30, 0, 0, time.Local)},
		{ID: 3, Content: "Entry 3", CreatedAt: time.Date(2025, 1, 20, 8, 0, 0, 0, time.Local)},
		{ID: 4, Content: "Entry 4", CreatedAt: time.Date(2025, 1, 25, 18, 45, 0, 0, time.Local)},
	}

	tests := []struct {
		name        string
		until       string
		expectedIDs []int
	}{
		{
			name:        "until includes the specified date",
			until:       "2025-01-15",
			expectedIDs: []int{1, 2},
		},
		{
			name:        "until includes entry at end of day",
			until:       "2025-01-20",
			expectedIDs: []int{1, 2, 3},
		},
		{
			name:        "until after all entries returns all",
			until:       "2025-01-31",
			expectedIDs: []int{1, 2, 3, 4},
		},
		{
			name:        "until before all entries returns none",
			until:       "2025-01-05",
			expectedIDs: []int{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			journalListUntil = tt.until

			filtered := filterJournalsByDate(entries)

			if len(filtered) != len(tt.expectedIDs) {
				t.Errorf("filterJournalsByDate() returned %d entries, want %d", len(filtered), len(tt.expectedIDs))
				return
			}

			for i, entry := range filtered {
				if entry.ID != tt.expectedIDs[i] {
					t.Errorf("filterJournalsByDate()[%d].ID = %d, want %d", i, entry.ID, tt.expectedIDs[i])
				}
			}

			// Reset for next test
			journalListUntil = ""
		})
	}
}

func TestFilterJournalsByDate_SinceAndUntil(t *testing.T) {
	// Reset flags after test
	defer func() {
		journalListToday = false
		journalListLast = 0
		journalListSince = ""
		journalListUntil = ""
	}()

	// Create test entries with different dates
	entries := []*types.Comment{
		{ID: 1, Content: "Entry 1", CreatedAt: time.Date(2025, 1, 10, 10, 0, 0, 0, time.Local)},
		{ID: 2, Content: "Entry 2", CreatedAt: time.Date(2025, 1, 15, 14, 30, 0, 0, time.Local)},
		{ID: 3, Content: "Entry 3", CreatedAt: time.Date(2025, 1, 20, 8, 0, 0, 0, time.Local)},
		{ID: 4, Content: "Entry 4", CreatedAt: time.Date(2025, 1, 25, 18, 45, 0, 0, time.Local)},
	}

	tests := []struct {
		name        string
		since       string
		until       string
		expectedIDs []int
	}{
		{
			name:        "date range includes middle entries",
			since:       "2025-01-15",
			until:       "2025-01-20",
			expectedIDs: []int{2, 3},
		},
		{
			name:        "date range with no entries",
			since:       "2025-01-11",
			until:       "2025-01-14",
			expectedIDs: []int{},
		},
		{
			name:        "date range includes all entries",
			since:       "2025-01-01",
			until:       "2025-01-31",
			expectedIDs: []int{1, 2, 3, 4},
		},
		{
			name:        "single day range",
			since:       "2025-01-15",
			until:       "2025-01-15",
			expectedIDs: []int{2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			journalListSince = tt.since
			journalListUntil = tt.until

			filtered := filterJournalsByDate(entries)

			if len(filtered) != len(tt.expectedIDs) {
				t.Errorf("filterJournalsByDate() returned %d entries, want %d", len(filtered), len(tt.expectedIDs))
				return
			}

			for i, entry := range filtered {
				if entry.ID != tt.expectedIDs[i] {
					t.Errorf("filterJournalsByDate()[%d].ID = %d, want %d", i, entry.ID, tt.expectedIDs[i])
				}
			}

			// Reset for next test
			journalListSince = ""
			journalListUntil = ""
		})
	}
}

func TestFilterJournalsByDate_InvalidUntilDate(t *testing.T) {
	// Reset flags after test
	defer func() {
		journalListUntil = ""
	}()

	entries := []*types.Comment{
		{ID: 1, Content: "Entry 1", CreatedAt: time.Date(2025, 1, 15, 10, 0, 0, 0, time.Local)},
	}

	// Invalid date format should be ignored, returning all entries
	journalListUntil = "invalid-date"

	filtered := filterJournalsByDate(entries)

	if len(filtered) != 1 {
		t.Errorf("filterJournalsByDate() with invalid date returned %d entries, want 1", len(filtered))
	}
}
