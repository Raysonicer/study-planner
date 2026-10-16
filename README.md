# StudyPilot · 学习任务计划系统

面向个人学习者的「目标 → 任务分解 → 日程 → 统计 → 周复盘」闭环工具。
单机优先：SQLite 单文件，单二进制 / Docker 部署，Web 界面（PWA 可安装）。

## 功能

- **学习目标**：日目标打卡（未达成徽章）、周目标双模式（每日打卡 / 一次性）、达成率实时计算、任务关联参考口径
- **任务分解**：任意层级任务树、批量分解（粘贴多行一键生成子任务）、父任务进度条、一键/自动完成父任务、组合筛选、拖拽排序、逾期标红
- **日程管理**：24h 时间轴（当前时间线）、点击创建、拖动平移 / 边缘缩放、任务拖入生成时间块、完成高亮、周概览
- **统计**：今日面板、GitHub 风格完成度热力图（点击下钻）、近 8 周趋势折线、按目标/标签分布、未完成清单；权重可配置
- **周复盘**：Markdown 编辑器、KPT 模板、统计快照自动注入与冻结（改动数据不影响历史复盘）、草稿自动保存
- **其它**：浅色/深色渐变主题、桌面逾期通知、JSON 全量导出/导入（往返无损）、启动自动备份（保留 14 份）、Swagger UI（开发模式 /api/docs）

## 快速开始

```powershell
# 开发
cd server; go run ./cmd/server          # :8080
cd web; pnpm install; pnpm dev          # :5173（/api 代理）

# 一键构建单二进制
.\build.ps1
.\bin\studypilot.exe                    # http://localhost:8080

# Docker（多阶段构建）
cd deploy; docker compose up -d --build
```

## 技术栈

- 后端：Go 1.27 + Echo + GORM + SQLite（纯 Go 驱动，免 CGO），版本化迁移内嵌执行
- 前端：Vue 3 + TypeScript + Vite 8 + Pinia + Element Plus + ECharts（按需引入）+ md-editor-v3 + PWA
- 部署：前端产物 embed 进单二进制；Docker 多阶段构建；数据目录外挂

## 文档

- [落地方案](docs/落地方案.md) · [需求明细](docs/需求明细.md) · [OpenAPI](docs/openapi.yaml) · [开发进度](docs/开发进度.md) · [迁移 SQL](server/internal/store/migrations/000001_init.up.sql)

## 环境变量

| 变量 | 默认 | 说明 |
|---|---|---|
| SP_PORT | 8080 | 监听端口 |
| SP_DATA_DIR | data | SQLite 数据目录（自动备份到 data/backups，保留 14 份） |
| SP_DOCS_DIR | docs | 文档目录（含 openapi.yaml 时启用 /api/docs Swagger UI） |

## 里程碑

M0 工程地基 ✅ M1 数据与 API ✅ M2 目标+任务 ✅ M3 日程 ✅ M4 统计 ✅ M5 复盘 ✅ M6 打磨发布 ✅