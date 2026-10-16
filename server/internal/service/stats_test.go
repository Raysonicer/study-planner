package service

import (
	"testing"
)

func TestDailyStatsRates(t *testing.T) {
	s := newTestService(t)
	d := today()

	// 任务：2 个 due 今日，完成 1 个
	a := mkTask(t, s, "任务A", &d)
	b := mkTask(t, s, "任务B", &d)
	if _, err := s.CompleteTask(a); err != nil {
		t.Fatalf("complete A: %v", err)
	}
	_ = b
	// 日目标：2 个，打卡 1 个
	g1 := mkGoal(t, s, "目标1", "day", "daily", &d, nil, 5)
	mkGoal(t, s, "目标2", "day", "daily", &d, nil, 5)
	if _, err := s.Checkin(g1, ""); err != nil {
		t.Fatalf("checkin: %v", err)
	}
	// 日程：1 小时块
	if _, err := s.CreateBlock(BlockParams{BlockDate: d, StartMin: 1200, EndMin: 1260, Title: "学习", TaskID: &a}); err != nil {
		t.Fatalf("block: %v", err)
	}

	st, err := s.DailyStats(d)
	if err != nil {
		t.Fatalf("daily: %v", err)
	}
	if st.TaskTotal != 2 || st.TaskDone != 1 {
		t.Fatalf("task total/done = %d/%d, want 2/1", st.TaskTotal, st.TaskDone)
	}
	if st.TaskCompletionRate == nil || *st.TaskCompletionRate != 0.5 {
		t.Fatalf("task rate = %v, want 0.5", st.TaskCompletionRate)
	}
	if st.GoalTotal != 2 || st.GoalDone != 1 {
		t.Fatalf("goal total/done = %d/%d, want 2/1", st.GoalTotal, st.GoalDone)
	}
	if st.GoalAchievementRate == nil || *st.GoalAchievementRate != 0.5 {
		t.Fatalf("goal rate = %v, want 0.5", st.GoalAchievementRate)
	}
	if st.ScheduledMinutes != 60 {
		t.Fatalf("scheduled = %d, want 60", st.ScheduledMinutes)
	}
	if st.CompletedBlockMin != 60 {
		t.Fatalf("completed block min = %d, want 60（关联任务已完成）", st.CompletedBlockMin)
	}
}

func TestWeeklyStats(t *testing.T) {
	s := newTestService(t)
	ws := currentMonday()
	// 本周 4 个任务，完成 3 个
	ids := []int64{}
	for i := 0; i < 4; i++ {
		due := addDays(ws, i)
		ids = append(ids, mkTask(t, s, "周任务"+string(rune('A'+i)), &due))
	}
	for i := 0; i < 3; i++ {
		if _, err := s.CompleteTask(ids[i]); err != nil {
			t.Fatalf("complete: %v", err)
		}
	}
	// 打卡型周目标：应打卡 5 天，实际 2 天
	g := mkGoal(t, s, "周目标", "week", "daily", nil, &ws, 5)
	if _, err := s.Checkin(g, ws); err != nil {
		t.Fatalf("checkin 1: %v", err)
	}
	if _, err := s.Checkin(g, addDays(ws, 1)); err != nil {
		t.Fatalf("checkin 2: %v", err)
	}

	w, err := s.WeeklyStats(ws)
	if err != nil {
		t.Fatalf("weekly: %v", err)
	}
	if w.TaskTotal != 4 || w.TaskDone != 3 {
		t.Fatalf("task total/done = %d/%d, want 4/3", w.TaskTotal, w.TaskDone)
	}
	if w.TaskCompletionRate == nil || *w.TaskCompletionRate != 0.75 {
		t.Fatalf("task rate = %v, want 0.75", w.TaskCompletionRate)
	}
	if w.GoalAchievementRate == nil || *w.GoalAchievementRate < 0.39 || *w.GoalAchievementRate > 0.41 {
		t.Fatalf("goal rate = %v, want ~0.4", w.GoalAchievementRate)
	}
	if len(w.Trend) != 8 {
		t.Fatalf("trend len = %d, want 8", len(w.Trend))
	}
	if len(w.Unfinished) != 1 {
		t.Fatalf("unfinished = %d, want 1", len(w.Unfinished))
	}
}

func TestHeatmapWithWeights(t *testing.T) {
	s := newTestService(t)
	d := today()
	a := mkTask(t, s, "A", &d)
	mkTask(t, s, "B", &d)
	if _, err := s.CompleteTask(a); err != nil {
		t.Fatalf("complete: %v", err)
	}
	g := mkGoal(t, s, "G", "day", "daily", &d, nil, 5)
	if _, err := s.Checkin(g, ""); err != nil {
		t.Fatalf("checkin: %v", err)
	}

	// 默认 0.5/0.5：task 0.5、goal 1.0 → score 0.75
	hm, err := s.Heatmap(d, d)
	if err != nil {
		t.Fatalf("heatmap: %v", err)
	}
	if len(hm.Cells) != 1 {
		t.Fatalf("cells = %d, want 1", len(hm.Cells))
	}
	cell := hm.Cells[0]
	if cell.Score == nil || *cell.Score < 0.74 || *cell.Score > 0.76 {
		t.Fatalf("score = %v, want ~0.75", cell.Score)
	}
	// 权重调成 1/0 → score = task rate = 0.5
	if err := s.SetSetting("heatmap_task_weight", "1"); err != nil {
		t.Fatalf("set weight: %v", err)
	}
	if err := s.SetSetting("heatmap_goal_weight", "0"); err != nil {
		t.Fatalf("set weight: %v", err)
	}
	hm2, err := s.Heatmap(d, d)
	if err != nil {
		t.Fatalf("heatmap2: %v", err)
	}
	if hm2.Cells[0].Score == nil || *hm2.Cells[0].Score != 0.5 {
		t.Fatalf("weighted score = %v, want 0.5", hm2.Cells[0].Score)
	}
}

func TestOverdue(t *testing.T) {
	s := newTestService(t)
	due := addDays(today(), -1)
	mkTask(t, s, "逾期任务", &due)
	st, err := s.DailyStats(today())
	if err != nil {
		t.Fatalf("daily: %v", err)
	}
	if st.OverdueCount != 1 {
		t.Fatalf("overdue = %d, want 1", st.OverdueCount)
	}
}

func TestDashboard(t *testing.T) {
	s := newTestService(t)
	d := today()
	mkGoal(t, s, "目标", "day", "daily", &d, nil, 5)
	task := mkTask(t, s, "任务", &d)
	if _, err := s.CreateBlock(BlockParams{BlockDate: d, StartMin: 540, EndMin: 600, Title: "晨读", TaskID: &task}); err != nil {
		t.Fatalf("block: %v", err)
	}
	dash, err := s.Dashboard(d)
	if err != nil {
		t.Fatalf("dashboard: %v", err)
	}
	if len(dash.DayGoals) != 1 || len(dash.TasksDue) != 1 || len(dash.Blocks) != 1 {
		t.Fatalf("lists = %d/%d/%d, want 1/1/1", len(dash.DayGoals), len(dash.TasksDue), len(dash.Blocks))
	}
	if dash.Blocks[0].Done {
		t.Fatalf("block done should be false（任务未完成）")
	}
	if _, err := s.CompleteTask(task); err != nil {
		t.Fatalf("complete: %v", err)
	}
	dash2, _ := s.Dashboard(d)
	if !dash2.Blocks[0].Done {
		t.Fatalf("block done should be true（任务已完成）")
	}
}
