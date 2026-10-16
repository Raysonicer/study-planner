package service

import (
	"errors"
	"testing"

	"studypilot/server/internal/model"
)

func countEvents(t *testing.T, s *Service, taskID int64, typ string) int64 {
	t.Helper()
	var n int64
	if err := s.db.Model(&model.TaskEvent{}).Where("task_id = ? AND event_type = ?", taskID, typ).Count(&n).Error; err != nil {
		t.Fatalf("count events: %v", err)
	}
	return n
}

func TestTaskTreeCyclePrevention(t *testing.T) {
	s := newTestService(t)
	root := mkTask(t, s, "根任务", nil)
	child := mkTask(t, s, "子任务", nil)
	grand := mkTask(t, s, "孙任务", nil)

	upd := func(id, parent int64) error {
		_, err := s.UpdateTask(id, TaskParams{Title: "t", ParentID: &parent})
		return err
	}
	if err := upd(child, root); err != nil {
		t.Fatalf("child under root: %v", err)
	}
	if err := upd(grand, child); err != nil {
		t.Fatalf("grand under child: %v", err)
	}
	// root 挂到 grand 下 → 环
	if err := upd(root, grand); !errors.Is(err, ErrInvalid) {
		t.Fatalf("cycle err = %v, want invalid", err)
	}
	// 自己挂自己 → 拒绝
	if err := upd(root, root); !errors.Is(err, ErrInvalid) {
		t.Fatalf("self-parent err = %v, want invalid", err)
	}
}

func TestBatchTasks(t *testing.T) {
	s := newTestService(t)
	root := mkTask(t, s, "学习 Go", nil)
	lines := []string{" 语法 ", "", "并发", "标准库 "}
	tasks, err := s.BatchTasks(&root, nil, lines)
	if err != nil {
		t.Fatalf("batch: %v", err)
	}
	if len(tasks) != 3 {
		t.Fatalf("created = %d, want 3", len(tasks))
	}
	if tasks[0].Title != "语法" {
		t.Fatalf("title[0] = %q, want 语法", tasks[0].Title)
	}
	if tasks[0].ParentID == nil || *tasks[0].ParentID != root {
		t.Fatalf("parent mismatch")
	}
}

func TestCompleteReopenIdempotentEvents(t *testing.T) {
	s := newTestService(t)
	id := mkTask(t, s, "任务", nil)

	if n := countEvents(t, s, id, model.EventCreated); n != 1 {
		t.Fatalf("created events = %d, want 1", n)
	}
	done, err := s.CompleteTask(id)
	if err != nil {
		t.Fatalf("complete: %v", err)
	}
	if done.Status != model.TaskStatusDone || done.CompletedAt == nil {
		t.Fatalf("completed state wrong: %+v", done)
	}
	// 幂等：再次完成不产生新事件
	if _, err := s.CompleteTask(id); err != nil {
		t.Fatalf("complete again: %v", err)
	}
	if n := countEvents(t, s, id, model.EventDone); n != 1 {
		t.Fatalf("done events = %d, want 1", n)
	}
	// 重开
	re, err := s.ReopenTask(id)
	if err != nil {
		t.Fatalf("reopen: %v", err)
	}
	if re.Status != model.TaskStatusTodo || re.CompletedAt != nil {
		t.Fatalf("reopened state wrong: %+v", re)
	}
	if n := countEvents(t, s, id, model.EventReopen); n != 1 {
		t.Fatalf("reopen events = %d, want 1", n)
	}
	// 未完成再重开 → 无新事件
	if _, err := s.ReopenTask(id); err != nil {
		t.Fatalf("reopen again: %v", err)
	}
	if n := countEvents(t, s, id, model.EventReopen); n != 1 {
		t.Fatalf("reopen events duplicated")
	}
}

func TestDeleteTaskCascades(t *testing.T) {
	s := newTestService(t)
	root := mkTask(t, s, "根", nil)
	childID := int64(0)
	{
		child, err := s.CreateTask(TaskParams{Title: "子", ParentID: &root})
		if err != nil {
			t.Fatalf("create child: %v", err)
		}
		childID = child.ID
	}
	if err := s.DeleteTask(root); err != nil {
		t.Fatalf("delete root: %v", err)
	}
	if _, err := s.GetTask(childID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("child should be cascaded: %v", err)
	}
}
