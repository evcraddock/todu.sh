package cmd

import (
	"testing"
	"time"

	"github.com/evcraddock/todu.sh/internal/api"
)

func TestBuildCommentListOptions_Until(t *testing.T) {
	tests := []struct {
		name           string
		until          string
		expectedBefore string
	}{
		{
			name:           "until adds one day for inclusive behavior",
			until:          "2025-01-15",
			expectedBefore: "2025-01-16",
		},
		{
			name:           "until end of month",
			until:          "2025-01-31",
			expectedBefore: "2025-02-01",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &api.CommentListOptions{
				Type: "journal",
			}

			// Simulate the until logic from runJournalList
			untilDate, err := time.ParseInLocation("2006-01-02", tt.until, time.Local)
			if err == nil {
				opts.CreatedBefore = untilDate.AddDate(0, 0, 1).Format("2006-01-02")
			}

			if opts.CreatedBefore != tt.expectedBefore {
				t.Errorf("CreatedBefore = %s, want %s", opts.CreatedBefore, tt.expectedBefore)
			}
		})
	}
}

func TestBuildCommentListOptions_SinceAndUntil(t *testing.T) {
	tests := []struct {
		name           string
		since          string
		until          string
		expectedAfter  string
		expectedBefore string
	}{
		{
			name:           "date range for January 2025",
			since:          "2025-01-01",
			until:          "2025-01-31",
			expectedAfter:  "2025-01-01",
			expectedBefore: "2025-02-01",
		},
		{
			name:           "single day range",
			since:          "2025-01-15",
			until:          "2025-01-15",
			expectedAfter:  "2025-01-15",
			expectedBefore: "2025-01-16",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := &api.CommentListOptions{
				Type: "journal",
			}

			// Simulate the since/until logic from runJournalList
			opts.CreatedAfter = tt.since

			untilDate, err := time.ParseInLocation("2006-01-02", tt.until, time.Local)
			if err == nil {
				opts.CreatedBefore = untilDate.AddDate(0, 0, 1).Format("2006-01-02")
			}

			if opts.CreatedAfter != tt.expectedAfter {
				t.Errorf("CreatedAfter = %s, want %s", opts.CreatedAfter, tt.expectedAfter)
			}
			if opts.CreatedBefore != tt.expectedBefore {
				t.Errorf("CreatedBefore = %s, want %s", opts.CreatedBefore, tt.expectedBefore)
			}
		})
	}
}
