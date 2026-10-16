package service

import (
	"fmt"

	"studypilot/server/internal/model"
)

type BlockParams struct {
	BlockDate string
	StartMin  int
	EndMin    int
	Title     string
	TaskID    *int64
	Color     string
	Note      string
}

func validateBlockParams(p *BlockParams) error {
	if p.Title == "" {
		return fmt.Errorf("%w: title 不能为空", ErrInvalid)
	}
	if !validDate(p.BlockDate) {
		return fmt.Errorf("%w: block_date 格式非法", ErrInvalid)
	}
	if p.StartMin < 0 || p.StartMin > 1439 {
		return fmt.Errorf("%w: start_min 必须在 0-1439", ErrInvalid)
	}
	if p.EndMin < 1 || p.EndMin > 1440 {
		return fmt.Errorf("%w: end_min 必须在 1-1440", ErrInvalid)
	}
	if p.EndMin <= p.StartMin {
		return fmt.Errorf("%w: end_min 必须大于 start_min", ErrInvalid)
	}
	return nil
}

func (s *Service) ListBlocks(from, to string) ([]model.Block, error) {
	if !validDate(from) || !validDate(to) {
		return nil, fmt.Errorf("%w: from/to 格式非法", ErrInvalid)
	}
	var blocks []model.Block
	if err := s.db.Where("block_date BETWEEN ? AND ?", from, to).Order("block_date ASC, start_min ASC, id ASC").Find(&blocks).Error; err != nil {
		return nil, err
	}
	return blocks, nil
}

func (s *Service) GetBlock(id int64) (*model.Block, error) {
	var b model.Block
	if err := s.db.First(&b, id).Error; err != nil {
		return nil, ErrNotFound
	}
	return &b, nil
}

func (s *Service) CreateBlock(p BlockParams) (*model.Block, error) {
	if err := validateBlockParams(&p); err != nil {
		return nil, err
	}
	if p.TaskID != nil {
		if _, err := s.GetTask(*p.TaskID); err != nil {
			return nil, fmt.Errorf("%w: 关联任务不存在", ErrInvalid)
		}
	}
	b := model.Block{
		BlockDate: p.BlockDate,
		StartMin:  p.StartMin,
		EndMin:    p.EndMin,
		Title:     p.Title,
		TaskID:    p.TaskID,
		Color:     p.Color,
		Note:      p.Note,
	}
	if err := s.db.Create(&b).Error; err != nil {
		return nil, err
	}
	return &b, nil
}

func (s *Service) UpdateBlock(id int64, p BlockParams) (*model.Block, error) {
	b, err := s.GetBlock(id)
	if err != nil {
		return nil, err
	}
	if err := validateBlockParams(&p); err != nil {
		return nil, err
	}
	if p.TaskID != nil {
		if _, err := s.GetTask(*p.TaskID); err != nil {
			return nil, fmt.Errorf("%w: 关联任务不存在", ErrInvalid)
		}
	}
	b.BlockDate = p.BlockDate
	b.StartMin = p.StartMin
	b.EndMin = p.EndMin
	b.Title = p.Title
	b.TaskID = p.TaskID
	b.Color = p.Color
	b.Note = p.Note
	if err := s.db.Save(b).Error; err != nil {
		return nil, err
	}
	return b, nil
}

func (s *Service) DeleteBlock(id int64) error {
	if _, err := s.GetBlock(id); err != nil {
		return err
	}
	return s.db.Delete(&model.Block{}, id).Error
}
