package api

import (
	"github.com/labstack/echo/v4"

	"studypilot/server/internal/service"
)

// Register 挂载全部 /api/v1 路由（顺序：字面路由先于参数路由）。
func Register(e *echo.Echo, svc *service.Service) {
	h := &handlers{svc: svc}
	api := e.Group("/api/v1")

	// goals
	api.GET("/checkins", h.listCheckins)
	api.GET("/goals", h.listGoals)
	api.POST("/goals", h.createGoal)
	api.GET("/goals/:id", h.getGoal)
	api.PUT("/goals/:id", h.updateGoal)
	api.DELETE("/goals/:id", h.deleteGoal)
	api.POST("/goals/:id/checkin", h.checkin)
	api.DELETE("/goals/:id/checkin", h.uncheckin)

	// tasks（batch/reorder 先于 :id 注册）
	api.GET("/tasks", h.listTasks)
	api.POST("/tasks", h.createTask)
	api.POST("/tasks/batch", h.batchTasks)
	api.PATCH("/tasks/reorder", h.reorderTasks)
	api.GET("/tasks/:id", h.getTask)
	api.PUT("/tasks/:id", h.updateTask)
	api.DELETE("/tasks/:id", h.deleteTask)
	api.POST("/tasks/:id/complete", h.completeTask)
	api.POST("/tasks/:id/reopen", h.reopenTask)

	// blocks
	api.GET("/blocks", h.listBlocks)
	api.POST("/blocks", h.createBlock)
	api.PUT("/blocks/:id", h.updateBlock)
	api.DELETE("/blocks/:id", h.deleteBlock)

	// dashboard & stats
	api.GET("/dashboard", h.dashboard)
	api.GET("/stats/daily", h.statsDaily)
	api.GET("/stats/weekly", h.statsWeekly)
	api.GET("/stats/heatmap", h.statsHeatmap)

	// reviews
	api.GET("/reviews", h.listReviews)
	api.GET("/reviews/:week_start", h.getReview)
	api.PUT("/reviews/:week_start", h.putReview)
	api.DELETE("/reviews/:week_start", h.deleteReview)
	api.POST("/reviews/:week_start/snapshot", h.snapshot)

	// export/import
	api.GET("/export", h.exportData)
	api.POST("/export/import", h.importData)

	// settings
	api.GET("/settings", h.getSettings)
	api.PUT("/settings", h.putSettings)
}
