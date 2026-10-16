package service

import (
	"errors"
	"testing"
)

func TestExportImportRoundTrip(t *testing.T) {
	s := newTestService(t)
	ws := currentMonday()
	d := today()

	g1 := mkGoal(t, s, "日目标", "day", "daily", &d, nil, 5)
	g2 := mkGoal(t, s, "周目标", "week", "daily", nil, &ws, 5)
	if _, err := s.Checkin(g1, ""); err != nil {
		t.Fatalf("checkin: %v", err)
	}
	task := mkTask(t, s, "任务", &d)
	if _, err := s.CompleteTask(task); err != nil {
		t.Fatalf("complete: %v", err)
	}
	if _, err := s.CreateBlock(BlockParams{BlockDate: d, StartMin: 600, EndMin: 660, Title: "块", TaskID: &task}); err != nil {
		t.Fatalf("block: %v", err)
	}
	title := "复盘"
	if _, err := s.UpsertReview(ws, ReviewParams{Title: &title}); err != nil {
		t.Fatalf("review: %v", err)
	}
	if err := s.SetSetting("theme", "dark"); err != nil {
		t.Fatalf("setting: %v", err)
	}
	_ = g2

	payload, err := s.Export()
	if err != nil {
		t.Fatalf("export: %v", err)
	}
	if len(payload.Goals) != 2 || len(payload.Checkins) != 1 || len(payload.Tasks) != 1 ||
		len(payload.TaskEvents) < 2 || len(payload.Blocks) != 1 || len(payload.Reviews) != 1 || len(payload.Settings) != 1 {
		t.Fatalf("export counts wrong: %+v", payload)
	}

	if err := s.ImportAll(payload); err != nil {
		t.Fatalf("import: %v", err)
	}
	after, err := s.Export()
	if err != nil {
		t.Fatalf("re-export: %v", err)
	}
	if len(after.Goals) != len(payload.Goals) || len(after.Tasks) != len(payload.Tasks) || len(after.Checkins) != len(payload.Checkins) {
		t.Fatalf("round trip counts mismatch")
	}
	if after.Goals[0].ID != payload.Goals[0].ID {
		t.Fatalf("round trip ids broken")
	}
	if (after.Tasks[0].GoalID == nil) != (payload.Tasks[0].GoalID == nil) {
		t.Fatalf("round trip references broken")
	}
	if _, err := s.GetTask(task); err != nil {
		t.Fatalf("task should survive round trip: %v", err)
	}
	if err := s.ImportAll(&ExportPayload{}); !errors.Is(err, ErrInvalid) {
		t.Fatalf("invalid payload should be rejected")
	}
}

func TestCustomWeekStartDay(t *testing.T) {
	s := newTestService(t)
	// 周日作为周起始
	if err := s.SetSetting("week_start_day", "0"); err != nil {
		t.Fatalf("set: %v", err)
	}
	sunday := currentMonday()
	for sunday != "" && !isSunday(sunday) {
		sunday = addDays(sunday, -1)
	}
	if _, err := s.CreateGoal(GoalParams{Title: "x", GoalType: "week", GoalMode: "daily", WeekStart: &sunday, CheckinDays: ptr(5)}); err != nil {
		t.Fatalf("sunday start should be accepted: %v", err)
	}
	// 周一不再被接受
	mon := addDays(sunday, 1)
	if _, err := s.CreateGoal(GoalParams{Title: "x", GoalType: "week", GoalMode: "daily", WeekStart: &mon, CheckinDays: ptr(5)}); err == nil {
		t.Fatalf("monday start should be rejected when week starts on sunday")
	}
	// 周统计同样按设置校验
	if _, err := s.WeeklyStats(sunday); err != nil {
		t.Fatalf("weekly with sunday start: %v", err)
	}
}

func isSunday(d string) bool {
	t, err := parseDate(d)
	return err == nil && t.Weekday() == 0
}
