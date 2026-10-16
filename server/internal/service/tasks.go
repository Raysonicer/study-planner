package service

import (
	"fmt"
	"strings"

	"studypilot/server/internal/model"
)

type TaskFilter struct {
	Status   string
	GoalID   *int64
	ParentID *int64
	Tag      string
	DueFrom  string
	DueTo    string
	Q        string
}

type TaskParams struct {
	Title      string
	Note       *string
	ParentID   *int64
	GoalID     *int64
	Status     *string
	Priority   *int
	DueDate    *string
	EstMinutes *int
	Tags       []string
	Sort       *int
}

func validateTaskParams(p *TaskParams) error {
	if p.Title == "" {
		return fmt.Errorf("%w: title 不能为空", ErrInvalid)
	}
	if p.Status != nil {
		switch *p.Status {
		case model.TaskStatusTodo, model.TaskStatusDoing, model.TaskStatusDone:
		default:
			return fmt.Errorf("%w: 非法 status", ErrInvalid)
		}
	}
	if p.Priority != nil && (*p.Priority < 0 || *p.Priority > 3) {
		return fmt.Errorf("%w: priority 必须在 0-3", ErrInvalid)
	}
	if p.DueDate != nil && *p.DueDate != "" && !validDate(*p.DueDate) {
		return fmt.Errorf("%w: due_date 格式非法", ErrInvalid)
	}
	if p.EstMinutes != nil && *p.EstMinutes <= 0 {
		return fmt.Errorf("%w: est_minutes 必须为正数", ErrInvalid)
	}
	return nil
}

// wouldCreateCycle 判断把 taskID 挂到 parentID 下是否会形成环。
func (s *Service) wouldCreateCycle(taskID, parentID int64) (bool, error) {
	cur := parentID
	for depth := 0; depth < 10000; depth++ {
		if cur == taskID {
			return true, nil
		}
		var t model.Task
		if err := s.db.Select("parent_id").First(&t, cur).Error; err != nil {
			return false, nil // 父链到头或不存在
		}
		if t.ParentID == nil {
			return false, nil
		}
		cur = *t.ParentID
	}
	return false, fmt.Errorf("%w: 任务树深度超限", ErrInvalid)
}

func (s *Service) applyTask(p *TaskParams, t *model.Task) {
	t.Title = p.Title
	if p.Note != nil {
		t.Note = *p.Note
	}
	t.ParentID = p.ParentID
	t.GoalID = p.GoalID
	if p.Status != nil {
		t.Status = *p.Status
	}
	if p.Priority != nil {
		t.Priority = *p.Priority
	}
	t.DueDate = p.DueDate
	t.EstMinutes = p.EstMinutes
	if p.Tags != nil {
		t.Tags = p.Tags
	}
	if p.Sort != nil {
		t.Sort = *p.Sort
	}
}

func (s *Service) ListTasks(f TaskFilter) ([]model.Task, error) {
	q := s.db.Order("sort ASC, id ASC")
	if f.Status != "" {
		q = q.Where("status = ?", f.Status)
	}
	if f.GoalID != nil {
		q = q.Where("goal_id = ?", *f.GoalID)
	}
	if f.ParentID != nil {
		q = q.Where("parent_id = ?", *f.ParentID)
	}
	if f.Tag != "" {
		q = q.Where("tags LIKE ?", "%\""+f.Tag+"\"%")
	}
	if f.DueFrom != "" {
		q = q.Where("due_date >= ?", f.DueFrom)
	}
	if f.DueTo != "" {
		q = q.Where("due_date <= ?", f.DueTo)
	}
	if f.Q != "" {
		q = q.Where("title LIKE ?", "%"+f.Q+"%")
	}
	var tasks []model.Task
	if err := q.Find(&tasks).Error; err != nil {
		return nil, err
	}
	return tasks, nil
}

func (s *Service) GetTask(id int64) (*model.Task, error) {
	var t model.Task
	if err := s.db.First(&t, id).Error; err != nil {
		return nil, ErrNotFound
	}
	return &t, nil
}

func (s *Service) CreateTask(p TaskParams) (*model.Task, error) {
	if p.Status == nil {
		p.Status = ptr(model.TaskStatusTodo)
	}
	if p.Priority == nil {
		p.Priority = ptr(1)
	}
	if err := validateTaskParams(&p); err != nil {
		return nil, err
	}
	if p.ParentID != nil {
		if _, err := s.GetTask(*p.ParentID); err != nil {
			return nil, fmt.Errorf("%w: 父任务不存在", ErrInvalid)
		}
	}
	if p.GoalID != nil {
		if _, err := s.GetGoal(*p.GoalID); err != nil {
			return nil, fmt.Errorf("%w: 目标不存在", ErrInvalid)
		}
	}
	t := model.Task{}
	s.applyTask(&p, &t)
	if err := s.db.Create(&t).Error; err != nil {
		return nil, err
	}
	_ = s.writeEvent(t.ID, model.EventCreated)
	return &t, nil
}

func (s *Service) UpdateTask(id int64, p TaskParams) (*model.Task, error) {
	t, err := s.GetTask(id)
	if err != nil {
		return nil, err
	}
	if p.Status == nil {
		p.Status = ptr(t.Status)
	}
	if p.Priority == nil {
		p.Priority = ptr(t.Priority)
	}
	if err := validateTaskParams(&p); err != nil {
		return nil, err
	}
	if p.ParentID != nil {
		if *p.ParentID == id {
			return nil, fmt.Errorf("%w: 任务不能作为自己的子任务", ErrInvalid)
		}
		if _, err := s.GetTask(*p.ParentID); err != nil {
			return nil, fmt.Errorf("%w: 父任务不存在", ErrInvalid)
		}
		if cycle, err := s.wouldCreateCycle(id, *p.ParentID); err != nil || cycle {
			return nil, fmt.Errorf("%w: 父子关系会形成环", ErrInvalid)
		}
	} else {
		p.ParentID = nil
	}
	if p.GoalID != nil {
		if _, err := s.GetGoal(*p.GoalID); err != nil {
			return nil, fmt.Errorf("%w: 目标不存在", ErrInvalid)
		}
	}
	s.applyTask(&p, t)
	if err := s.db.Save(t).Error; err != nil {
		return nil, err
	}
	return t, nil
}

func (s *Service) DeleteTask(id int64) error {
	if _, err := s.GetTask(id); err != nil {
		return err
	}
	return s.db.Delete(&model.Task{}, id).Error
}

func (s *Service) writeEvent(taskID int64, typ string) error {
	ev := model.TaskEvent{
		TaskID:       taskID,
		EventType:    typ,
		OccurredDate: today(),
	}
	return s.db.Create(&ev).Error
}

// CompleteTask 幂等：已完成则直接返回，不重复写事件。
func (s *Service) CompleteTask(id int64) (*model.Task, error) {
	t, err := s.GetTask(id)
	if err != nil {
		return nil, err
	}
	if t.Status == model.TaskStatusDone {
		return t, nil
	}
	t.Status = model.TaskStatusDone
	t.CompletedAt = ptr(today())
	if err := s.db.Save(t).Error; err != nil {
		return nil, err
	}
	if err := s.writeEvent(id, model.EventDone); err != nil {
		return nil, err
	}
	return t, nil
}

// ReopenTask 幂等：未完成则直接返回。
func (s *Service) ReopenTask(id int64) (*model.Task, error) {
	t, err := s.GetTask(id)
	if err != nil {
		return nil, err
	}
	if t.Status != model.TaskStatusDone {
		return t, nil
	}
	t.Status = model.TaskStatusTodo
	t.CompletedAt = nil
	if err := s.db.Save(t).Error; err != nil {
		return nil, err
	}
	if err := s.writeEvent(id, model.EventReopen); err != nil {
		return nil, err
	}
	return t, nil
}

// BatchTasks 批量创建子任务：每行一个，自动去空行（T-05）。
func (s *Service) BatchTasks(parentID, goalID *int64, lines []string) ([]model.Task, error) {
	var titles []string
	for _, ln := range lines {
		if t := strings.TrimSpace(ln); t != "" {
			titles = append(titles, t)
		}
	}
	if len(titles) == 0 {
		return nil, fmt.Errorf("%w: lines 至少包含一行非空文本", ErrInvalid)
	}
	if parentID != nil {
		if _, err := s.GetTask(*parentID); err != nil {
			return nil, fmt.Errorf("%w: 父任务不存在", ErrInvalid)
		}
	}
	if goalID != nil {
		if _, err := s.GetGoal(*goalID); err != nil {
			return nil, fmt.Errorf("%w: 目标不存在", ErrInvalid)
		}
	}
	var maxSort int
	sc := s.db.Model(&model.Task{})
	if parentID != nil {
		sc = sc.Where("parent_id = ?", *parentID)
	} else {
		sc = sc.Where("parent_id IS NULL")
	}
	_ = sc.Select("COALESCE(MAX(sort), -1)").Scan(&maxSort).Error
	created := make([]model.Task, 0, len(titles))
	for i, title := range titles {
		t := model.Task{
			Title:    title,
			ParentID: parentID,
			GoalID:   goalID,
			Status:   model.TaskStatusTodo,
			Priority: 1,
			Sort:     maxSort + 1 + i,
			Tags:     []string{},
		}
		if err := s.db.Create(&t).Error; err != nil {
			return nil, err
		}
		_ = s.writeEvent(t.ID, model.EventCreated)
		created = append(created, t)
	}
	return created, nil
}

// ReorderTasks 同层级排序：按传入顺序写 sort（T-10）。
func (s *Service) ReorderTasks(parentID *int64, orderedIDs []int64) error {
	for i, id := range orderedIDs {
		if err := s.db.Model(&model.Task{}).Where("id = ?", id).Update("sort", i).Error; err != nil {
			return err
		}
	}
	return nil
}
