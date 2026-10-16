package api

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"studypilot/server/internal/service"
)

// ---- dashboard & stats ----

func (h *handlers) dashboard(c echo.Context) error {
	date := c.QueryParam("date")
	if date == "" {
		return fail(c, http.StatusBadRequest, "invalid_input", "date 必填")
	}
	d, err := h.svc.Dashboard(date)
	if err != nil {
		return handleErr(c, err)
	}
	return c.JSON(http.StatusOK, d)
}

func (h *handlers) statsDaily(c echo.Context) error {
	date := c.QueryParam("date")
	if date == "" {
		return fail(c, http.StatusBadRequest, "invalid_input", "date 必填")
	}
	d, err := h.svc.DailyStats(date)
	if err != nil {
		return handleErr(c, err)
	}
	return c.JSON(http.StatusOK, d)
}

func (h *handlers) statsWeekly(c echo.Context) error {
	ws := c.QueryParam("week_start")
	if ws == "" {
		return fail(c, http.StatusBadRequest, "invalid_input", "week_start 必填")
	}
	w, err := h.svc.WeeklyStats(ws)
	if err != nil {
		return handleErr(c, err)
	}
	return c.JSON(http.StatusOK, w)
}

func (h *handlers) statsHeatmap(c echo.Context) error {
	start := c.QueryParam("start")
	end := c.QueryParam("end")
	if start == "" || end == "" {
		return fail(c, http.StatusBadRequest, "invalid_input", "start/end 必填")
	}
	hm, err := h.svc.Heatmap(start, end)
	if err != nil {
		return handleErr(c, err)
	}
	return c.JSON(http.StatusOK, hm)
}

// ---- reviews ----

func (h *handlers) listReviews(c echo.Context) error {
	rs, err := h.svc.ListReviews()
	if err != nil {
		return handleErr(c, err)
	}
	return c.JSON(http.StatusOK, rs)
}

func (h *handlers) getReview(c echo.Context) error {
	r, err := h.svc.GetReview(c.Param("week_start"))
	if err != nil {
		return handleErr(c, err)
	}
	return c.JSON(http.StatusOK, r)
}

func (h *handlers) putReview(c echo.Context) error {
	var in reviewInput
	if err := c.Bind(&in); err != nil {
		return fail(c, http.StatusBadRequest, "invalid_input", "请求体格式错误")
	}
	r, err := h.svc.UpsertReview(c.Param("week_start"), service.ReviewParams{
		Title:         in.Title,
		ContentMD:     in.ContentMD,
		StatsSnapshot: in.StatsSnapshot,
	})
	if err != nil {
		return handleErr(c, err)
	}
	return c.JSON(http.StatusOK, r)
}

func (h *handlers) deleteReview(c echo.Context) error {
	if err := h.svc.DeleteReview(c.Param("week_start")); err != nil {
		return handleErr(c, err)
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *handlers) snapshot(c echo.Context) error {
	s, err := h.svc.ComputeSnapshot(c.Param("week_start"))
	if err != nil {
		return handleErr(c, err)
	}
	return c.JSON(http.StatusOK, s)
}

// ---- export/import ----

func (h *handlers) exportData(c echo.Context) error {
	p, err := h.svc.Export()
	if err != nil {
		return handleErr(c, err)
	}
	return c.JSON(http.StatusOK, p)
}

func (h *handlers) importData(c echo.Context) error {
	var p service.ExportPayload
	if err := c.Bind(&p); err != nil {
		return fail(c, http.StatusBadRequest, "invalid_input", "请求体格式错误")
	}
	if err := h.svc.ImportAll(&p); err != nil {
		return handleErr(c, err)
	}
	return c.JSON(http.StatusOK, map[string]any{
		"imported": true,
		"goals":    len(p.Goals),
		"tasks":    len(p.Tasks),
		"blocks":   len(p.Blocks),
		"reviews":  len(p.Reviews),
	})
}

// ---- settings ----

func (h *handlers) getSettings(c echo.Context) error {
	m, err := h.svc.AllSettings()
	if err != nil {
		return handleErr(c, err)
	}
	return c.JSON(http.StatusOK, m)
}

func (h *handlers) putSettings(c echo.Context) error {
	var m map[string]string
	if err := c.Bind(&m); err != nil {
		return fail(c, http.StatusBadRequest, "invalid_input", "请求体格式错误")
	}
	for k, v := range m {
		if err := h.svc.SetSetting(k, v); err != nil {
			return handleErr(c, err)
		}
	}
	return c.JSON(http.StatusOK, m)
}
