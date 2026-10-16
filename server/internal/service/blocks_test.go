package service

import (
	"errors"
	"testing"
)

func TestBlockValidationAndCRUD(t *testing.T) {
	s := newTestService(t)
	d := today()
	if _, err := s.CreateBlock(BlockParams{BlockDate: d, StartMin: 600, EndMin: 600, Title: "x"}); !errors.Is(err, ErrInvalid) {
		t.Fatalf("end<=start err = %v, want invalid", err)
	}
	if _, err := s.CreateBlock(BlockParams{BlockDate: "2026/10/16", StartMin: 600, EndMin: 660, Title: "x"}); !errors.Is(err, ErrInvalid) {
		t.Fatalf("bad date err = %v, want invalid", err)
	}
	b, err := s.CreateBlock(BlockParams{BlockDate: d, StartMin: 600, EndMin: 660, Title: "刷题"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	// 关联不存在的任务 → 拒绝
	ghost := int64(99999)
	if _, err := s.UpdateBlock(b.ID, BlockParams{BlockDate: d, StartMin: 600, EndMin: 660, Title: "刷题", TaskID: &ghost}); !errors.Is(err, ErrInvalid) {
		t.Fatalf("ghost task err = %v, want invalid", err)
	}
	// 平移
	b2, err := s.UpdateBlock(b.ID, BlockParams{BlockDate: d, StartMin: 900, EndMin: 960, Title: "刷题"})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if b2.StartMin != 900 || b2.EndMin != 960 {
		t.Fatalf("moved block = %d-%d", b2.StartMin, b2.EndMin)
	}
	if err := s.DeleteBlock(b.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := s.GetBlock(b.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("block should be gone")
	}
}
