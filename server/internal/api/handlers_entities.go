package api

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	"studypilot/server/internal/service"
)

type handlers struct {
	svc *service.Service
}

func pathID(c echo.Context) (int64, error) {
	return strconv.ParseInt(c.Param("id"), 10, 64)
}

func queryInt64(c echo.Context, key string) (*int64, error) {
	raw := c.QueryParam(key)
	if raw == "" {
		return nil, nil
	}
	v, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return nil, err
	}
	return &v, nil
}

// ---- goals ----

func (h *handlers) listGoals(c echo.Context) error {
	goals, err := h.svc.ListGoals(service.GoalFilter{
		GoalType:  c.QueryParam("goal_type"),
		Date:      c.QueryParam("date"),
		WeekStart: c.QueryParam("week_start"),
		Status:    c.QueryParam("status"),
	})
	if err != nil {
		return handleErr(c, err)
	}
	return c.JSON(http.StatusOK, goals)
}

func (h *handlers) createGoal(c echo.Context) error {
	var in goalInput
	if err := c.Bind(&in); err != nil {
		return fail(c, http.StatusBadRequest, "invalid_input", "请求体格式错误")
	}
	g, err := h.svc.CreateGoal(in.params())
	if err != nil {
		return handleErr(c, err)
	}
	return c.JSON(http.StatusCreated, g)
}

func (h *handlers) getGoal(c echo.Context) error {
	id, err := pathID(c)
	if err != nil {
		return fail(c, http.StatusBadRequest, "invalid_input", "id 非法")
	}
	g, err := h.svc.GetGoal(id)
	if err != nil {
		return handleErr(c, err)
	}
	return c.JSON(http.StatusOK, g)
}

func (h *handlers) updateGoal(c echo.Context) error {
	id, err := pathID(c)
	if err != nil {
		return fail(c, http.StatusBadRequest, "invalid_input", "id 非法")
	}
	var in goalInput
	if err := c.Bind(&in); err != nil {
		return fail(c, http.StatusBadRequest, "invalid_input", "请求体格式错误")
	}
	g, err := h.svc.UpdateGoal(id, in.params())
	if err != nil {
		return handleErr(c, err)
	}
	return c.JSON(http.StatusOK, g)
}

func (h *handlers) deleteGoal(c echo.Context) error {
	id, err := pathID(c)
	if err != nil {
		return fail(c, http.StatusBadRequest, "invalid_input", "id 非法")
	}
	if err := h.svc.DeleteGoal(id); err != nil {
		return handleErr(c, err)
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *handlers) checkin(c echo.Context) error {
	id, err := pathID(c)
	if err != nil {
		return fail(c, http.StatusBadRequest, "invalid_input", "id 非法")
	}
	var in checkinInput
	_ = c.Bind(&in)
	date := ""
	if in.CheckinDate != nil {
		date = *in.CheckinDate
	}
	ck, err := h.svc.Checkin(id, date)
	if err != nil {
		return handleErr(c, err)
	}
	return c.JSON(http.StatusCreated, ck)
}

func (h *handlers) uncheckin(c echo.Context) error {
	id, err := pathID(c)
	if err != nil {
		return fail(c, http.StatusBadRequest, "invalid_input", "id 非法")
	}
	if err := h.svc.Uncheckin(id, c.QueryParam("checkin_date")); err != nil {
		return handleErr(c, err)
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *handlers) listCheckins(c echo.Context) error {
	from := c.QueryParam("from")
	to := c.QueryParam("to")
	if from == "" || to == "" {
		return fail(c, http.StatusBadRequest, "invalid_input", "from/to 必填")
	}
	list, err := h.svc.ListCheckins(from, to)
	if err != nil {
		return handleErr(c, err)
	}
	return c.JSON(http.StatusOK, list)
}

// ---- tasks ----

func (h *handlers) listTasks(c echo.Context) error {
	goalID, err := queryInt64(c, "goal_id")
	if err != nil {
		return fail(c, http.StatusBadRequest, "invalid_input", "goal_id 非法")
	}
	parentID, err := queryInt64(c, "parent_id")
	if err != nil {
		return fail(c, http.StatusBadRequest, "invalid_input", "parent_id 非法")
	}
	tasks, err := h.svc.ListTasks(service.TaskFilter{
		Status:   c.QueryParam("status"),
		GoalID:   goalID,
		ParentID: parentID,
		Tag:      c.QueryParam("tag"),
		DueFrom:  c.QueryParam("due_from"),
		DueTo:    c.QueryParam("due_to"),
		Q:        c.QueryParam("q"),
	})
	if err != nil {
		return handleErr(c, err)
	}
	return c.JSON(http.StatusOK, tasks)
}

func (h *handlers) createTask(c echo.Context) error {
	var in taskInput
	if err := c.Bind(&in); err != nil {
		return fail(c, http.StatusBadRequest, "invalid_input", "请求体格式错误")
	}
	t, err := h.svc.CreateTask(in.params())
	if err != nil {
		return handleErr(c, err)
	}
	return c.JSON(http.StatusCreated, t)
}

func (h *handlers) getTask(c echo.Context) error {
	id, err := pathID(c)
	if err != nil {
		return fail(c, http.StatusBadRequest, "invalid_input", "id 非法")
	}
	t, err := h.svc.GetTask(id)
	if err != nil {
		return handleErr(c, err)
	}
	return c.JSON(http.StatusOK, t)
}

func (h *handlers) updateTask(c echo.Context) error {
	id, err := pathID(c)
	if err != nil {
		return fail(c, http.StatusBadRequest, "invalid_input", "id 非法")
	}
	var in taskInput
	if err := c.Bind(&in); err != nil {
		return fail(c, http.StatusBadRequest, "invalid_input", "请求体格式错误")
	}
	t, err := h.svc.UpdateTask(id, in.params())
	if err != nil {
		return handleErr(c, err)
	}
	return c.JSON(http.StatusOK, t)
}

func (h *handlers) deleteTask(c echo.Context) error {
	id, err := pathID(c)
	if err != nil {
		return fail(c, http.StatusBadRequest, "invalid_input", "id 非法")
	}
	if err := h.svc.DeleteTask(id); err != nil {
		return handleErr(c, err)
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *handlers) completeTask(c echo.Context) error {
	id, err := pathID(c)
	if err != nil {
		return fail(c, http.StatusBadRequest, "invalid_input", "id 非法")
	}
	t, err := h.svc.CompleteTask(id)
	if err != nil {
		return handleErr(c, err)
	}
	return c.JSON(http.StatusOK, t)
}

func (h *handlers) reopenTask(c echo.Context) error {
	id, err := pathID(c)
	if err != nil {
		return fail(c, http.StatusBadRequest, "invalid_input", "id 非法")
	}
	t, err := h.svc.ReopenTask(id)
	if err != nil {
		return handleErr(c, err)
	}
	return c.JSON(http.StatusOK, t)
}

func (h *handlers) batchTasks(c echo.Context) error {
	var in batchTasksInput
	if err := c.Bind(&in); err != nil {
		return fail(c, http.StatusBadRequest, "invalid_input", "请求体格式错误")
	}
	tasks, err := h.svc.BatchTasks(in.ParentID, in.GoalID, in.Lines)
	if err != nil {
		return handleErr(c, err)
	}
	return c.JSON(http.StatusCreated, tasks)
}

func (h *handlers) reorderTasks(c echo.Context) error {
	var in reorderInput
	if err := c.Bind(&in); err != nil {
		return fail(c, http.StatusBadRequest, "invalid_input", "请求体格式错误")
	}
	if err := h.svc.ReorderTasks(in.ParentID, in.OrderedIDs); err != nil {
		return handleErr(c, err)
	}
	return c.NoContent(http.StatusNoContent)
}

// ---- blocks ----

func (h *handlers) listBlocks(c echo.Context) error {
	blocks, err := h.svc.ListBlocks(c.QueryParam("from"), c.QueryParam("to"))
	if err != nil {
		return handleErr(c, err)
	}
	return c.JSON(http.StatusOK, blocks)
}

func (h *handlers) createBlock(c echo.Context) error {
	var in blockInput
	if err := c.Bind(&in); err != nil {
		return fail(c, http.StatusBadRequest, "invalid_input", "请求体格式错误")
	}
	b, err := h.svc.CreateBlock(in.params())
	if err != nil {
		return handleErr(c, err)
	}
	return c.JSON(http.StatusCreated, b)
}

func (h *handlers) updateBlock(c echo.Context) error {
	id, err := pathID(c)
	if err != nil {
		return fail(c, http.StatusBadRequest, "invalid_input", "id 非法")
	}
	var in blockInput
	if err := c.Bind(&in); err != nil {
		return fail(c, http.StatusBadRequest, "invalid_input", "请求体格式错误")
	}
	b, err := h.svc.UpdateBlock(id, in.params())
	if err != nil {
		return handleErr(c, err)
	}
	return c.JSON(http.StatusOK, b)
}

func (h *handlers) deleteBlock(c echo.Context) error {
	id, err := pathID(c)
	if err != nil {
		return fail(c, http.StatusBadRequest, "invalid_input", "id 非法")
	}
	if err := h.svc.DeleteBlock(id); err != nil {
		return handleErr(c, err)
	}
	return c.NoContent(http.StatusNoContent)
}
