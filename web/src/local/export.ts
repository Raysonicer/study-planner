import { db, type BlockRow, type GoalCheckinRow, type GoalRow, type ReviewRow, type SettingRow, type TaskEventRow, type TaskRow } from '@/db/db'
import { errInvalid } from './errors'
import { nowString } from './time'

// ===== 与 Go 版 service/export.go 对齐：导出/导入往返无损 =====

export interface ExportPayload {
  version: string
  exported_at: string
  goals: GoalRow[]
  goal_checkins: GoalCheckinRow[]
  tasks: TaskRow[]
  task_events: TaskEventRow[]
  blocks: BlockRow[]
  reviews: ReviewRow[]
  settings: SettingRow[]
}

export async function exportAll(): Promise<ExportPayload> {
  const [goals, checkins, tasks, taskEvents, blocks, reviews, settings] = await Promise.all([
    db.goals.toArray(),
    db.goal_checkins.toArray(),
    db.tasks.toArray(),
    db.task_events.toArray(),
    db.blocks.toArray(),
    db.reviews.toArray(),
    db.settings.toArray(),
  ])
  const byId = (a: { id?: number }, b: { id?: number }) => (a.id ?? 0) - (b.id ?? 0)
  settings.sort((a, b) => (a.key < b.key ? -1 : 1))
  return {
    version: '1',
    exported_at: nowString(),
    goals: goals.sort(byId),
    goal_checkins: checkins.sort(byId),
    tasks: tasks.sort(byId),
    task_events: taskEvents.sort(byId),
    blocks: blocks.sort(byId),
    reviews: reviews.sort(byId),
    settings,
  }
}

export async function importAll(p: ExportPayload): Promise<void> {
  if (!p.version) throw errInvalid('无效的导出文件（缺少 version）')
  const total =
    (p.goals?.length ?? 0) + (p.goal_checkins?.length ?? 0) + (p.tasks?.length ?? 0) +
    (p.task_events?.length ?? 0) + (p.blocks?.length ?? 0) + (p.reviews?.length ?? 0) + (p.settings?.length ?? 0)
  if (total > 200000) throw errInvalid('数据量过大')
  await db.transaction('rw', db.tables, async () => {
    await db.settings.clear()
    await db.goals.clear()
    await db.goal_checkins.clear()
    await db.tasks.clear()
    await db.task_events.clear()
    await db.blocks.clear()
    await db.reviews.clear()
    await db.goals.bulkAdd((p.goals ?? []).map((g) => ({ ...g })))
    await db.goal_checkins.bulkAdd(p.goal_checkins ?? [])
    await db.tasks.bulkAdd((p.tasks ?? []).map((t) => ({ ...t, tags: t.tags ?? [] })))
    await db.task_events.bulkAdd(p.task_events ?? [])
    await db.blocks.bulkAdd(p.blocks ?? [])
    await db.reviews.bulkAdd(p.reviews ?? [])
    await db.settings.bulkAdd(p.settings ?? [])
  })
}
