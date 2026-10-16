package service

import (
	"fmt"

	"studypilot/server/internal/model"
	"studypilot/server/internal/store"
)

// ---- DTO（与 docs/openapi.yaml 对齐） ----

type TrendPoint struct {
	WeekStart string   `json:"week_start"`
	TaskRate  *float64 `json:"task_rate"`
	GoalRate  *float64 `json:"goal_rate"`
}

type DistItem struct {
	Label string `json:"label"`
	Done  int64  `json:"done"`
	Total int64  `json:"total"`
}

type BlockView struct {
	model.Block
	Done bool `json:"done"`
}

type StatsDaily struct {
	Date                string   `json:"date"`
	TaskCompletionRate  *float64 `json:"task_completion_rate"`
	GoalAchievementRate *float64 `json:"goal_achievement_rate"`
	ScheduledMinutes    int64    `json:"scheduled_minutes"`
	CompletedBlockMin   int64    `json:"completed_block_minutes"`
	OverdueCount        int64    `json:"overdue_count"`
	TaskTotal           int64    `json:"task_total"`
	TaskDone            int64    `json:"task_done"`
	GoalTotal           int64    `json:"goal_total"`
	GoalDone            int64    `json:"goal_done"`
}

type StatsWeekly struct {
	WeekStart           string       `json:"week_start"`
	WeekEnd             string       `json:"week_end"`
	TaskCompletionRate  *float64     `json:"task_completion_rate"`
	GoalAchievementRate *float64     `json:"goal_achievement_rate"`
	TaskTotal           int64        `json:"task_total"`
	TaskDone            int64        `json:"task_done"`
	OverdueCount        int64        `json:"overdue_count"`
	GoalTotal           int64        `json:"goal_total"`
	GoalDone            int64        `json:"goal_done"`
	ScheduledMinutes    int64        `json:"scheduled_minutes"`
	CompletedBlockMin   int64        `json:"completed_block_minutes"`
	Trend               []TrendPoint `json:"trend"`
	ByGoal              []DistItem   `json:"by_goal"`
	ByTag               []DistItem   `json:"by_tag"`
	Unfinished          []model.Task `json:"unfinished"`
	UnfinishedTitles    []string     `json:"-"`
}

type HeatmapCell struct {
	Date      string   `json:"date"`
	Score     *float64 `json:"score"`
	TaskRate  *float64 `json:"task_rate"`
	GoalRate  *float64 `json:"goal_rate"`
	TaskTotal int64    `json:"task_total"`
	TaskDone  int64    `json:"task_done"`
	GoalTotal int64    `json:"goal_total"`
	GoalDone  int64    `json:"goal_done"`
}

type Heatmap struct {
	Start string        `json:"start"`
	End   string        `json:"end"`
	Cells []HeatmapCell `json:"cells"`
}

type StatsSnapshot struct {
	GeneratedAt     string     `json:"generated_at"`
	WeekStart       string     `json:"week_start"`
	WeekEnd         string     `json:"week_end"`
	TaskRate        *float64   `json:"task_completion_rate"`
	GoalRate        *float64   `json:"goal_achievement_rate"`
	TaskTotal       int64      `json:"task_total"`
	TaskDone        int64      `json:"task_done"`
	OverdueCount    int64      `json:"overdue_count"`
	GoalTotal       int64      `json:"goal_total"`
	GoalDone        int64      `json:"goal_done"`
	CheckinTotal    int64      `json:"checkin_total"`
	CheckinDays     int64      `json:"checkin_days_total"`
	ScheduledMin    int64      `json:"scheduled_minutes"`
	CompletedMin    int64      `json:"completed_block_minutes"`
	UnfinishedTasks []string   `json:"unfinished_tasks"`
	ByTag           []DistItem `json:"by_tag"`
}

type Dashboard struct {
	Date                string       `json:"date"`
	TaskCompletionRate  *float64     `json:"task_completion_rate"`
	GoalAchievementRate *float64     `json:"goal_achievement_rate"`
	ScheduledMinutes    int64        `json:"scheduled_minutes"`
	OverdueCount        int64        `json:"overdue_count"`
	DayGoals            []model.Goal `json:"day_goals"`
	TasksDue            []model.Task `json:"tasks_due"`
	Blocks              []BlockView  `json:"blocks"`
}

// ---- 工具 ----

func rate(total, done int64) *float64 {
	if total <= 0 {
		return nil
	}
	r := float64(done) / float64(total)
	return &r
}

// ---- 日统计 ----

func (s *Service) DailyStats(date string) (*StatsDaily, error) {
	if !validDate(date) {
		return nil, fmt.Errorf("%w: date 格式非法", ErrInvalid)
	}
	out := &StatsDaily{Date: date}
	// 任务口径：due_date = 当日
	if err := s.db.Model(&model.Task{}).Where("due_date = ?", date).Count(&out.TaskTotal).Error; err != nil {
		return nil, err
	}
	if err := s.db.Model(&model.Task{}).Where("due_date = ? AND status = ?", date, model.TaskStatusDone).Count(&out.TaskDone).Error; err != nil {
		return nil, err
	}
	out.TaskCompletionRate = rate(out.TaskTotal, out.TaskDone)
	// 目标口径：当日日目标（排除归档）的打卡率
	if err := s.db.Model(&model.Goal{}).Where("goal_type = ? AND date = ? AND status <> ?", model.GoalTypeDay, date, model.GoalStatusArchived).Count(&out.GoalTotal).Error; err != nil {
		return nil, err
	}
	if err := s.db.Model(&model.GoalCheckin{}).
		Joins("JOIN goals ON goals.id = goal_checkins.goal_id").
		Where("goals.goal_type = ? AND goals.date = ? AND goals.status <> ?", model.GoalTypeDay, date, model.GoalStatusArchived).
		Distinct("goal_checkins.id").Count(&out.GoalDone).Error; err != nil {
		return nil, err
	}
	out.GoalAchievementRate = rate(out.GoalTotal, out.GoalDone)
	// 日程口径
	if err := s.db.Model(&model.Block{}).Where("block_date = ?", date).
		Select("COALESCE(SUM(end_min - start_min), 0)").Scan(&out.ScheduledMinutes).Error; err != nil {
		return nil, err
	}
	if err := s.db.Model(&model.Block{}).
		Joins("JOIN tasks ON tasks.id = blocks.task_id").
		Where("blocks.block_date = ? AND tasks.status = ?", date, model.TaskStatusDone).
		Select("COALESCE(SUM(blocks.end_min - blocks.start_min), 0)").Scan(&out.CompletedBlockMin).Error; err != nil {
		return nil, err
	}
	// 逾期：以查询日为准（due < 查询日 且未完成）
	if err := s.db.Model(&model.Task{}).
		Where("due_date IS NOT NULL AND due_date < ? AND status <> ?", date, model.TaskStatusDone).
		Count(&out.OverdueCount).Error; err != nil {
		return nil, err
	}
	return out, nil
}

// ---- 周统计 ----

func (s *Service) weekTaskStats(weekStart string) (total, done, overdue int64, err error) {
	we, err := weekEndDate(weekStart)
	if err != nil {
		return 0, 0, 0, err
	}
	if err := s.db.Model(&model.Task{}).Where("due_date BETWEEN ? AND ?", weekStart, we).Count(&total).Error; err != nil {
		return 0, 0, 0, err
	}
	if err := s.db.Model(&model.Task{}).Where("due_date BETWEEN ? AND ? AND status = ?", weekStart, we, model.TaskStatusDone).Count(&done).Error; err != nil {
		return 0, 0, 0, err
	}
	if err := s.db.Model(&model.Task{}).
		Where("due_date BETWEEN ? AND ? AND due_date < ? AND status <> ?", weekStart, we, today(), model.TaskStatusDone).
		Count(&overdue).Error; err != nil {
		return 0, 0, 0, err
	}
	return total, done, overdue, nil
}

func (s *Service) weekGoalStats(weekStart string) (total, done int64, rateV *float64, err error) {
	var goals []model.Goal
	if err := s.db.Where("goal_type = ? AND week_start = ? AND status <> ?", model.GoalTypeWeek, weekStart, model.GoalStatusArchived).Find(&goals).Error; err != nil {
		return 0, 0, nil, err
	}
	total = int64(len(goals))
	if total == 0 {
		return 0, 0, nil, nil
	}
	we, _ := weekEndDate(weekStart)
	sum := 0.0
	for _, g := range goals {
		if g.Status == model.GoalStatusDone {
			done++
		}
		switch g.GoalMode {
		case model.GoalModeOnce:
			if g.Status == model.GoalStatusDone {
				sum += 1
			}
		default: // daily
			var n int64
			if err := s.db.Model(&model.GoalCheckin{}).
				Where("goal_id = ? AND checkin_date BETWEEN ? AND ?", g.ID, weekStart, we).
				Count(&n).Error; err != nil {
				return 0, 0, nil, err
			}
			r := float64(n) / float64(g.CheckinDays)
			if r > 1 {
				r = 1
			}
			sum += r
		}
	}
	avg := sum / float64(total)
	return total, done, &avg, nil
}

func (s *Service) WeeklyStats(weekStart string) (*StatsWeekly, error) {
	if !validDate(weekStart) || !s.isWeekStart(weekStart) {
		return nil, fmt.Errorf("%w: week_start 必须是周一日期", ErrInvalid)
	}
	we, _ := weekEndDate(weekStart)
	out := &StatsWeekly{WeekStart: weekStart, WeekEnd: we}

	total, done, overdue, err := s.weekTaskStats(weekStart)
	if err != nil {
		return nil, err
	}
	out.TaskTotal, out.TaskDone, out.OverdueCount = total, done, overdue
	out.TaskCompletionRate = rate(total, done)

	gt, gd, gRate, err := s.weekGoalStats(weekStart)
	if err != nil {
		return nil, err
	}
	out.GoalTotal, out.GoalDone, out.GoalAchievementRate = gt, gd, gRate

	if err := s.db.Model(&model.Block{}).Where("block_date BETWEEN ? AND ?", weekStart, we).
		Select("COALESCE(SUM(end_min - start_min), 0)").Scan(&out.ScheduledMinutes).Error; err != nil {
		return nil, err
	}
	if err := s.db.Model(&model.Block{}).
		Joins("JOIN tasks ON tasks.id = blocks.task_id").
		Where("blocks.block_date BETWEEN ? AND ? AND tasks.status = ?", weekStart, we, model.TaskStatusDone).
		Select("COALESCE(SUM(blocks.end_min - blocks.start_min), 0)").Scan(&out.CompletedBlockMin).Error; err != nil {
		return nil, err
	}

	// 近 8 周趋势（旧 → 新）
	start, _ := parseDate(weekStart)
	trendStart := start.AddDate(0, 0, -7*7)
	out.Trend = make([]TrendPoint, 0, 8)
	for i := 0; i < 8; i++ {
		ws := trendStart.AddDate(0, 0, 7*i).Format(dateLayout)
		tt, td, _, err := s.weekTaskStats(ws)
		if err != nil {
			return nil, err
		}
		_, _, gr, err := s.weekGoalStats(ws)
		if err != nil {
			return nil, err
		}
		out.Trend = append(out.Trend, TrendPoint{WeekStart: ws, TaskRate: rate(tt, td), GoalRate: gr})
	}

	// 任务分布：按目标 / 按标签 / 未完成
	var weekTasks []model.Task
	if err := s.db.Where("due_date BETWEEN ? AND ?", weekStart, we).Find(&weekTasks).Error; err != nil {
		return nil, err
	}
	goalLabel := map[int64]string{}
	for _, t := range weekTasks {
		if t.GoalID != nil {
			if _, ok := goalLabel[*t.GoalID]; !ok {
				var g model.Goal
				if err := s.db.Select("title").First(&g, *t.GoalID).Error; err == nil {
					goalLabel[*t.GoalID] = g.Title
				}
			}
		}
	}
	byGoal := map[string]*DistItem{}
	byTag := map[string]*DistItem{}
	for _, t := range weekTasks {
		key := "未关联目标"
		if t.GoalID != nil {
			key = goalLabel[*t.GoalID]
		}
		it := byGoal[key]
		if it == nil {
			it = &DistItem{Label: key}
			byGoal[key] = it
		}
		it.Total++
		if t.Status == model.TaskStatusDone {
			it.Done++
		}
		for _, tag := range t.Tags {
			ti := byTag[tag]
			if ti == nil {
				ti = &DistItem{Label: tag}
				byTag[tag] = ti
			}
			ti.Total++
			if t.Status == model.TaskStatusDone {
				ti.Done++
			}
		}
	}
	for _, it := range byGoal {
		out.ByGoal = append(out.ByGoal, *it)
	}
	for _, it := range byTag {
		out.ByTag = append(out.ByTag, *it)
	}
	for _, t := range weekTasks {
		if t.Status != model.TaskStatusDone {
			out.Unfinished = append(out.Unfinished, t)
			out.UnfinishedTitles = append(out.UnfinishedTitles, t.Title)
		}
	}
	return out, nil
}

// ---- 热力图 ----

func (s *Service) Heatmap(start, end string) (*Heatmap, error) {
	if !validDate(start) || !validDate(end) {
		return nil, fmt.Errorf("%w: start/end 格式非法", ErrInvalid)
	}
	sd, _ := parseDate(start)
	ed, _ := parseDate(end)
	if ed.Before(sd) {
		return nil, fmt.Errorf("%w: end 必须不早于 start", ErrInvalid)
	}
	days := int(ed.Sub(sd).Hours()/24) + 1
	if days > 366 {
		return nil, fmt.Errorf("%w: 热力图范围不能超过 366 天", ErrInvalid)
	}
	taskW := store.GetSettingFloat(s.db, "heatmap_task_weight", 0.5)
	goalW := store.GetSettingFloat(s.db, "heatmap_goal_weight", 0.5)
	out := &Heatmap{Start: start, End: end, Cells: make([]HeatmapCell, 0, days)}
	for i := 0; i < days; i++ {
		date := sd.AddDate(0, 0, i).Format(dateLayout)
		cell := HeatmapCell{Date: date}
		if err := s.db.Model(&model.Task{}).Where("due_date = ?", date).Count(&cell.TaskTotal).Error; err != nil {
			return nil, err
		}
		if err := s.db.Model(&model.Task{}).Where("due_date = ? AND status = ?", date, model.TaskStatusDone).Count(&cell.TaskDone).Error; err != nil {
			return nil, err
		}
		cell.TaskRate = rate(cell.TaskTotal, cell.TaskDone)
		if err := s.db.Model(&model.Goal{}).Where("goal_type = ? AND date = ? AND status <> ?", model.GoalTypeDay, date, model.GoalStatusArchived).Count(&cell.GoalTotal).Error; err != nil {
			return nil, err
		}
		if err := s.db.Model(&model.GoalCheckin{}).
			Joins("JOIN goals ON goals.id = goal_checkins.goal_id").
			Where("goals.goal_type = ? AND goals.date = ? AND goals.status <> ?", model.GoalTypeDay, date, model.GoalStatusArchived).
			Distinct("goal_checkins.id").Count(&cell.GoalDone).Error; err != nil {
			return nil, err
		}
		cell.GoalRate = rate(cell.GoalTotal, cell.GoalDone)
		// 综合分：可用口径的加权平均；两者皆无则 null
		if cell.TaskRate != nil || cell.GoalRate != nil {
			var sum, wsum float64
			if cell.TaskRate != nil {
				sum += taskW * *cell.TaskRate
				wsum += taskW
			}
			if cell.GoalRate != nil {
				sum += goalW * *cell.GoalRate
				wsum += goalW
			}
			score := sum / wsum
			cell.Score = &score
		}
		out.Cells = append(out.Cells, cell)
	}
	return out, nil
}

// ---- 今日面板 ----

func (s *Service) Dashboard(date string) (*Dashboard, error) {
	d, err := s.DailyStats(date)
	if err != nil {
		return nil, err
	}
	out := &Dashboard{
		Date:                date,
		TaskCompletionRate:  d.TaskCompletionRate,
		GoalAchievementRate: d.GoalAchievementRate,
		ScheduledMinutes:    d.ScheduledMinutes,
		OverdueCount:        d.OverdueCount,
	}
	if err := s.db.Where("goal_type = ? AND date = ? AND status <> ?", model.GoalTypeDay, date, model.GoalStatusArchived).
		Order("sort ASC, id ASC").Find(&out.DayGoals).Error; err != nil {
		return nil, err
	}
	if err := s.db.Where("due_date = ?", date).Order("priority ASC, sort ASC, id ASC").Find(&out.TasksDue).Error; err != nil {
		return nil, err
	}
	var blocks []model.Block
	if err := s.db.Where("block_date = ?", date).Order("start_min ASC, id ASC").Find(&blocks).Error; err != nil {
		return nil, err
	}
	for _, b := range blocks {
		done := false
		if b.TaskID != nil {
			var t model.Task
			if err := s.db.Select("status").First(&t, *b.TaskID).Error; err == nil && t.Status == model.TaskStatusDone {
				done = true
			}
		}
		out.Blocks = append(out.Blocks, BlockView{Block: b, Done: done})
	}
	return out, nil
}
