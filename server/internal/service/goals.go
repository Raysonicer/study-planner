package service

import (
	"fmt"

	"studypilot/server/internal/model"
)

type GoalFilter struct {
	GoalType  string
	Date      string
	WeekStart string
	Status    string
}

type GoalParams struct {
	Title       string
	GoalType    string
	GoalMode    string
	Date        *string
	WeekStart   *string
	CheckinDays *int
	Status      *string
	Note        *string
	Sort        *int
}

func (s *Service) ListGoals(f GoalFilter) ([]model.Goal, error) {
	q := s.db.Order("sort ASC, id ASC")
	if f.GoalType != "" {
		q = q.Where("goal_type = ?", f.GoalType)
	}
	if f.Date != "" {
		q = q.Where("date = ?", f.Date)
	}
	if f.WeekStart != "" {
		q = q.Where("week_start = ?", f.WeekStart)
	}
	switch f.Status {
	case "":
		q = q.Where("status <> ?", model.GoalStatusArchived)
	default:
		q = q.Where("status = ?", f.Status)
	}
	var goals []model.Goal
	if err := q.Find(&goals).Error; err != nil {
		return nil, err
	}
	return goals, nil
}

func (s *Service) GetGoal(id int64) (*model.Goal, error) {
	var g model.Goal
	if err := s.db.First(&g, id).Error; err != nil {
		return nil, ErrNotFound
	}
	return &g, nil
}

// validateGoal 校验目标的最终状态一致性（周起始日以设置项为准）。
func (s *Service) validateGoal(p *GoalParams) error {
	if p.Title == "" {
		return fmt.Errorf("%w: title 不能为空", ErrInvalid)
	}
	switch p.GoalType {
	case model.GoalTypeDay:
		if p.Date == nil || !validDate(*p.Date) {
			return fmt.Errorf("%w: 日目标需要合法的 date", ErrInvalid)
		}
		if p.WeekStart != nil && *p.WeekStart != "" {
			return fmt.Errorf("%w: 日目标不能设置 week_start", ErrInvalid)
		}
	case model.GoalTypeWeek:
		if p.WeekStart == nil || !validDate(*p.WeekStart) {
			return fmt.Errorf("%w: 周目标需要合法的 week_start", ErrInvalid)
		}
		if !s.isWeekStart(*p.WeekStart) {
			return fmt.Errorf("%w: week_start 必须是周一", ErrInvalid)
		}
		if p.Date != nil && *p.Date != "" {
			return fmt.Errorf("%w: 周目标不能设置 date", ErrInvalid)
		}
	default:
		return fmt.Errorf("%w: goal_type 必须是 day/week", ErrInvalid)
	}
	switch p.GoalMode {
	case model.GoalModeDaily:
	case model.GoalModeOnce:
		if p.GoalType != model.GoalTypeWeek {
			return fmt.Errorf("%w: 一次性模式仅周目标可用", ErrInvalid)
		}
	default:
		return fmt.Errorf("%w: goal_mode 必须是 daily/once", ErrInvalid)
	}
	if p.CheckinDays != nil && (*p.CheckinDays < 1 || *p.CheckinDays > 7) {
		return fmt.Errorf("%w: checkin_days 必须在 1-7", ErrInvalid)
	}
	if p.Status != nil {
		switch *p.Status {
		case model.GoalStatusActive, model.GoalStatusDone, model.GoalStatusArchived:
		default:
			return fmt.Errorf("%w: 非法 status", ErrInvalid)
		}
	}
	return nil
}

func applyGoal(p *GoalParams, g *model.Goal) {
	g.Title = p.Title
	g.GoalType = p.GoalType
	g.GoalMode = p.GoalMode
	g.Date = p.Date
	g.WeekStart = p.WeekStart
	if p.CheckinDays != nil {
		g.CheckinDays = *p.CheckinDays
	}
	if p.Note != nil {
		g.Note = *p.Note
	}
	if p.Sort != nil {
		g.Sort = *p.Sort
	}
	if p.Status != nil {
		g.Status = *p.Status
	}
	// 完成/未完成状态与 completed_at 保持一致
	if g.Status == model.GoalStatusDone {
		if g.CompletedAt == nil {
			g.CompletedAt = ptr(today())
		}
	} else {
		g.CompletedAt = nil
	}
}

func (s *Service) CreateGoal(p GoalParams) (*model.Goal, error) {
	if p.GoalMode == "" {
		p.GoalMode = model.GoalModeDaily
	}
	if p.CheckinDays == nil {
		p.CheckinDays = ptr(5)
	}
	if err := s.validateGoal(&p); err != nil {
		return nil, err
	}
	g := model.Goal{Status: model.GoalStatusActive}
	applyGoal(&p, &g)
	if err := s.db.Create(&g).Error; err != nil {
		return nil, err
	}
	return &g, nil
}

func (s *Service) UpdateGoal(id int64, p GoalParams) (*model.Goal, error) {
	g, err := s.GetGoal(id)
	if err != nil {
		return nil, err
	}
	if p.GoalMode == "" {
		p.GoalMode = g.GoalMode
	}
	if p.CheckinDays == nil {
		p.CheckinDays = ptr(g.CheckinDays)
	}
	if p.Status == nil {
		p.Status = ptr(g.Status)
	}
	if err := s.validateGoal(&p); err != nil {
		return nil, err
	}
	applyGoal(&p, g)
	if err := s.db.Save(g).Error; err != nil {
		return nil, err
	}
	return g, nil
}

func (s *Service) DeleteGoal(id int64) error {
	if _, err := s.GetGoal(id); err != nil {
		return err
	}
	return s.db.Delete(&model.Goal{}, id).Error
}

// Checkin 打卡。日目标按自身日期；周目标（daily 型）日期必须落在绑定周内；once 型不允许打卡。
func (s *Service) Checkin(goalID int64, date string) (*model.GoalCheckin, error) {
	g, err := s.GetGoal(goalID)
	if err != nil {
		return nil, err
	}
	if g.Status == model.GoalStatusArchived {
		return nil, fmt.Errorf("%w: 已归档目标不能打卡", ErrInvalid)
	}
	var effective string
	switch g.GoalType {
	case model.GoalTypeDay:
		effective = *g.Date
		if date != "" && date != effective {
			return nil, fmt.Errorf("%w: 打卡日期必须等于目标日期 %s", ErrInvalid, effective)
		}
	case model.GoalTypeWeek:
		if g.GoalMode == model.GoalModeOnce {
			return nil, fmt.Errorf("%w: 一次性周目标不支持打卡，请通过状态完成", ErrInvalid)
		}
		if date == "" || !validDate(date) {
			return nil, fmt.Errorf("%w: 周目标打卡需要 checkin_date", ErrInvalid)
		}
		we, _ := weekEndDate(*g.WeekStart)
		if date < *g.WeekStart || date > we {
			return nil, fmt.Errorf("%w: 打卡日期必须落在绑定周内（%s ~ %s）", ErrInvalid, *g.WeekStart, we)
		}
		effective = date
	}
	var exist int64
	if err := s.db.Model(&model.GoalCheckin{}).Where("goal_id = ? AND checkin_date = ?", goalID, effective).Count(&exist).Error; err != nil {
		return nil, err
	}
	if exist > 0 {
		return nil, fmt.Errorf("%w: 该日期已打卡", ErrConflict)
	}
	ck := model.GoalCheckin{GoalID: goalID, CheckinDate: effective}
	if err := s.db.Create(&ck).Error; err != nil {
		return nil, err
	}
	return &ck, nil
}

// ListCheckins 查询区间内的打卡记录（按日期升序）。
func (s *Service) ListCheckins(from, to string) ([]model.GoalCheckin, error) {
	if !validDate(from) || !validDate(to) {
		return nil, fmt.Errorf("%w: from/to 格式非法", ErrInvalid)
	}
	var list []model.GoalCheckin
	if err := s.db.Where("checkin_date BETWEEN ? AND ?", from, to).Order("checkin_date ASC, id ASC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// Uncheckin 撤销打卡；不存在则幂等成功。
func (s *Service) Uncheckin(goalID int64, date string) error {
	g, err := s.GetGoal(goalID)
	if err != nil {
		return err
	}
	effective := date
	if g.GoalType == model.GoalTypeDay {
		effective = *g.Date
	} else if date == "" || !validDate(date) {
		return fmt.Errorf("%w: 周目标撤销打卡需要 checkin_date", ErrInvalid)
	}
	res := s.db.Where("goal_id = ? AND checkin_date = ?", goalID, effective).Delete(&model.GoalCheckin{})
	if res.Error != nil {
		return res.Error
	}
	return nil
}
