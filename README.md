# StudyPilot · 学习任务计划系统

面向个人学习者的「目标 → 任务分解 → 日程 → 统计 → 周复盘」闭环工具。
纯前端应用：数据全部保存在浏览器（IndexedDB），无需服务器，部署到 GitHub Pages 即用，人人独立数据。

> 原服务器版（Go + SQLite 单二进制）保留在 `main` 分支；本分支 `codex/pure-frontend` 为纯前端版（v0.2.0）。

## 功能

- **学习目标**：日目标打卡（未达成徽章）、周目标双模式（每日打卡 / 一次性）、达成率实时计算、任务关联参考口径
- **任务分解**：任意层级任务树、批量分解（粘贴多行一键生成子任务）、父任务进度条、一键/自动完成父任务、组合筛选、拖拽排序、逾期标红
- **日程管理**：24h 时间轴（当前时间线）、点击创建、拖动平移 / 边缘缩放、任务拖入生成时间块、完成高亮、周概览
- **统计**：今日面板、周视图总览、完成度热力图（点击下钻）、近 8 周趋势折线、按目标/标签分布、未完成清单；权重可配置
- **周复盘**：Markdown 编辑器、KPT 模板、统计快照自动注入与冻结（改动数据不影响历史复盘）、草稿自动保存
- **其它**：浅色/深色渐变主题、桌面逾期通知、JSON 全量导出/导入（往返无损）、PWA 可安装

## 快速开始

```powershell
# 开发
cd web; pnpm install; pnpm dev        # http://localhost:5173

# 测试与构建
pnpm test                             # 18 个口径/规则测试
pnpm build                            # 产物在 web/dist（任意静态托管可部署）
```

## 部署到 GitHub Pages

1. 推送 `main` 分支 → Actions 自动运行 `pages.yml` 构建并部署
2. 仓库 Settings → Pages → Source 选 `GitHub Actions`
3. 首次部署后获得地址：`https://<用户名>.github.io/study-planner/`

## 技术栈

- 前端：Vue 3 + TypeScript + Vite 8 + Pinia + Element Plus + ECharts + md-editor-v3 + PWA
- 存储：IndexedDB（Dexie.js，7 张表，与 Go 版 SQLite schema 一一对应）
- 业务层：`web/src/local/`（目标/任务/日程/统计/复盘/导出，与 Go 版 service 层逐条对账，18 个测试全绿）
- 测试：Vitest + fake-indexeddb

## 文档

- [落地方案](docs/落地方案.md) · [需求明细](docs/需求明细.md) · [开发进度](docs/开发进度.md)

## 数据迁移（从 Go 版到纯前端版）

1. 打开 Go 版「设置 → 数据 → 导出 JSON」得到备份文件
2. 打开纯前端版「设置 → 数据 → 导入 JSON（覆盖）」
3. 全部目标/任务/日程/复盘/设置原样恢复（保留 ID 与关联关系）
