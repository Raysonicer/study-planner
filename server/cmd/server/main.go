package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	swgui "github.com/swaggest/swgui/v5emb"

	"studypilot/server/internal/api"
	"studypilot/server/internal/config"
	"studypilot/server/internal/frontend"
	"studypilot/server/internal/service"
	"studypilot/server/internal/store"
	"studypilot/server/internal/web"
)

func main() {
	cfg := config.Load()

	if err := os.MkdirAll(cfg.DataDir, 0o755); err != nil {
		panic(fmt.Sprintf("create data dir: %v", err))
	}
	dbPath := filepath.Join(cfg.DataDir, "app.db")

	// 启动备份（SYS-02）：迁移前先留底
	if err := store.Backup(dbPath, filepath.Join(cfg.DataDir, "backups"), 14); err != nil {
		panic(fmt.Sprintf("backup: %v", err))
	}

	db, err := store.Open(dbPath)
	if err != nil {
		panic(err)
	}
	if err := store.Migrate(db); err != nil {
		panic(fmt.Sprintf("migrate: %v", err))
	}
	svc := service.New(db)

	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	e.Use(middleware.Recover())
	e.Use(middleware.Logger())

	apiGroup := e.Group("/api/v1")
	apiGroup.GET("/healthz", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]any{
			"status":  "ok",
			"version": config.Version,
		})
	})
	api.Register(e, svc)

	// 开发环境提供 OpenAPI 设计文档与 Swagger UI（文档目录存在时启用）
	if _, err := os.Stat(filepath.Join(cfg.DocsDir, "openapi.yaml")); err == nil {
		e.Static("/api/openapi.yaml", filepath.Join(cfg.DocsDir, "openapi.yaml"))
		e.Any("/api/docs", echo.WrapHandler(swgui.NewHandler("StudyPilot API", "/api/openapi.yaml", "/api/docs")))
	}

	sub, err := frontend.Sub()
	if err != nil {
		panic(fmt.Sprintf("embed frontend: %v", err))
	}
	e.GET("/*", echo.WrapHandler(web.SPAHandler(sub)))

	e.Logger.Fatal(e.Start(":" + cfg.Port))
}
