package cmd

import (
	"testing"
)

func TestValidateRRule(t *testing.T) {
	tests := []struct {
		name    string
		rrule   string
		wantErr bool
	}{
		{
			name:    "valid daily rule",
			rrule:   "FREQ=DAILY;INTERVAL=1",
			wantErr: false,
		},
		{
			name:    "valid weekly rule with days",
			rrule:   "FREQ=WEEKLY;BYDAY=MO,WE,FR",
			wantErr: false,
		},
		{
			name:    "valid monthly rule",
			rrule:   "FREQ=MONTHLY;BYMONTHDAY=1",
			wantErr: false,
		},
		{
			name:    "valid yearly rule",
			rrule:   "FREQ=YEARLY;BYMONTH=1;BYMONTHDAY=1",
			wantErr: false,
		},
		{
			name:    "lowercase freq is valid",
			rrule:   "freq=daily",
			wantErr: false,
		},
		{
			name:    "missing FREQ",
			rrule:   "INTERVAL=1;BYDAY=MO",
			wantErr: true,
		},
		{
			name:    "invalid frequency",
			rrule:   "FREQ=HOURLY",
			wantErr: true,
		},
		{
			name:    "empty string",
			rrule:   "",
			wantErr: true,
		},
		{
			name:    "invalid part",
			rrule:   "FREQ=DAILY;INVALID=VALUE",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRRule(tt.rrule)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateRRule(%q) error = %v, wantErr %v", tt.rrule, err, tt.wantErr)
			}
		})
	}
}

func TestValidateTimezone(t *testing.T) {
	tests := []struct {
		name    string
		tz      string
		wantErr bool
	}{
		{
			name:    "UTC",
			tz:      "UTC",
			wantErr: false,
		},
		{
			name:    "America/New_York",
			tz:      "America/New_York",
			wantErr: false,
		},
		{
			name:    "Europe/London",
			tz:      "Europe/London",
			wantErr: false,
		},
		{
			name:    "Asia/Tokyo",
			tz:      "Asia/Tokyo",
			wantErr: false,
		},
		{
			name:    "Local",
			tz:      "Local",
			wantErr: false,
		},
		{
			name:    "invalid timezone",
			tz:      "Invalid/Zone",
			wantErr: true,
		},
		{
			name:    "empty string returns UTC",
			tz:      "",
			wantErr: false, // Empty string is valid and returns UTC
		},
		{
			name:    "partial timezone",
			tz:      "America",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateTimezone(tt.tz)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateTimezone(%q) error = %v, wantErr %v", tt.tz, err, tt.wantErr)
			}
		})
	}
}

func TestRRuleToHuman(t *testing.T) {
	tests := []struct {
		name     string
		rrule    string
		expected string
	}{
		{
			name:     "daily",
			rrule:    "FREQ=DAILY",
			expected: "Daily",
		},
		{
			name:     "daily interval 3",
			rrule:    "FREQ=DAILY;INTERVAL=3",
			expected: "Every 3 days",
		},
		{
			name:     "weekly",
			rrule:    "FREQ=WEEKLY",
			expected: "Weekly",
		},
		{
			name:     "weekly on Monday",
			rrule:    "FREQ=WEEKLY;BYDAY=MO",
			expected: "Weekly on Mon",
		},
		{
			name:     "weekly on Mon, Wed, Fri",
			rrule:    "FREQ=WEEKLY;BYDAY=MO,WE,FR",
			expected: "Weekly on Mon, Wed, Fri",
		},
		{
			name:     "every 2 weeks",
			rrule:    "FREQ=WEEKLY;INTERVAL=2",
			expected: "Every 2 weeks",
		},
		{
			name:     "monthly",
			rrule:    "FREQ=MONTHLY",
			expected: "Monthly",
		},
		{
			name:     "monthly on 1st",
			rrule:    "FREQ=MONTHLY;BYMONTHDAY=1",
			expected: "Monthly on day 1",
		},
		{
			name:     "yearly",
			rrule:    "FREQ=YEARLY",
			expected: "Yearly",
		},
		{
			name:     "every 2 years",
			rrule:    "FREQ=YEARLY;INTERVAL=2",
			expected: "Every 2 years",
		},
		{
			name:     "daily on weekdays",
			rrule:    "FREQ=DAILY;BYDAY=MO,TU,WE,TH,FR",
			expected: "Daily on Mon, Tue, Wed, Thu, Fri",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rruleToHuman(tt.rrule)
			if result != tt.expected {
				t.Errorf("rruleToHuman(%q) = %q, expected %q", tt.rrule, result, tt.expected)
			}
		})
	}
}

func TestFormatDays(t *testing.T) {
	tests := []struct {
		name     string
		byday    string
		expected string
	}{
		{
			name:     "single day",
			byday:    "MO",
			expected: "Mon",
		},
		{
			name:     "multiple days",
			byday:    "MO,WE,FR",
			expected: "Mon, Wed, Fri",
		},
		{
			name:     "all weekdays",
			byday:    "MO,TU,WE,TH,FR",
			expected: "Mon, Tue, Wed, Thu, Fri",
		},
		{
			name:     "weekend",
			byday:    "SA,SU",
			expected: "Sat, Sun",
		},
		{
			name:     "lowercase",
			byday:    "mo,we,fr",
			expected: "Mon, Wed, Fri",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatDays(tt.byday)
			if result != tt.expected {
				t.Errorf("formatDays(%q) = %q, expected %q", tt.byday, result, tt.expected)
			}
		})
	}
}
