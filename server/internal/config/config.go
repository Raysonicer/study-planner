package config

import "os"

// Version 与 docs/openapi.yaml 的 info.version 保持一致。
const Version = "0.1.0"

type Config struct {
	Port    string // 监听端口（SP_PORT，默认 8080）
	DataDir string // 数据目录（SP_DATA_DIR，默认 data）
	DocsDir string // 设计文档目录（SP_DOCS_DIR，默认 docs；开发环境用于暴露 OpenAPI）
}

func Load() Config {
	return Config{
		Port:    envOr("SP_PORT", "8080"),
		DataDir: envOr("SP_DATA_DIR", "data"),
		DocsDir: envOr("SP_DOCS_DIR", "docs"),
	}
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
