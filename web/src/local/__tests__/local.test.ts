import { beforeEach, describe, expect, it } from 'vitest'
import 'fake-indexeddb/auto'
import { db } from '@/db/db'
import * as goals from '@/local/goals'
import * as tasks from '@/local/tasks'
import * as blocks from '@/local/blocks'
import * as stats from '@/local/stats'
import * as reviews from '@/local/reviews'
import * as exp from '@/local/export'
import * as settings from '@/local/settings'
import { LocalError } from '@/local/errors'
import { today, addDays, weekdayOf } from '@/local/time'
import type { Goal, Task, TaskInput } from '@/types/api'

function currentMonday(): string {
  let d = today()
  while (weekdayOf(d) !== 1) d = addDays(d, -1)
  return d
}

function isInvalid(e: unknown): boolean {
  return e instanceof LocalError && e.code === 'invalid_input'
}
function isConflict(e: unknown): boolean {
  return e instanceof LocalError && e.code === 'conflict'
}
function isNotFound(e: unknown): boolean {
  return e instanceof LocalError && e.code === 'not_found'
}

async function mkGoal(
  title: string, goalType: 'day' | 'week', goalMode: 'daily' | 'once',
  date: string | null, weekStart: string | null, checkinDays: number,
): Promise<Goal> {
  return goals.createGoal({ title, goal_type: goalType, goal_mode: goalMode, date, week_start: weekStart, checkin_days: checkinDays })
}

async function mkTask(title: string, due: string | null): Promise<Task> {
  return tasks.createTask({ title, due_date: due })
}

async function countEvents(taskID: number, typ: string): Promise<number> {
  return db.task_events.where('task_id').equals(taskID).filter((e) => e.event_type === typ).count()
}

beforeEach(async () => {
  await db.delete()
  await db.open()
})

describe('goals', () => {
  it('日目标打卡生命周期：重复→conflict，错日期→invalid，撤销后可再打', async () => {
    const d = today()
    const g = await mkGoal('背单词', 'day', 'daily', d, null, 5)
    const ck = await goals.checkin(g.id, '')
    expect(ck.checkin_date).toBe(d)
    await expect(goals.checkin(g.id, '')).rejects.toSatisfy(isConflict)
    await expect(goals.checkin(g.id, addDays(d, 1))).rejects.toSatisfy(isInvalid)
    await goals.uncheckin(g.id, '')
    await expect(goals.checkin(g.id, '')).resolves.toBeTruthy()
  })

  it('周目标（打卡型）日期边界校验', async () => {
    const ws = currentMonday()
    const g = await mkGoal('刷题 5 次', 'week', 'daily', null, ws, 5)
    await expect(goals.checkin(g.id, ws)).resolves.toBeTruthy()
    await expect(goals.checkin(g.id, addDays(ws, 6))).resolves.toBeTruthy()
    await expect(goals.checkin(g.id, addDays(ws, 7))).rejects.toSatisfy(isInvalid)
    await expect(goals.checkin(g.id, addDays(ws, -1))).rejects.toSatisfy(isInvalid)
    await expect(goals.uncheckin(g.id, '')).rejects.toSatisfy(isInvalid)
    await expect(goals.uncheckin(g.id, ws)).resolves.toBeUndefined()
  })

  it('一次性周目标：不可打卡，通过状态完成/重开', async () => {
    const ws = currentMonday()
    const g = await mkGoal('完成课程项目', 'week', 'once', null, ws, 5)
    await expect(goals.checkin(g.id, ws)).rejects.toSatisfy(isInvalid)
    const done = await goals.updateGoal(g.id, { title: g.title, goal_type: 'week', goal_mode: 'once', week_start: ws, checkin_days: 5, status: 'done' })
    expect(done.completed_at).toBeTruthy()
    const active = await goals.updateGoal(g.id, { title: g.title, goal_type: 'week', goal_mode: 'once', week_start: ws, checkin_days: 5, status: 'active' })
    expect(active.completed_at).toBeNull()
  })

  it('区间打卡查询', async () => {
    const ws = currentMonday()
    const g = await mkGoal('周目标', 'week', 'daily', null, ws, 5)
    for (const d of [ws, addDays(ws, 2), addDays(ws, 5)]) await goals.checkin(g.id, d)
    const list = await goals.listCheckins(addDays(ws, 1), addDays(ws, 4))
    expect(list).toHaveLength(1)
    expect(list[0].checkin_date).toBe(addDays(ws, 2))
    await expect(goals.listCheckins('bad', ws)).rejects.toSatisfy(isInvalid)
  })

  it('目标校验：非周一/日目标带周键/日目标 once/天数越界/归档不可打卡', async () => {
    const ws = currentMonday()
    await expect(mkGoal('x', 'week', 'daily', null, addDays(ws, 2), 5)).rejects.toSatisfy(isInvalid)
    const d = today()
    await expect(mkGoal('x', 'day', 'daily', d, ws, 5)).rejects.toSatisfy(isInvalid)
    await expect(mkGoal('x', 'day', 'once', d, null, 5)).rejects.toSatisfy(isInvalid)
    await expect(mkGoal('x', 'week', 'daily', null, ws, 8)).rejects.toSatisfy(isInvalid)
    const g = await mkGoal('归档', 'week', 'daily', null, ws, 5)
    await goals.updateGoal(g.id, { title: '归档', goal_type: 'week', goal_mode: 'daily', week_start: ws, checkin_days: 5, status: 'archived' })
    await expect(goals.checkin(g.id, ws)).rejects.toSatisfy(isInvalid)
  })
})

describe('tasks', () => {
  it('任务树环与自挂预防', async () => {
    const root = await mkTask('根任务', null)
    const child = await mkTask('子任务', null)
    const grand = await mkTask('孙任务', null)
    const upd = (id: number, parent: number) => tasks.updateTask(id, { title: 't', parent_id: parent })
    await expect(upd(child.id, root.id)).resolves.toBeTruthy()
    await expect(upd(grand.id, child.id)).resolves.toBeTruthy()
    await expect(upd(root.id, grand.id)).rejects.toSatisfy(isInvalid)
    await expect(upd(root.id, root.id)).rejects.toSatisfy(isInvalid)
  })

  it('批量分解：去空行、去空白、按序挂父', async () => {
    const root = await mkTask('学习 Go', null)
    const list = await tasks.batchTasks(root.id, null, [' 语法 ', '', '并发', '标准库 '])
    expect(list).toHaveLength(3)
    expect(list[0].title).toBe('语法')
    expect(list[0].parent_id).toBe(root.id)
    await expect(tasks.batchTasks(root.id, null, ['   '])).rejects.toSatisfy(isInvalid)
  })

  it('完成/重开幂等事件', async () => {
    const t = await mkTask('任务', null)
    expect(await countEvents(t.id, 'created')).toBe(1)
    const done = await tasks.completeTask(t.id)
    expect(done.status).toBe('done')
    expect(done.completed_at).toBeTruthy()
    await tasks.completeTask(t.id)
    expect(await countEvents(t.id, 'done')).toBe(1)
    const re = await tasks.reopenTask(t.id)
    expect(re.status).toBe('todo')
    expect(re.completed_at).toBeNull()
    await tasks.reopenTask(t.id)
    expect(await countEvents(t.id, 'reopen')).toBe(1)
  })

  it('删除任务级联：子任务与事件一并删除，日程块解除关联', async () => {
    const root = await mkTask('根', null)
    const child = await tasks.createTask({ title: '子', parent_id: root.id })
    const b = await blocks.createBlock({ block_date: today(), start_min: 600, end_min: 660, title: '块', task_id: child.id })
    await tasks.deleteTask(root.id)
    await expect(tasks.getTask(child.id)).rejects.toSatisfy(isNotFound)
    await expect(tasks.getTask(root.id)).rejects.toSatisfy(isNotFound)
    const bAfter = await blocks.getBlock(b.id)
    expect(bAfter.task_id).toBeNull()
  })
})

describe('blocks', () => {
  it('校验与 CRUD', async () => {
    const d = today()
    await expect(blocks.createBlock({ block_date: d, start_min: 600, end_min: 600, title: 'x' })).rejects.toSatisfy(isInvalid)
    await expect(blocks.createBlock({ block_date: '2026/10/16', start_min: 600, end_min: 660, title: 'x' })).rejects.toSatisfy(isInvalid)
    const b = await blocks.createBlock({ block_date: d, start_min: 600, end_min: 660, title: '刷题' })
    await expect(blocks.updateBlock(b.id, { block_date: d, start_min: 600, end_min: 660, title: '刷题', task_id: 99999 })).rejects.toSatisfy(isInvalid)
    const b2 = await blocks.updateBlock(b.id, { block_date: d, start_min: 900, end_min: 960, title: '刷题' })
    expect(b2.start_min).toBe(900)
    expect(b2.end_min).toBe(960)
    await blocks.deleteBlock(b.id)
    await expect(blocks.getBlock(b.id)).rejects.toSatisfy(isNotFound)
  })
})

describe('stats', () => {
  it('日统计口径', async () => {
    const d = today()
    const a = await mkTask('任务A', d)
    await mkTask('任务B', d)
    await tasks.completeTask(a.id)
    const g1 = await mkGoal('目标1', 'day', 'daily', d, null, 5)
    await mkGoal('目标2', 'day', 'daily', d, null, 5)
    await goals.checkin(g1.id, '')
    await blocks.createBlock({ block_date: d, start_min: 1200, end_min: 1260, title: '学习', task_id: a.id })
    const st = await stats.dailyStats(d)
    expect(st.task_total).toBe(2)
    expect(st.task_done).toBe(1)
    expect(st.task_completion_rate).toBeCloseTo(0.5)
    expect(st.goal_total).toBe(2)
    expect(st.goal_done).toBe(1)
    expect(st.goal_achievement_rate).toBeCloseTo(0.5)
    expect(st.scheduled_minutes).toBe(60)
    expect(st.completed_block_minutes).toBe(60)
  })

  it('周统计口径与趋势', async () => {
    const ws = currentMonday()
    const ids: number[] = []
    for (let i = 0; i < 4; i++) ids.push((await mkTask('周任务' + String.fromCharCode(65 + i), addDays(ws, i))).id)
    for (let i = 0; i < 3; i++) await tasks.completeTask(ids[i])
    const g = await mkGoal('周目标', 'week', 'daily', null, ws, 5)
    await goals.checkin(g.id, ws)
    await goals.checkin(g.id, addDays(ws, 1))
    const w = await stats.weeklyStats(ws)
    expect(w.task_total).toBe(4)
    expect(w.task_done).toBe(3)
    expect(w.task_completion_rate).toBeCloseTo(0.75)
    expect(w.goal_achievement_rate!).toBeGreaterThan(0.39)
    expect(w.goal_achievement_rate!).toBeLessThan(0.41)
    expect(w.trend).toHaveLength(8)
    expect(w.unfinished).toHaveLength(1)
  })

  it('热力图加权：默认 0.5/0.5 → 0.75；权重 1/0 → 0.5', async () => {
    const d = today()
    const a = await mkTask('A', d)
    await mkTask('B', d)
    await tasks.completeTask(a.id)
    const g = await mkGoal('G', 'day', 'daily', d, null, 5)
    await goals.checkin(g.id, '')
    const hm = await stats.heatmap(d, d)
    expect(hm.cells).toHaveLength(1)
    expect(hm.cells[0].score!).toBeGreaterThan(0.74)
    expect(hm.cells[0].score!).toBeLessThan(0.76)
    await settings.setSetting('heatmap_task_weight', '1')
    await settings.setSetting('heatmap_goal_weight', '0')
    const hm2 = await stats.heatmap(d, d)
    expect(hm2.cells[0].score).toBeCloseTo(0.5)
  })

  it('逾期统计', async () => {
    await mkTask('逾期任务', addDays(today(), -1))
    const st = await stats.dailyStats(today())
    expect(st.overdue_count).toBe(1)
  })

  it('今日面板：清单与块完成标记', async () => {
    const d = today()
    await mkGoal('目标', 'day', 'daily', d, null, 5)
    const task = await mkTask('任务', d)
    await blocks.createBlock({ block_date: d, start_min: 540, end_min: 600, title: '晨读', task_id: task.id })
    const dash = await stats.dashboard(d)
    expect(dash.day_goals).toHaveLength(1)
    expect(dash.tasks_due).toHaveLength(1)
    expect(dash.blocks).toHaveLength(1)
    expect(dash.blocks[0].done).toBe(false)
    await tasks.completeTask(task.id)
    const dash2 = await stats.dashboard(d)
    expect(dash2.blocks[0].done).toBe(true)
  })
})

describe('reviews', () => {
  it('复盘快照冻结与周键校验', async () => {
    const ws = currentMonday()
    const task = await mkTask('复盘前任务', ws)
    const snap = await stats.computeSnapshot(ws)
    const r = await reviews.upsertReview(ws, { title: '第 N 周复盘', content_md: 'KPT 内容', stats_snapshot: snap })
    const frozen = JSON.stringify(r.stats_snapshot)
    expect(frozen.length).toBeGreaterThan(2)
    await tasks.completeTask(task.id)
    const got = await reviews.getReview(ws)
    expect(JSON.stringify(got.stats_snapshot)).toBe(frozen)
    await expect(reviews.upsertReview(addDays(ws, 1), {})).rejects.toSatisfy(isInvalid)
  })
})

describe('export', () => {
  it('导出导入往返无损', async () => {
    const ws = currentMonday()
    const d = today()
    const g1 = await mkGoal('日目标', 'day', 'daily', d, null, 5)
    await mkGoal('周目标', 'week', 'daily', null, ws, 5)
    await goals.checkin(g1.id, '')
    const task = await mkTask('任务', d)
    await tasks.completeTask(task.id)
    await blocks.createBlock({ block_date: d, start_min: 600, end_min: 660, title: '块', task_id: task.id })
    await reviews.upsertReview(ws, { title: '复盘' })
    await settings.setSetting('theme', 'dark')

    const payload = await exp.exportAll()
    expect(payload.goals).toHaveLength(2)
    expect(payload.goal_checkins).toHaveLength(1)
    expect(payload.tasks).toHaveLength(1)
    expect(payload.task_events.length).toBeGreaterThanOrEqual(2)
    expect(payload.blocks).toHaveLength(1)
    expect(payload.reviews).toHaveLength(1)
    expect(payload.settings).toHaveLength(1)

    await exp.importAll(payload)
    const after = await exp.exportAll()
    expect(after.goals).toHaveLength(payload.goals.length)
    expect(after.tasks).toHaveLength(payload.tasks.length)
    expect(after.goal_checkins).toHaveLength(payload.goal_checkins.length)
    expect(after.goals[0].id).toBe(payload.goals[0].id)
    await expect(tasks.getTask(task.id)).resolves.toBeTruthy()
    await expect(exp.importAll({ version: '', exported_at: '', goals: [], goal_checkins: [], tasks: [], task_events: [], blocks: [], reviews: [], settings: [] })).rejects.toSatisfy(isInvalid)
  })

  it('自定义周起始日（周日）', async () => {
    await settings.setSetting('week_start_day', '0')
    const ws = currentMonday()
    let sunday = ws
    while (weekdayOf(sunday) !== 0) sunday = addDays(sunday, -1)
    await expect(mkGoal('x', 'week', 'daily', null, sunday, 5)).resolves.toBeTruthy()
    await expect(mkGoal('x', 'week', 'daily', null, addDays(sunday, 1), 5)).rejects.toSatisfy(isInvalid)
    await expect(stats.weeklyStats(sunday)).resolves.toBeTruthy()
  })
})

describe('regression', () => {
  it('响应式 Proxy 入参可正常写入（DataCloneError 回归）', async () => {
    // 模拟 Vue reactive 表单：整个对象及 tags 数组都是 Proxy
    const form = new Proxy({ title: '代理任务', note: '', tags: ['vue'], due_date: null, est_minutes: null, parent_id: null, goal_id: null }, {})
    // 前置断言：该对象确实无法被浏览器/Node 原生结构化克隆（否则测不出回归）
    expect(() => structuredClone(form)).toThrow()
    const t = await tasks.createTask(form as unknown as TaskInput)
    expect(t.title).toBe('代理任务')
    const got = await tasks.getTask(t.id)
    expect(got.tags).toEqual(['vue'])

    const gform = new Proxy({ title: '代理目标', goal_type: 'day', goal_mode: 'daily', date: today(), week_start: null, checkin_days: 5 }, {})
    const g = await goals.createGoal(gform as unknown as Parameters<typeof goals.createGoal>[0])
    expect(g.title).toBe('代理目标')
  })
})

