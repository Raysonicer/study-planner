package service

import (
	"testing"
)

func TestReviewUpsertAndSnapshotFreeze(t *testing.T) {
	s := newTestService(t)
	ws := currentMonday()
	due := ws
	task := mkTask(t, s, "复盘前任务", &due)

	snap, err := s.ComputeSnapshot(ws)
	if err != nil {
		t.Fatalf("snapshot: %v", err)
	}
	title := "第 N 周复盘"
	content := "KPT 内容"
	r, err := s.UpsertReview(ws, ReviewParams{Title: &title, ContentMD: &content, StatsSnapshot: ptrJSON(snap)})
	if err != nil {
		t.Fatalf("upsert: %v", err)
	}
	frozenDone := r.StatsSnapshot
	if len(frozenDone) == 0 {
		t.Fatalf("snapshot not stored")
	}

	// 数据变化后再取复盘 → 快照不变
	if _, err := s.CompleteTask(task); err != nil {
		t.Fatalf("complete: %v", err)
	}
	got, err := s.GetReview(ws)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if string(got.StatsSnapshot) != string(frozenDone) {
		t.Fatalf("snapshot changed after data change（冻结失败）")
	}

	// 非周一 → 拒绝
	if _, err := s.UpsertReview(addDays(ws, 1), ReviewParams{}); err == nil {
		t.Fatalf("non-monday week_start should be rejected")
	}
}
