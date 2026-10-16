package service

import (
	"encoding/json"
	"path/filepath"
	"testing"
	"time"

	"studypilot/server/internal/store"
)

func newTestService(t *testing.T) *Service {
	t.Helper()
	db, err := store.Open(filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	if err := store.Migrate(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("sql db: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	return New(db)
}

// currentMonday 返回本周周一（本地日期）。
func currentMonday() string {
	d := time.Now()
	for d.Weekday() != time.Monday {
		d = d.AddDate(0, 0, -1)
	}
	return d.Format(dateLayout)
}

func ptrJSON(v any) *json.RawMessage {
	b, _ := json.Marshal(v)
	rm := json.RawMessage(b)
	return &rm
}

func addDays(date string, days int) string {
	d, _ := parseDate(date)
	return d.AddDate(0, 0, days).Format(dateLayout)
}

func mkGoal(t *testing.T, s *Service, title, goalType, goalMode string, date, weekStart *string, checkinDays int) int64 {
	t.Helper()
	g, err := s.CreateGoal(GoalParams{
		Title:       title,
		GoalType:    goalType,
		GoalMode:    goalMode,
		Date:        date,
		WeekStart:   weekStart,
		CheckinDays: &checkinDays,
	})
	if err != nil {
		t.Fatalf("create goal: %v", err)
	}
	return g.ID
}

func mkTask(t *testing.T, s *Service, title string, due *string) int64 {
	t.Helper()
	task, err := s.CreateTask(TaskParams{Title: title, DueDate: due})
	if err != nil {
		t.Fatalf("create task: %v", err)
	}
	return task.ID
}
