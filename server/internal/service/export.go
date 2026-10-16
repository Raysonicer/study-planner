package service

import (
	"fmt"
	"time"

	"gorm.io/gorm"

	"studypilot/server/internal/model"
)

// ExportPayload 全量导出结构（SYS-05：导出/导入往返无损）。
type ExportPayload struct {
	Version    string              `json:"version"`
	ExportedAt string              `json:"exported_at"`
	Goals      []model.Goal        `json:"goals"`
	Checkins   []model.GoalCheckin `json:"goal_checkins"`
	Tasks      []model.Task        `json:"tasks"`
	TaskEvents []model.TaskEvent   `json:"task_events"`
	Blocks     []model.Block       `json:"blocks"`
	Reviews    []model.Review      `json:"reviews"`
	Settings   []model.Setting     `json:"settings"`
}

func (s *Service) Export() (*ExportPayload, error) {
	out := &ExportPayload{Version: "1", ExportedAt: time.Now().Format(model.TimeLayout)}
	if err := s.db.Order("id ASC").Find(&out.Goals).Error; err != nil {
		return nil, err
	}
	if err := s.db.Order("id ASC").Find(&out.Checkins).Error; err != nil {
		return nil, err
	}
	if err := s.db.Order("id ASC").Find(&out.Tasks).Error; err != nil {
		return nil, err
	}
	if err := s.db.Order("id ASC").Find(&out.TaskEvents).Error; err != nil {
		return nil, err
	}
	if err := s.db.Order("id ASC").Find(&out.Blocks).Error; err != nil {
		return nil, err
	}
	if err := s.db.Order("id ASC").Find(&out.Reviews).Error; err != nil {
		return nil, err
	}
	if err := s.db.Order("key ASC").Find(&out.Settings).Error; err != nil {
		return nil, err
	}
	return out, nil
}

// ImportAll 用导出数据整体替换当前数据（事务，保留原始 ID 与外键关系）。
func (s *Service) ImportAll(p *ExportPayload) error {
	if p.Version == "" {
		return fmt.Errorf("%w: 无效的导出文件（缺少 version）", ErrInvalid)
	}
	total := len(p.Goals) + len(p.Checkins) + len(p.Tasks) + len(p.TaskEvents) + len(p.Blocks) + len(p.Reviews) + len(p.Settings)
	if total > 200000 {
		return fmt.Errorf("%w: 数据量过大", ErrInvalid)
	}
	return s.db.Transaction(func(tx *gorm.DB) error {
		for _, table := range []string{"goal_checkins", "task_events", "blocks", "reviews", "tasks", "goals", "settings"} {
			if err := tx.Exec("DELETE FROM " + table).Error; err != nil {
				return err
			}
		}
		for i := range p.Goals {
			if err := tx.Create(&p.Goals[i]).Error; err != nil {
				return fmt.Errorf("goals: %w", err)
			}
		}
		for i := range p.Checkins {
			if err := tx.Create(&p.Checkins[i]).Error; err != nil {
				return fmt.Errorf("goal_checkins: %w", err)
			}
		}
		for i := range p.Tasks {
			if err := tx.Create(&p.Tasks[i]).Error; err != nil {
				return fmt.Errorf("tasks: %w", err)
			}
		}
		for i := range p.TaskEvents {
			if err := tx.Create(&p.TaskEvents[i]).Error; err != nil {
				return fmt.Errorf("task_events: %w", err)
			}
		}
		for i := range p.Blocks {
			if err := tx.Create(&p.Blocks[i]).Error; err != nil {
				return fmt.Errorf("blocks: %w", err)
			}
		}
		for i := range p.Reviews {
			if err := tx.Create(&p.Reviews[i]).Error; err != nil {
				return fmt.Errorf("reviews: %w", err)
			}
		}
		for i := range p.Settings {
			if err := tx.Create(&p.Settings[i]).Error; err != nil {
				return fmt.Errorf("settings: %w", err)
			}
		}
		return nil
	})
}
