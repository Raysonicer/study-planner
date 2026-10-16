package model

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

const (
	GoalTypeDay  = "day"
	GoalTypeWeek = "week"

	GoalModeDaily = "daily" // 每日打卡型
	GoalModeOnce  = "once"  // 一次性（仅周目标）

	GoalStatusActive   = "active"
	GoalStatusDone     = "done"
	GoalStatusArchived = "archived"

	TaskStatusTodo  = "todo"
	TaskStatusDoing = "doing"
	TaskStatusDone  = "done"

	EventCreated = "created"
	EventDone    = "done"
	EventReopen  = "reopen"
)

// 时间统一存本地字符串（驱动对 TEXT->time.Time 扫描不友好），由钩子维护。
const TimeLayout = "2006-01-02 15:04:05"

func nowString() string { return time.Now().Format(TimeLayout) }

type Setting struct {
	Key   string `gorm:"primaryKey" json:"key"`
	Value string `gorm:"not null" json:"value"`
}

func (Setting) TableName() string { return "settings" }

type Goal struct {
	ID          int64   `gorm:"primaryKey" json:"id"`
	Title       string  `gorm:"not null" json:"title"`
	GoalType    string  `gorm:"not null" json:"goal_type"`
	GoalMode    string  `gorm:"not null;default:daily" json:"goal_mode"`
	Date        *string `json:"date"`
	WeekStart   *string `json:"week_start"`
	CheckinDays int     `gorm:"not null" json:"checkin_days"`
	Status      string  `gorm:"not null;default:active" json:"status"`
	Note        string  `gorm:"not null;default:''" json:"note"`
	Sort        int     `gorm:"not null;default:0" json:"sort"`
	CompletedAt *string `json:"completed_at"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

func (Goal) TableName() string { return "goals" }
func (g *Goal) BeforeCreate(*gorm.DB) error {
	g.CreatedAt, g.UpdatedAt = nowString(), nowString()
	return nil
}
func (g *Goal) BeforeUpdate(*gorm.DB) error { g.UpdatedAt = nowString(); return nil }

type GoalCheckin struct {
	ID          int64  `gorm:"primaryKey" json:"id"`
	GoalID      int64  `gorm:"not null;index" json:"goal_id"`
	CheckinDate string `gorm:"not null;index" json:"checkin_date"`
	CreatedAt   string `json:"created_at"`
}

func (GoalCheckin) TableName() string { return "goal_checkins" }
func (c *GoalCheckin) BeforeCreate(*gorm.DB) error {
	c.CreatedAt = nowString()
	return nil
}

type Task struct {
	ID          int64    `gorm:"primaryKey" json:"id"`
	Title       string   `gorm:"not null" json:"title"`
	Note        string   `gorm:"not null;default:''" json:"note"`
	ParentID    *int64   `gorm:"index" json:"parent_id"`
	GoalID      *int64   `gorm:"index" json:"goal_id"`
	Status      string   `gorm:"not null;default:todo" json:"status"`
	Priority    int      `gorm:"not null" json:"priority"`
	DueDate     *string  `json:"due_date"`
	EstMinutes  *int     `json:"est_minutes"`
	Tags        []string `gorm:"serializer:json;default:'[]'" json:"tags"`
	Sort        int      `gorm:"not null;default:0" json:"sort"`
	CompletedAt *string  `json:"completed_at"`
	CreatedAt   string   `json:"created_at"`
	UpdatedAt   string   `json:"updated_at"`
}

func (Task) TableName() string { return "tasks" }
func (t *Task) BeforeCreate(*gorm.DB) error {
	if t.Tags == nil {
		t.Tags = []string{}
	}
	t.CreatedAt, t.UpdatedAt = nowString(), nowString()
	return nil
}
func (t *Task) BeforeUpdate(*gorm.DB) error { t.UpdatedAt = nowString(); return nil }

type TaskEvent struct {
	ID           int64  `gorm:"primaryKey" json:"id"`
	TaskID       int64  `gorm:"not null;index" json:"task_id"`
	EventType    string `gorm:"not null" json:"event_type"`
	OccurredDate string `gorm:"not null;index" json:"occurred_date"`
	OccurredAt   string `json:"occurred_at"`
}

func (TaskEvent) TableName() string { return "task_events" }
func (e *TaskEvent) BeforeCreate(*gorm.DB) error {
	e.OccurredAt = nowString()
	return nil
}

type Block struct {
	ID        int64  `gorm:"primaryKey" json:"id"`
	BlockDate string `gorm:"not null;index" json:"block_date"`
	StartMin  int    `gorm:"not null" json:"start_min"`
	EndMin    int    `gorm:"not null" json:"end_min"`
	Title     string `gorm:"not null" json:"title"`
	TaskID    *int64 `gorm:"index" json:"task_id"`
	Color     string `gorm:"not null;default:''" json:"color"`
	Note      string `gorm:"not null;default:''" json:"note"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func (Block) TableName() string { return "blocks" }
func (b *Block) BeforeCreate(*gorm.DB) error {
	b.CreatedAt, b.UpdatedAt = nowString(), nowString()
	return nil
}
func (b *Block) BeforeUpdate(*gorm.DB) error { b.UpdatedAt = nowString(); return nil }

type Review struct {
	ID            int64           `gorm:"primaryKey" json:"id"`
	WeekStart     string          `gorm:"not null;uniqueIndex" json:"week_start"`
	Title         string          `gorm:"not null;default:''" json:"title"`
	ContentMD     string          `gorm:"not null;default:''" json:"content_md"`
	StatsSnapshot json.RawMessage `gorm:"type:text;not null;default:'{}'" json:"stats_snapshot"`
	CreatedAt     string          `json:"created_at"`
	UpdatedAt     string          `json:"updated_at"`
}

func (Review) TableName() string { return "reviews" }
func (r *Review) BeforeCreate(*gorm.DB) error {
	r.CreatedAt, r.UpdatedAt = nowString(), nowString()
	return nil
}
func (r *Review) BeforeUpdate(*gorm.DB) error { r.UpdatedAt = nowString(); return nil }
