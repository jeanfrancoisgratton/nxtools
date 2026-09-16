// nxtools
// Unit tests for the pure flag-validation logic backing `task create
// blob-compact`. These don't touch the network: buildFrequency,
// parseStartDate and normalizeNotifyCondition are deterministic functions of
// their string/int inputs.

package cmd

import (
	"testing"
	"time"
)

func TestBuildFrequency_Manual(t *testing.T) {
	freq, err := buildFrequency("manual", "", "", nil, "")
	if err != nil {
		t.Fatalf("buildFrequency(manual): %v", err)
	}
	if freq.Schedule != "manual" {
		t.Errorf("Schedule = %q, want manual", freq.Schedule)
	}
	if freq.StartDate != nil || freq.TimeZoneOffset != nil || freq.CronExpression != nil || len(freq.RecurringDays) != 0 {
		t.Errorf("manual schedule should carry no other fields, got %+v", freq)
	}
}

func TestBuildFrequency_InvalidSchedule(t *testing.T) {
	if _, err := buildFrequency("bogus", "", "", nil, ""); err == nil {
		t.Fatal("expected an error for an unsupported schedule value")
	}
}

func TestBuildFrequency_OnceRequiresStartDate(t *testing.T) {
	if _, err := buildFrequency("once", "", "", nil, ""); err == nil {
		t.Fatal("expected an error when --start-date is missing for a once schedule")
	}
	freq, err := buildFrequency("once", "2026-09-20T02:00", "", nil, "")
	if err != nil {
		t.Fatalf("buildFrequency(once): %v", err)
	}
	if freq.StartDate == nil {
		t.Fatal("expected StartDate to be set")
	}
}

func TestBuildFrequency_CronRequiresCronExpression(t *testing.T) {
	if _, err := buildFrequency("cron", "", "", nil, ""); err == nil {
		t.Fatal("expected an error when --cron is missing for a cron schedule")
	}
	freq, err := buildFrequency("cron", "", "", nil, "0 0 12 * * ?")
	if err != nil {
		t.Fatalf("buildFrequency(cron): %v", err)
	}
	if freq.CronExpression == nil || *freq.CronExpression != "0 0 12 * * ?" {
		t.Errorf("CronExpression = %v, want \"0 0 12 * * ?\"", freq.CronExpression)
	}
	if freq.StartDate != nil {
		t.Error("cron schedule shouldn't require a start date")
	}
}

func TestBuildFrequency_WeeklyRequiresStartDateAndDays(t *testing.T) {
	if _, err := buildFrequency("weekly", "2026-09-20", "", nil, ""); err == nil {
		t.Fatal("expected an error when --days is missing for a weekly schedule")
	}
	freq, err := buildFrequency("weekly", "2026-09-20", "", []int{1, 4}, "")
	if err != nil {
		t.Fatalf("buildFrequency(weekly): %v", err)
	}
	if len(freq.RecurringDays) != 2 {
		t.Errorf("RecurringDays = %v, want [1 4]", freq.RecurringDays)
	}
}

func TestBuildFrequency_WeeklyRejectsOutOfRangeDay(t *testing.T) {
	if _, err := buildFrequency("weekly", "2026-09-20", "", []int{8}, ""); err == nil {
		t.Fatal("expected an error for a weekly day out of the 1-7 range")
	}
}

func TestBuildFrequency_MonthlyAllowsLastDaySentinel(t *testing.T) {
	freq, err := buildFrequency("monthly", "2026-09-20", "", []int{999}, "")
	if err != nil {
		t.Fatalf("buildFrequency(monthly, 999): %v", err)
	}
	if len(freq.RecurringDays) != 1 || freq.RecurringDays[0] != 999 {
		t.Errorf("RecurringDays = %v, want [999]", freq.RecurringDays)
	}
}

func TestBuildFrequency_MonthlyRejectsOutOfRangeDay(t *testing.T) {
	if _, err := buildFrequency("monthly", "2026-09-20", "", []int{32}, ""); err == nil {
		t.Fatal("expected an error for a monthly day out of the 1-31 (or 999) range")
	}
}

func TestBuildFrequency_TZDefaultsWhenUnset(t *testing.T) {
	freq, err := buildFrequency("daily", "2026-09-20", "", nil, "")
	if err != nil {
		t.Fatalf("buildFrequency(daily): %v", err)
	}
	if freq.TimeZoneOffset == nil || *freq.TimeZoneOffset == "" {
		t.Error("expected TimeZoneOffset to default to the local machine's offset when unset")
	}
}

func TestBuildFrequency_TZExplicitIsPreserved(t *testing.T) {
	freq, err := buildFrequency("daily", "2026-09-20", "-05:00", nil, "")
	if err != nil {
		t.Fatalf("buildFrequency(daily): %v", err)
	}
	if freq.TimeZoneOffset == nil || *freq.TimeZoneOffset != "-05:00" {
		t.Errorf("TimeZoneOffset = %v, want -05:00", freq.TimeZoneOffset)
	}
}

func TestParseStartDate(t *testing.T) {
	cases := []struct {
		in      string
		wantErr bool
	}{
		{"2026-09-20", false},
		{"2026-09-20 02:00", false},
		{"2026-09-20T02:00", false},
		{"not-a-date", true},
		{"", true},
	}
	for _, c := range cases {
		_, err := parseStartDate(c.in)
		if (err != nil) != c.wantErr {
			t.Errorf("parseStartDate(%q): err = %v, wantErr = %v", c.in, err, c.wantErr)
		}
	}
}

func TestParseStartDate_RoundTripsToExpectedInstant(t *testing.T) {
	ms, err := parseStartDate("2026-09-20T02:00")
	if err != nil {
		t.Fatalf("parseStartDate: %v", err)
	}
	want := time.Date(2026, 9, 20, 2, 0, 0, 0, time.Local).UnixMilli()
	if ms != want {
		t.Errorf("got %d ms, want %d ms", ms, want)
	}
}

func TestNormalizeNotifyCondition(t *testing.T) {
	cases := map[string]string{
		"failure":           "FAILURE",
		"FAILURE":           "FAILURE",
		"success+failure":   "SUCCESS_FAILURE",
		"Success + Failure": "SUCCESS_FAILURE",
		"  failure  ":       "FAILURE",
	}
	for in, want := range cases {
		got, err := normalizeNotifyCondition(in)
		if err != nil {
			t.Fatalf("normalizeNotifyCondition(%q): unexpected error: %v", in, err)
		}
		if got != want {
			t.Errorf("normalizeNotifyCondition(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestNormalizeNotifyCondition_Invalid(t *testing.T) {
	if _, err := normalizeNotifyCondition("success"); err == nil {
		t.Error("expected an error: Nexus has no success-only notification condition")
	}
	if _, err := normalizeNotifyCondition("bogus"); err == nil {
		t.Error("expected an error for a nonsense value")
	}
}

func TestValidSchedules_CoversAllNexusScheduleTypes(t *testing.T) {
	want := []string{"manual", "once", "hourly", "daily", "weekly", "monthly", "cron"}
	for _, s := range want {
		if !validSchedules[s] {
			t.Errorf("validSchedules is missing %q", s)
		}
	}
	if len(validSchedules) != len(want) {
		t.Errorf("validSchedules has %d entries, want %d — did an unexpected value get added?", len(validSchedules), len(want))
	}
}

// sanity check that warnings-only paths (unused flag for a given schedule)
// don't turn into errors.
func TestBuildFrequency_IgnoredFlagsDoNotError(t *testing.T) {
	if _, err := buildFrequency("manual", "2026-09-20", "-05:00", []int{1}, "0 0 * * * ?"); err != nil {
		t.Errorf("manual schedule should ignore, not reject, unrelated flags: %v", err)
	}
	if _, err := buildFrequency("cron", "", "", []int{1}, "0 0 * * * ?"); err != nil {
		t.Errorf("cron schedule should ignore, not reject, --days: %v", err)
	}
}
