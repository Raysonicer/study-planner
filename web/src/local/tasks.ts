import { db, type TaskRow, type TaskEventRow } from '@/db/db'
import { errInvalid, errNotFound } from './errors'
import { today, nowString, validDate } from './time'
import type { Task, TaskInput } from '@/types/api'

// ===== 与 Go 版 service/tasks.go 对齐 =====

export interface TaskFilter {
  status?: string
  goal_id?: number | null
  parent_id?: number | null
  tag?: string
  due_from?: string
  due_to?: string
  q?: string
}

function rowToTask(r: TaskRow): Task {
  return { ...r, id: r.id!, tags: r.tags ?? [] }
}

function validateTaskParams(p: TaskInput): void {
  if (!p.title) throw errInvalid('title 不能为空')
  if (p.status !== undefined && !['todo', 'doing', 'done'].includes(p.status)) throw errInvalid('非法 status')
  if (p.priority !== undefined && (p.priority < 0 || p.priority > 3)) throw errInvalid('priority 必须在 0-3')
  if (p.due_date && !validDate(p.due_date)) throw errInvalid('due_date 格式非法')
  if (p.est_minutes !== undefined && p.est_minutes !== null && p.est_minutes <= 0) {
    throw errInvalid('est_minutes 必须为正数')
  }
}

export async function listTasks(f: TaskFilter = {}): Promise<Task[]> {
  let coll = db.tasks.toCollection()
  if (f.status) coll = coll.filter((t) => t.status === f.status)
  if (f.goal_id != null) coll = coll.filter((t) => t.goal_id === f.goal_id)
  if (f.parent_id != null) coll = coll.filter((t) => t.parent_id === f.parent_id)
  if (f.tag) coll = coll.filter((t) => (t.tags ?? []).includes(f.tag!))
  if (f.due_from) coll = coll.filter((t) => !!t.due_date && t.due_date >= f.due_from!)
  if (f.due_to) coll = coll.filter((t) => !!t.due_date && t.due_date <= f.due_to!)
  if (f.q) coll = coll.filter((t) => t.title.includes(f.q!))
  const rows = await coll.toArray()
  rows.sort((a, b) => a.sort - b.sort || a.id! - b.id!)
  return rows.map(rowToTask)
}

export async function getTask(id: number): Promise<Task> {
  const r = await db.tasks.get(id)
  if (!r) throw errNotFound()
  return rowToTask(r)
}

// wouldCreateCycle 判断把 taskID 挂到 parentID 下是否会形成环。
async function wouldCreateCycle(taskID: number, parentID: number): Promise<boolean> {
  let cur: number | null = parentID
  for (let depth = 0; depth < 10000; depth++) {
    if (cur === taskID) return true
    if (cur === null) return false
    const t: TaskRow | undefined = await db.tasks.get(cur)
    if (!t) return false // 父链到头或不存在
    cur = t.parent_id
  }
  throw errInvalid('任务树深度超限')
}

async function writeEvent(taskID: number, typ: 'created' | 'done' | 'reopen'): Promise<void> {
  const ev: TaskEventRow = { task_id: taskID, event_type: typ, occurred_date: today(), occurred_at: nowString() }
  await db.task_events.add(ev)
}

export async function createTask(p: TaskInput): Promise<Task> {
  const params: TaskInput = { ...p, status: p.status ?? 'todo', priority: p.priority ?? 1 }
  validateTaskParams(params)
  if (params.parent_id != null) {
    await getTask(params.parent_id) // 父任务不存在 → not_found（Go 返回 invalid，这里与前端提示兼容）
  }
  if (params.goal_id != null) {
    const g = await db.goals.get(params.goal_id)
    if (!g) throw errInvalid('目标不存在')
  }
  const now = nowString()
  const t: TaskRow = {
    title: params.title,
    note: params.note ?? '',
    parent_id: params.parent_id ?? null,
    goal_id: params.goal_id ?? null,
    status: params.status!,
    priority: params.priority!,
    due_date: params.due_date ?? null,
    est_minutes: params.est_minutes ?? null,
    tags: params.tags ?? [],
    sort: params.sort ?? 0,
    completed_at: null,
    created_at: now,
    updated_at: now,
  }
  const id = await db.tasks.add(t)
  await writeEvent(id, 'created')
  return rowToTask({ ...t, id })
}

export async function updateTask(id: number, p: TaskInput): Promise<Task> {
  const cur = await getTask(id)
  const params: TaskInput = {
    ...p,
    status: p.status ?? cur.status,
    priority: p.priority ?? cur.priority,
  }
  validateTaskParams(params)
  // Go 语义：parent_id 缺失/为空 → 置为 null；显式给定则校验存在性与环
  if (params.parent_id != null) {
    if (params.parent_id === id) throw errInvalid('任务不能作为自己的子任务')
    await getTask(params.parent_id)
    if (await wouldCreateCycle(id, params.parent_id)) throw errInvalid('父子关系会形成环')
  }
  if (params.goal_id != null) {
    const g = await db.goals.get(params.goal_id)
    if (!g) throw errInvalid('目标不存在')
  }
  const row: TaskRow = { ...cur }
  row.title = params.title
  if (params.note !== undefined) row.note = params.note
  row.parent_id = params.parent_id ?? null
  row.goal_id = params.goal_id ?? null
  row.status = params.status!
  row.priority = params.priority!
  row.due_date = params.due_date ?? null
  row.est_minutes = params.est_minutes ?? null
  if (params.tags !== undefined) row.tags = params.tags
  if (params.sort !== undefined) row.sort = params.sort
  row.updated_at = nowString()
  await db.tasks.put(row)
  return rowToTask(row)
}

// 删除任务：子任务与事件级联删除；日程块解除关联（ON DELETE SET NULL）。
export async function deleteTask(id: number): Promise<void> {
  await getTask(id)
  await db.transaction('rw', [db.tasks, db.task_events, db.blocks], async () => {
    const ids = [id]
    // 收集全部后代
    const all = await db.tasks.toArray()
    const childrenOf = new Map<number, number[]>()
    for (const t of all) {
      if (t.parent_id != null) {
        if (!childrenOf.has(t.parent_id)) childrenOf.set(t.parent_id, [])
        childrenOf.get(t.parent_id)!.push(t.id!)
      }
    }
    const queue = [id]
    while (queue.length) {
      const cur = queue.shift()!
      for (const c of childrenOf.get(cur) ?? []) {
        if (!ids.includes(c)) {
          ids.push(c)
          queue.push(c)
        }
      }
    }
    await db.task_events.where('task_id').anyOf(ids).delete()
    const linked = await db.blocks.where('task_id').anyOf(ids).toArray()
    for (const b of linked) {
      await db.blocks.update(b.id!, { task_id: null, updated_at: nowString() })
    }
    await db.tasks.bulkDelete(ids)
  })
}

// CompleteTask 幂等：已完成则直接返回，不重复写事件。
export async function completeTask(id: number): Promise<Task> {
  const t = await getTask(id)
  if (t.status === 'done') return t
  const row: TaskRow = { ...t, status: 'done', completed_at: today(), updated_at: nowString() }
  await db.tasks.put(row)
  await writeEvent(id, 'done')
  return rowToTask(row)
}

// ReopenTask 幂等：未完成则直接返回。
export async function reopenTask(id: number): Promise<Task> {
  const t = await getTask(id)
  if (t.status !== 'done') return t
  const row: TaskRow = { ...t, status: 'todo', completed_at: null, updated_at: nowString() }
  await db.tasks.put(row)
  await writeEvent(id, 'reopen')
  return rowToTask(row)
}

// BatchTasks 批量创建子任务：每行一个，自动去空行（T-05）。
export async function batchTasks(
  parentID: number | null,
  goalID: number | null,
  lines: string[],
): Promise<Task[]> {
  const titles = lines.map((l) => l.trim()).filter((t) => t !== '')
  if (titles.length === 0) throw errInvalid('lines 至少包含一行非空文本')
  if (parentID != null) await getTask(parentID)
  if (goalID != null) {
    const g = await db.goals.get(goalID)
    if (!g) throw errInvalid('目标不存在')
  }
  // 同级最大 sort
  const all = await db.tasks.toArray()
  let maxSort = -1
  for (const t of all) {
    if (t.parent_id === parentID && t.sort > maxSort) maxSort = t.sort
  }
  const created: Task[] = []
  for (let i = 0; i < titles.length; i++) {
    const now = nowString()
    const t: TaskRow = {
      title: titles[i],
      note: '',
      parent_id: parentID,
      goal_id: goalID,
      status: 'todo',
      priority: 1,
      due_date: null,
      est_minutes: null,
      tags: [],
      sort: maxSort + 1 + i,
      completed_at: null,
      created_at: now,
      updated_at: now,
    }
    const id = await db.tasks.add(t)
    await writeEvent(id, 'created')
    created.push(rowToTask({ ...t, id }))
  }
  return created
}

// ReorderTasks 同层级排序：按传入顺序写 sort（T-10）。
export async function reorderTasks(_parentID: number | null, orderedIDs: number[]): Promise<void> {
  await db.transaction('rw', db.tasks, async () => {
    for (let i = 0; i < orderedIDs.length; i++) {
      await db.tasks.update(orderedIDs[i], { sort: i, updated_at: nowString() })
    }
  })
}

