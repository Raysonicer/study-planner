package service

import (
	"errors"
	"testing"
)

func TestDayGoalCheckinLifecycle(t *testing.T) {
	s := newTestService(t)
	d := today()
	g := mkGoal(t, s, "背单词", "day", "daily", &d, nil, 5)

	ck, err := s.Checkin(g, "")
	if err != nil {
		t.Fatalf("checkin: %v", err)
	}
	if ck.CheckinDate != d {
		t.Fatalf("checkin date = %q, want %q", ck.CheckinDate, d)
	}
	// 重复打卡 → 409
	if _, err := s.Checkin(g, ""); !errors.Is(err, ErrConflict) {
		t.Fatalf("dup checkin err = %v, want conflict", err)
	}
	// 错误日期 → 400
	if _, err := s.Checkin(g, addDays(d, 1)); !errors.Is(err, ErrInvalid) {
		t.Fatalf("wrong date err = %v, want invalid", err)
	}
	// 撤销后再次打卡成功
	if err := s.Uncheckin(g, ""); err != nil {
		t.Fatalf("uncheckin: %v", err)
	}
	if _, err := s.Checkin(g, ""); err != nil {
		t.Fatalf("re-checkin: %v", err)
	}
}

func TestWeekGoalDailyCheckinBounds(t *testing.T) {
	s := newTestService(t)
	ws := currentMonday()
	g := mkGoal(t, s, "刷题 5 次", "week", "daily", nil, &ws, 5)

	if _, err := s.Checkin(g, ws); err != nil {
		t.Fatalf("monday checkin: %v", err)
	}
	if _, err := s.Checkin(g, addDays(ws, 6)); err != nil {
		t.Fatalf("sunday checkin: %v", err)
	}
	// 下周的周一 → 越界
	if _, err := s.Checkin(g, addDays(ws, 7)); !errors.Is(err, ErrInvalid) {
		t.Fatalf("cross-week err = %v, want invalid", err)
	}
	// 上周日 → 越界
	if _, err := s.Checkin(g, addDays(ws, -1)); !errors.Is(err, ErrInvalid) {
		t.Fatalf("pre-week err = %v, want invalid", err)
	}
	// 撤销必须带日期
	if err := s.Uncheckin(g, ""); !errors.Is(err, ErrInvalid) {
		t.Fatalf("uncheckin without date err = %v, want invalid", err)
	}
	if err := s.Uncheckin(g, ws); err != nil {
		t.Fatalf("uncheckin: %v", err)
	}
}

func TestWeekGoalOnce(t *testing.T) {
	s := newTestService(t)
	ws := currentMonday()
	g := mkGoal(t, s, "完成课程项目", "week", "once", nil, &ws, 5)

	// 一次性目标不允许打卡
	if _, err := s.Checkin(g, ws); !errors.Is(err, ErrInvalid) {
		t.Fatalf("once checkin err = %v, want invalid", err)
	}
	// 通过状态完成
	done := "done"
	upd, err := s.UpdateGoal(g, GoalParams{Title: "完成课程项目", GoalType: "week", GoalMode: "once", WeekStart: &ws, CheckinDays: ptr(5), Status: &done})
	if err != nil {
		t.Fatalf("complete once goal: %v", err)
	}
	if upd.CompletedAt == nil {
		t.Fatalf("completed_at should be set")
	}
	// 撤销完成
	active := "active"
	upd2, err := s.UpdateGoal(g, GoalParams{Title: "完成课程项目", GoalType: "week", GoalMode: "once", WeekStart: &ws, CheckinDays: ptr(5), Status: &active})
	if err != nil {
		t.Fatalf("reopen once goal: %v", err)
	}
	if upd2.CompletedAt != nil {
		t.Fatalf("completed_at should be cleared")
	}
}

func TestListCheckinsRange(t *testing.T) {
	s := newTestService(t)
	ws := currentMonday()
	g := mkGoal(t, s, "周目标", "week", "daily", nil, &ws, 5)
	for _, d := range []string{ws, addDays(ws, 2), addDays(ws, 5)} {
		if _, err := s.Checkin(g, d); err != nil {
			t.Fatalf("checkin %s: %v", d, err)
		}
	}
	list, err := s.ListCheckins(addDays(ws, 1), addDays(ws, 4))
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list) != 1 || list[0].CheckinDate != addDays(ws, 2) {
		t.Fatalf("range checkins = %+v, want 仅第 3 天", list)
	}
	if _, err := s.ListCheckins("bad", ws); err == nil {
		t.Fatalf("bad range should fail")
	}
}

func TestGoalValidation(t *testing.T) {
	s := newTestService(t)
	ws := currentMonday()
	// 周三不是周一
	bad := addDays(ws, 2)
	if _, err := s.CreateGoal(GoalParams{Title: "x", GoalType: "week", GoalMode: "daily", WeekStart: &bad, CheckinDays: ptr(5)}); !errors.Is(err, ErrInvalid) {
		t.Fatalf("non-monday week_start err = %v, want invalid", err)
	}
	// 日目标不能有 week_start
	d := today()
	if _, err := s.CreateGoal(GoalParams{Title: "x", GoalType: "day", GoalMode: "daily", Date: &d, WeekStart: &ws, CheckinDays: ptr(5)}); !errors.Is(err, ErrInvalid) {
		t.Fatalf("day goal with week_start err = %v, want invalid", err)
	}
	// 日目标不支持 once
	if _, err := s.CreateGoal(GoalParams{Title: "x", GoalType: "day", GoalMode: "once", Date: &d, CheckinDays: ptr(5)}); !errors.Is(err, ErrInvalid) {
		t.Fatalf("day once err = %v, want invalid", err)
	}
	// checkin_days 越界
	badDays := 8
	if _, err := s.CreateGoal(GoalParams{Title: "x", GoalType: "week", GoalMode: "daily", WeekStart: &ws, CheckinDays: &badDays}); !errors.Is(err, ErrInvalid) {
		t.Fatalf("checkin_days=8 err = %v, want invalid", err)
	}
	// 归档目标不能打卡
	g := mkGoal(t, s, "归档", "week", "daily", nil, &ws, 5)
	arch := "archived"
	if _, err := s.UpdateGoal(g, GoalParams{Title: "归档", GoalType: "week", GoalMode: "daily", WeekStart: &ws, CheckinDays: ptr(5), Status: &arch}); err != nil {
		t.Fatalf("archive: %v", err)
	}
	if _, err := s.Checkin(g, ws); !errors.Is(err, ErrInvalid) {
		t.Fatalf("archived checkin err = %v, want invalid", err)
	}
}
