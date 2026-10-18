import { db, type BlockRow, type GoalRow, type TaskRow } from '@/db/db'
import { errInvalid } from './errors'
import { today, validDate, addDays, weekEndDate, isWeekStart, weekdayOf } from './time'
import { getSettingFloat, getWeekStartDay } from './settings'
import type {
  BlockView, Dashboard, DistItem, Heatmap, HeatmapCell, Snapshot, StatsDaily,
  StatsWeekly, Task, TrendPoint,
} from '@/types/api'

// ===== 与 Go 版 service/stats.go 对齐（统计口径冻结） =====

function rate(total: number, done: number): number | null {
  if (total <= 0) return null
  return done / total
}

async function countTasksWhere(pred: (t: TaskRow) => boolean): Promise<number> {
  const all = await db.tasks.toArray()
  return all.filter(pred).length
}

async function tasksWhere(pred: (t: TaskRow) => boolean): Promise<TaskRow[]> {
  const all = await db.tasks.toArray()
  return all.filter(pred)
}

async function countDayGoals(date: string, excludeArchived: boolean): Promise<number> {
  let coll = db.goals.where('date').equals(date).filter((g) => g.goal_type === 'day')
  if (excludeArchived) coll = coll.filter((g) => g.status !== 'archived')
  return coll.count()
}

async function countDayGoalCheckins(date: string): Promise<number> {
  const goals = await db.goals.where('date').equals(date).filter((g) => g.goal_type === 'day' && g.status !== 'archived').toArray()
  const ids = new Set(goals.map((g) => g.id!))
  if (ids.size === 0) return 0
  const cks = await db.goal_checkins.where('checkin_date').equals(date).toArray()
  return cks.filter((c) => ids.has(c.goal_id)).length
}

async function sumBlockMinutes(pred: (b: BlockRow) => boolean): Promise<number> {
  const blocks = await db.blocks.toArray()
  return blocks.filter(pred).reduce((s, b) => s + (b.end_min - b.start_min), 0)
}

// ---- 日统计 ----
export async function dailyStats(date: string): Promise<StatsDaily> {
  if (!validDate(date)) throw errInvalid('date 格式非法')
  const out: StatsDaily = {
    date,
    task_completion_rate: null,
    goal_achievement_rate: null,
    scheduled_minutes: 0,
    completed_block_minutes: 0,
    overdue_count: 0,
    task_total: 0,
    task_done: 0,
    goal_total: 0,
    goal_done: 0,
  }
  // 任务口径：due_date = 当日
  out.task_total = await countTasksWhere((t) => t.due_date === date)
  out.task_done = await countTasksWhere((t) => t.due_date === date && t.status === 'done')
  out.task_completion_rate = rate(out.task_total, out.task_done)
  // 目标口径：当日日目标（排除归档）的打卡率
  out.goal_total = await countDayGoals(date, true)
  out.goal_done = await countDayGoalCheckins(date)
  out.goal_achievement_rate = rate(out.goal_total, out.goal_done)
  // 日程口径
  out.scheduled_minutes = await sumBlockMinutes((b) => b.block_date === date)
  const doneTaskIds = new Set((await tasksWhere((t) => t.status === 'done')).map((t) => t.id!))
  out.completed_block_minutes = await sumBlockMinutes(
    (b) => b.block_date === date && b.task_id != null && doneTaskIds.has(b.task_id),
  )
  // 逾期：以查询日为准（due < 查询日 且未完成）
  out.overdue_count = await countTasksWhere((t) => t.due_date != null && t.due_date < date && t.status !== 'done')
  return out
}

// ---- 周统计 ----
async function weekTaskStats(weekStart: string): Promise<{ total: number; done: number; overdue: number }> {
  const we = weekEndDate(weekStart)
  const weekTasks = await tasksWhere((t) => t.due_date != null && t.due_date >= weekStart && t.due_date <= we)
  const total = weekTasks.length
  const done = weekTasks.filter((t) => t.status === 'done').length
  const overdue = weekTasks.filter((t) => t.due_date! < today() && t.status !== 'done').length
  return { total, done, overdue }
}

async function weekGoalStats(weekStart: string): Promise<{ total: number; done: number; rateV: number | null }> {
  const we = weekEndDate(weekStart)
  const goals = await db.goals
    .where('week_start').equals(weekStart)
    .filter((g) => g.goal_type === 'week' && g.status !== 'archived')
    .toArray()
  const total = goals.length
  if (total === 0) return { total: 0, done: 0, rateV: null }
  let done = 0
  let sum = 0
  for (const g of goals) {
    if (g.status === 'done') done++
    if (g.goal_mode === 'once') {
      if (g.status === 'done') sum += 1
    } else {
      const n = await db.goal_checkins
        .where('goal_id').equals(g.id!)
        .filter((c) => c.checkin_date >= weekStart && c.checkin_date <= we)
        .count()
      let r = n / g.checkin_days
      if (r > 1) r = 1
      sum += r
    }
  }
  return { total, done, rateV: sum / total }
}

export async function weeklyStats(weekStart: string): Promise<StatsWeekly> {
  if (!validDate(weekStart) || !isWeekStart(weekStart, await getWeekStartDay())) {
    throw errInvalid('week_start 必须是周一日期')
  }
  const we = weekEndDate(weekStart)
  const out: StatsWeekly = {
    week_start: weekStart,
    week_end: we,
    task_completion_rate: null,
    goal_achievement_rate: null,
    task_total: 0,
    task_done: 0,
    overdue_count: 0,
    goal_total: 0,
    goal_done: 0,
    scheduled_minutes: 0,
    completed_block_minutes: 0,
    trend: [],
    by_goal: [],
    by_tag: [],
    unfinished: [],
  }

  const tk = await weekTaskStats(weekStart)
  out.task_total = tk.total
  out.task_done = tk.done
  out.overdue_count = tk.overdue
  out.task_completion_rate = rate(tk.total, tk.done)

  const g = await weekGoalStats(weekStart)
  out.goal_total = g.total
  out.goal_done = g.done
  out.goal_achievement_rate = g.rateV

  out.scheduled_minutes = await sumBlockMinutes((b) => b.block_date >= weekStart && b.block_date <= we)
  const doneTaskIds = new Set((await tasksWhere((t) => t.status === 'done')).map((t) => t.id!))
  out.completed_block_minutes = await sumBlockMinutes(
    (b) => b.block_date >= weekStart && b.block_date <= we && b.task_id != null && doneTaskIds.has(b.task_id),
  )

  // 近 8 周趋势（旧 → 新）
  const trendStart = addDays(weekStart, -7 * 7)
  for (let i = 0; i < 8; i++) {
    const ws = addDays(trendStart, 7 * i)
    const tt = await weekTaskStats(ws)
    const tg = await weekGoalStats(ws)
    out.trend.push({ week_start: ws, task_rate: rate(tt.total, tt.done), goal_rate: tg.rateV })
  }

  // 任务分布：按目标 / 按标签 / 未完成
  const weekTasks = (await tasksWhere((t) => t.due_date != null && t.due_date >= weekStart && t.due_date <= we)).sort(
    (a, b) => a.sort - b.sort || a.id! - b.id!,
  )
  const goalTitles = new Map<number, string>()
  for (const t of weekTasks) {
    if (t.goal_id != null && !goalTitles.has(t.goal_id)) {
      const goal = await db.goals.get(t.goal_id)
      goalTitles.set(t.goal_id, goal?.title ?? `#${t.goal_id}`)
    }
  }
  const byGoal = new Map<string, DistItem>()
  const byTag = new Map<string, DistItem>()
  for (const t of weekTasks) {
    const key = t.goal_id != null ? goalTitles.get(t.goal_id)! : '未关联目标'
    const it = byGoal.get(key) ?? { label: key, done: 0, total: 0 }
    it.total++
    if (t.status === 'done') it.done++
    byGoal.set(key, it)
    for (const tag of t.tags ?? []) {
      const ti = byTag.get(tag) ?? { label: tag, done: 0, total: 0 }
      ti.total++
      if (t.status === 'done') ti.done++
      byTag.set(tag, ti)
    }
  }
  out.by_goal = [...byGoal.values()]
  out.by_tag = [...byTag.values()]
  out.unfinished = weekTasks.filter((t) => t.status !== 'done').map(rowToTask)
  return out
}

// ---- 热力图 ----
export async function heatmap(start: string, end: string): Promise<Heatmap> {
  if (!validDate(start) || !validDate(end)) throw errInvalid('start/end 格式非法')
  if (end < start) throw errInvalid('end 必须不早于 start')
  const sd = new Date(start + 'T00:00:00')
  const ed = new Date(end + 'T00:00:00')
  const days = Math.round((ed.getTime() - sd.getTime()) / 86400000) + 1
  if (days > 366) throw errInvalid('热力图范围不能超过 366 天')
  const taskW = await getSettingFloat('heatmap_task_weight', 0.5)
  const goalW = await getSettingFloat('heatmap_goal_weight', 0.5)
  const cells: HeatmapCell[] = []
  for (let i = 0; i < days; i++) {
    const date = addDays(start, i)
    const cell: HeatmapCell = {
      date,
      score: null,
      task_rate: null,
      goal_rate: null,
      task_total: 0,
      task_done: 0,
      goal_total: 0,
      goal_done: 0,
    }
    cell.task_total = await countTasksWhere((t) => t.due_date === date)
    cell.task_done = await countTasksWhere((t) => t.due_date === date && t.status === 'done')
    cell.task_rate = rate(cell.task_total, cell.task_done)
    cell.goal_total = await countDayGoals(date, true)
    cell.goal_done = await countDayGoalCheckins(date)
    cell.goal_rate = rate(cell.goal_total, cell.goal_done)
    // 综合分：可用口径的加权平均；两者皆无则 null
    if (cell.task_rate !== null || cell.goal_rate !== null) {
      let sum = 0
      let wsum = 0
      if (cell.task_rate !== null) {
        sum += taskW * cell.task_rate
        wsum += taskW
      }
      if (cell.goal_rate !== null) {
        sum += goalW * cell.goal_rate
        wsum += goalW
      }
      cell.score = sum / wsum
    }
    cells.push(cell)
  }
  return { start, end, cells }
}

// ---- 今日面板 ----
function rowToTask(r: TaskRow): Task {
  return { ...r, id: r.id!, tags: r.tags ?? [] }
}

export async function dashboard(date: string): Promise<Dashboard> {
  const d = await dailyStats(date)
  const out: Dashboard = {
    date,
    task_completion_rate: d.task_completion_rate,
    goal_achievement_rate: d.goal_achievement_rate,
    scheduled_minutes: d.scheduled_minutes,
    overdue_count: d.overdue_count,
    day_goals: [],
    tasks_due: [],
    blocks: [],
  }
  const goals = await db.goals.where('date').equals(date).filter((g) => g.goal_type === 'day' && g.status !== 'archived').toArray()
  goals.sort((a, b) => a.sort - b.sort || a.id! - b.id!)
  out.day_goals = goals.map((g) => ({ ...g, id: g.id! }))
  const tasks = await db.tasks.where('due_date').equals(date).toArray()
  tasks.sort((a, b) => a.priority - b.priority || a.sort - b.sort || a.id! - b.id!)
  out.tasks_due = tasks.map(rowToTask)
  const blocks = await db.blocks.where('block_date').equals(date).toArray()
  blocks.sort((a, b) => a.start_min - b.start_min || a.id! - b.id!)
  const doneIds = new Set((await tasksWhere((t) => t.status === 'done')).map((t) => t.id!))
  out.blocks = blocks.map((b) => ({ ...b, id: b.id!, done: b.task_id != null && doneIds.has(b.task_id) })) as BlockView[]
  return out
}

// ---- 周复盘快照（ComputeSnapshot） ----
export async function computeSnapshot(weekStart: string): Promise<Snapshot> {
  const w = await weeklyStats(weekStart)
  const we = weekEndDate(weekStart)
  // 打卡统计：打卡型周目标在绑定周内的打卡总数
  const weekGoals = await db.goals
    .where('week_start').equals(weekStart)
    .filter((g) => g.goal_type === 'week' && g.status !== 'archived')
    .toArray()
  let checkinTotal = 0
  for (const g of weekGoals) {
    if (g.goal_mode === 'once') continue
    checkinTotal += await db.goal_checkins
      .where('goal_id').equals(g.id!)
      .filter((c) => c.checkin_date >= weekStart && c.checkin_date <= we)
      .count()
  }
  let checkinDaysTotal = 0
  for (const g of weekGoals) {
    if (g.goal_mode === 'daily') checkinDaysTotal += g.checkin_days
  }
  const now = new Date()
  const pad = (n: number) => String(n).padStart(2, '0')
  return {
    generated_at: `${today()} ${pad(now.getHours())}:${pad(now.getMinutes())}:${pad(now.getSeconds())}`,
    week_start: weekStart,
    week_end: we,
    task_completion_rate: w.task_completion_rate,
    goal_achievement_rate: w.goal_achievement_rate,
    task_total: w.task_total,
    task_done: w.task_done,
    overdue_count: w.overdue_count,
    goal_total: w.goal_total,
    goal_done: w.goal_done,
    checkin_total: checkinTotal,
    checkin_days_total: checkinDaysTotal,
    scheduled_minutes: w.scheduled_minutes,
    completed_block_minutes: w.completed_block_minutes,
    unfinished_tasks: w.unfinished.map((t) => t.title),
    by_tag: w.by_tag,
  }
}


