package service

import (
	"encoding/json"
	"fmt"
	"time"

	"studypilot/server/internal/model"
)

type ReviewParams struct {
	Title         *string
	ContentMD     *string
	StatsSnapshot *json.RawMessage
}

func (s *Service) ListReviews() ([]model.Review, error) {
	var rs []model.Review
	if err := s.db.Order("week_start DESC").Find(&rs).Error; err != nil {
		return nil, err
	}
	return rs, nil
}

func (s *Service) GetReview(weekStart string) (*model.Review, error) {
	var r model.Review
	if err := s.db.Where("week_start = ?", weekStart).First(&r).Error; err != nil {
		return nil, ErrNotFound
	}
	return &r, nil
}

// UpsertReview 按周创建或更新复盘（R-02/R-04：快照随保存冻结）。
func (s *Service) UpsertReview(weekStart string, p ReviewParams) (*model.Review, error) {
	if !validDate(weekStart) || !s.isWeekStart(weekStart) {
		return nil, fmt.Errorf("%w: week_start 必须是周一日期", ErrInvalid)
	}
	var r model.Review
	err := s.db.Where("week_start = ?", weekStart).First(&r).Error
	created := false
	if err != nil {
		r = model.Review{WeekStart: weekStart}
		created = true
	}
	if p.Title != nil {
		r.Title = *p.Title
	}
	if p.ContentMD != nil {
		r.ContentMD = *p.ContentMD
	}
	if p.StatsSnapshot != nil {
		r.StatsSnapshot = *p.StatsSnapshot
	}
	if created {
		if r.StatsSnapshot == nil {
			r.StatsSnapshot = json.RawMessage(`{}`)
		}
		if e := s.db.Create(&r).Error; e != nil {
			return nil, e
		}
	} else {
		if e := s.db.Save(&r).Error; e != nil {
			return nil, e
		}
	}
	return &r, nil
}

func (s *Service) DeleteReview(weekStart string) error {
	if _, err := s.GetReview(weekStart); err != nil {
		return err
	}
	return s.db.Where("week_start = ?", weekStart).Delete(&model.Review{}).Error
}

// ComputeSnapshot 生成周统计快照（不落库，由复盘保存时冻结）。
func (s *Service) ComputeSnapshot(weekStart string) (*StatsSnapshot, error) {
	w, err := s.WeeklyStats(weekStart)
	if err != nil {
		return nil, err
	}
	we, _ := weekEndDate(weekStart)
	// 打卡统计
	var checkinTotal int64
	if err := s.db.Model(&model.GoalCheckin{}).
		Joins("JOIN goals ON goals.id = goal_checkins.goal_id").
		Where("goals.goal_type = ? AND goals.week_start = ? AND goal_checkins.checkin_date BETWEEN ? AND ?",
			model.GoalTypeWeek, weekStart, weekStart, we).
		Count(&checkinTotal).Error; err != nil {
		return nil, err
	}
	var checkinDaysTotal int64
	if err := s.db.Model(&model.Goal{}).
		Where("goal_type = ? AND goal_mode = ? AND week_start = ? AND status <> ?",
			model.GoalTypeWeek, model.GoalModeDaily, weekStart, model.GoalStatusArchived).
		Select("COALESCE(SUM(checkin_days), 0)").Scan(&checkinDaysTotal).Error; err != nil {
		return nil, err
	}
	snap := StatsSnapshot{
		GeneratedAt:     time.Now().Format(dateLayout + " 15:04:05"),
		WeekStart:       weekStart,
		WeekEnd:         we,
		TaskRate:        w.TaskCompletionRate,
		GoalRate:        w.GoalAchievementRate,
		TaskTotal:       w.TaskTotal,
		TaskDone:        w.TaskDone,
		OverdueCount:    w.OverdueCount,
		GoalTotal:       w.GoalTotal,
		GoalDone:        w.GoalDone,
		CheckinTotal:    checkinTotal,
		CheckinDays:     checkinDaysTotal,
		ScheduledMin:    w.ScheduledMinutes,
		CompletedMin:    w.CompletedBlockMin,
		UnfinishedTasks: w.UnfinishedTitles,
		ByTag:           w.ByTag,
	}
	return &snap, nil
}
