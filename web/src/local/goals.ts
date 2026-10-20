import { db, type GoalRow, type GoalCheckinRow } from '@/db/db'
import { errConflict, errInvalid, errNotFound } from './errors'
import { plain } from './util'
import { today, nowString, validDate, weekEndDate, isWeekStart } from './time'
import { getWeekStartDay } from './settings'
import type { Goal, GoalCheckin, GoalInput } from '@/types/api'

// ===== 与 Go 版 service/goals.go 对齐 =====

export interface GoalFilter {
  goal_type?: string
  date?: string
  week_start?: string
  status?: string
}

type GoalParams = GoalInput

export async function listGoals(f: GoalFilter = {}): Promise<Goal[]> {
  let coll = db.goals.orderBy('sort')
  if (f.goal_type) coll = coll.filter((g) => g.goal_type === f.goal_type)
  if (f.date) coll = coll.filter((g) => g.date === f.date)
  if (f.week_start) coll = coll.filter((g) => g.week_start === f.week_start)
  const status = f.status || ''
  if (status === '') coll = coll.filter((g) => g.status !== 'archived')
  else coll = coll.filter((g) => g.status === status)
  const rows = await coll.sortBy('id')
  return rows.map(rowToGoal)
}

function rowToGoal(r: GoalRow): Goal {
  return { ...r, id: r.id! }
}

export async function getGoal(id: number): Promise<Goal> {
  const r = await db.goals.get(id)
  if (!r) throw errNotFound()
  return rowToGoal(r)
}

async function validateGoal(p: GoalParams): Promise<void> {
  if (!p.title) throw errInvalid('title 不能为空')
  switch (p.goal_type) {
    case 'day':
      if (!p.date || !validDate(p.date)) throw errInvalid('日目标需要合法的 date')
      if (p.week_start) throw errInvalid('日目标不能设置 week_start')
      break
    case 'week': {
      if (!p.week_start || !validDate(p.week_start)) throw errInvalid('周目标需要合法的 week_start')
      if (!isWeekStart(p.week_start, await getWeekStartDay())) throw errInvalid('week_start 必须是周一')
      if (p.date) throw errInvalid('周目标不能设置 date')
      break
    }
    default:
      throw errInvalid('goal_type 必须是 day/week')
  }
  switch (p.goal_mode ?? 'daily') {
    case 'daily':
      break
    case 'once':
      if (p.goal_type !== 'week') throw errInvalid('一次性模式仅周目标可用')
      break
    default:
      throw errInvalid('goal_mode 必须是 daily/once')
  }
  if (p.checkin_days !== undefined && (p.checkin_days < 1 || p.checkin_days > 7)) {
    throw errInvalid('checkin_days 必须在 1-7')
  }
  if (p.status !== undefined && !['active', 'done', 'archived'].includes(p.status)) {
    throw errInvalid('非法 status')
  }
}

function applyGoal(p: GoalParams, g: GoalRow): void {
  g.title = p.title
  g.goal_type = p.goal_type
  g.goal_mode = p.goal_mode ?? 'daily'
  g.date = p.date ?? null
  g.week_start = p.week_start ?? null
  if (p.checkin_days !== undefined) g.checkin_days = p.checkin_days
  if (p.note !== undefined) g.note = p.note
  if (p.sort !== undefined) g.sort = p.sort
  if (p.status !== undefined) g.status = p.status
  // 完成/未完成状态与 completed_at 保持一致
  if (g.status === 'done') {
    if (!g.completed_at) g.completed_at = today()
  } else {
    g.completed_at = null
  }
}

export async function createGoal(p: GoalParams): Promise<Goal> {
  const params: GoalParams = { ...plain(p), goal_mode: p.goal_mode || 'daily', checkin_days: p.checkin_days ?? 5 }
  await validateGoal(params)
  const now = nowString()
  const g: GoalRow = {
    title: params.title,
    goal_type: params.goal_type,
    goal_mode: params.goal_mode!,
    date: params.date ?? null,
    week_start: params.week_start ?? null,
    checkin_days: params.checkin_days!,
    status: 'active',
    note: params.note ?? '',
    sort: params.sort ?? 0,
    completed_at: null,
    created_at: now,
    updated_at: now,
  }
  applyGoal(params, g)
  const id = await db.goals.add(g)
  return rowToGoal({ ...g, id })
}

export async function updateGoal(id: number, p: GoalParams): Promise<Goal> {
  const g = await getGoal(id)
  const params: GoalParams = {
    ...plain(p),
    goal_mode: p.goal_mode || g.goal_mode,
    checkin_days: p.checkin_days ?? g.checkin_days,
    status: p.status ?? g.status,
  }
  await validateGoal(params)
  const row: GoalRow = { ...g }
  applyGoal(params, row)
  row.updated_at = nowString()
  await db.goals.put(row)
  return rowToGoal(row)
}

export async function deleteGoal(id: number): Promise<void> {
  await getGoal(id) // 不存在 → not_found
  await db.transaction('rw', [db.goals, db.goal_checkins, db.tasks], async () => {
    // 打卡记录级联删除
    await db.goal_checkins.where('goal_id').equals(id).delete()
    // 关联任务解除目标（SQLite ON DELETE SET NULL）
    const linked = await db.tasks.where('goal_id').equals(id).toArray()
    for (const t of linked) {
      await db.tasks.update(t.id!, { goal_id: null, updated_at: nowString() })
    }
    await db.goals.delete(id)
  })
}

// Checkin：日目标按自身日期；周目标（daily 型）日期必须落在绑定周内；once 型不允许打卡。
export async function checkin(goalID: number, date?: string): Promise<GoalCheckin> {
  const g = await getGoal(goalID)
  if (g.status === 'archived') throw errInvalid('已归档目标不能打卡')
  let effective: string
  switch (g.goal_type) {
    case 'day': {
      const d = g.date!
      effective = d
      if (date && date !== effective) throw errInvalid(`打卡日期必须等于目标日期 ${effective}`)
      break
    }
    case 'week': {
      if (g.goal_mode === 'once') throw errInvalid('一次性周目标不支持打卡，请通过状态完成')
      if (!date || !validDate(date)) throw errInvalid('周目标打卡需要 checkin_date')
      const we = weekEndDate(g.week_start!)
      if (date < g.week_start! || date > we) {
        throw errInvalid(`打卡日期必须落在绑定周内（${g.week_start} ~ ${we}）`)
      }
      effective = date
      break
    }
    default:
      throw errInvalid('非法目标类型')
  }
  const exist = await db.goal_checkins.where('[goal_id+checkin_date]').equals([goalID, effective]).first()
  if (exist) throw errConflict('该日期已打卡')
  const ck: GoalCheckinRow = { goal_id: goalID, checkin_date: effective, created_at: nowString() }
  const id = await db.goal_checkins.add(ck)
  return { id: id!, goal_id: goalID, checkin_date: effective, created_at: ck.created_at }
}

// Uncheckin：撤销打卡；不存在则幂等成功。
export async function uncheckin(goalID: number, date?: string): Promise<void> {
  const g = await getGoal(goalID)
  let effective = date
  if (g.goal_type === 'day') {
    effective = g.date!
  } else if (!date || !validDate(date)) {
    throw errInvalid('周目标撤销打卡需要 checkin_date')
  }
  await db.goal_checkins.where('[goal_id+checkin_date]').equals([goalID, effective!]).delete()
}

// ListCheckins 查询区间内的打卡记录（按日期升序）。
export async function listCheckins(from: string, to: string): Promise<GoalCheckin[]> {
  if (!validDate(from) || !validDate(to)) throw errInvalid('from/to 格式非法')
  const rows = await db.goal_checkins
    .where('checkin_date')
    .between(from, to, true, true)
    .toArray()
  rows.sort((a, b) => (a.checkin_date < b.checkin_date ? -1 : a.checkin_date > b.checkin_date ? 1 : a.id! - b.id!))
  return rows.map((r) => ({ id: r.id!, goal_id: r.goal_id, checkin_date: r.checkin_date, created_at: r.created_at }))
}


