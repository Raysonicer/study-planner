package api

import (
	"encoding/json"

	"studypilot/server/internal/service"
)

// ---- 输入 DTO（与 docs/openapi.yaml 对齐） ----

type goalInput struct {
	Title       string  `json:"title"`
	GoalType    string  `json:"goal_type"`
	GoalMode    string  `json:"goal_mode"`
	Date        *string `json:"date"`
	WeekStart   *string `json:"week_start"`
	CheckinDays *int    `json:"checkin_days"`
	Status      *string `json:"status"`
	Note        *string `json:"note"`
	Sort        *int    `json:"sort"`
}

func (in goalInput) params() service.GoalParams {
	return service.GoalParams{
		Title:       in.Title,
		GoalType:    in.GoalType,
		GoalMode:    in.GoalMode,
		Date:        in.Date,
		WeekStart:   in.WeekStart,
		CheckinDays: in.CheckinDays,
		Status:      in.Status,
		Note:        in.Note,
		Sort:        in.Sort,
	}
}

type checkinInput struct {
	CheckinDate *string `json:"checkin_date"`
}

type taskInput struct {
	Title      string   `json:"title"`
	Note       *string  `json:"note"`
	ParentID   *int64   `json:"parent_id"`
	GoalID     *int64   `json:"goal_id"`
	Status     *string  `json:"status"`
	Priority   *int     `json:"priority"`
	DueDate    *string  `json:"due_date"`
	EstMinutes *int     `json:"est_minutes"`
	Tags       []string `json:"tags"`
	Sort       *int     `json:"sort"`
}

func (in taskInput) params() service.TaskParams {
	return service.TaskParams{
		Title:      in.Title,
		Note:       in.Note,
		ParentID:   in.ParentID,
		GoalID:     in.GoalID,
		Status:     in.Status,
		Priority:   in.Priority,
		DueDate:    in.DueDate,
		EstMinutes: in.EstMinutes,
		Tags:       in.Tags,
		Sort:       in.Sort,
	}
}

type batchTasksInput struct {
	ParentID *int64   `json:"parent_id"`
	GoalID   *int64   `json:"goal_id"`
	Lines    []string `json:"lines"`
}

type reorderInput struct {
	ParentID   *int64  `json:"parent_id"`
	OrderedIDs []int64 `json:"ordered_ids"`
}

type blockInput struct {
	BlockDate string `json:"block_date"`
	StartMin  int    `json:"start_min"`
	EndMin    int    `json:"end_min"`
	Title     string `json:"title"`
	TaskID    *int64 `json:"task_id"`
	Color     string `json:"color"`
	Note      string `json:"note"`
}

func (in blockInput) params() service.BlockParams {
	return service.BlockParams{
		BlockDate: in.BlockDate,
		StartMin:  in.StartMin,
		EndMin:    in.EndMin,
		Title:     in.Title,
		TaskID:    in.TaskID,
		Color:     in.Color,
		Note:      in.Note,
	}
}

type reviewInput struct {
	Title         *string          `json:"title"`
	ContentMD     *string          `json:"content_md"`
	StatsSnapshot *json.RawMessage `json:"stats_snapshot"`
}
