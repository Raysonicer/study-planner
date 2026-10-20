import { db, type BlockRow } from '@/db/db'
import { errInvalid, errNotFound } from './errors'
import { plain } from './util'
import { nowString, validDate } from './time'
import type { Block } from '@/types/api'

// ===== 与 Go 版 service/blocks.go 对齐 =====

export interface BlockParams {
  block_date: string
  start_min: number
  end_min: number
  title: string
  task_id?: number | null
  color?: string
  note?: string
}

function rowToBlock(r: BlockRow): Block {
  return { ...r, id: r.id! }
}

function validateBlockParams(p: BlockParams): void {
  if (!p.title) throw errInvalid('title 不能为空')
  if (!validDate(p.block_date)) throw errInvalid('block_date 格式非法')
  if (p.start_min < 0 || p.start_min > 1439) throw errInvalid('start_min 必须在 0-1439')
  if (p.end_min < 1 || p.end_min > 1440) throw errInvalid('end_min 必须在 1-1440')
  if (p.end_min <= p.start_min) throw errInvalid('end_min 必须大于 start_min')
}

export async function listBlocks(from: string, to: string): Promise<Block[]> {
  if (!validDate(from) || !validDate(to)) throw errInvalid('from/to 格式非法')
  const rows = await db.blocks.where('block_date').between(from, to, true, true).toArray()
  rows.sort((a, b) => (a.block_date < b.block_date ? -1 : a.block_date > b.block_date ? 1 : a.start_min - b.start_min || a.id! - b.id!))
  return rows.map(rowToBlock)
}

export async function getBlock(id: number): Promise<Block> {
  const r = await db.blocks.get(id)
  if (!r) throw errNotFound()
  return rowToBlock(r)
}

async function assertTaskExists(taskID: number | null | undefined): Promise<void> {
  if (taskID == null) return
  const t = await db.tasks.get(taskID)
  if (!t) throw errInvalid('关联任务不存在')
}

export async function createBlock(p: BlockParams): Promise<Block> {
  validateBlockParams(p)
  await assertTaskExists(p.task_id)
  const now = nowString()
  const b: BlockRow = {
    block_date: p.block_date,
    start_min: p.start_min,
    end_min: p.end_min,
    title: p.title,
    task_id: p.task_id ?? null,
    color: p.color ?? '',
    note: p.note ?? '',
    created_at: now,
    updated_at: now,
  }
  const id = await db.blocks.add(b)
  return rowToBlock({ ...b, id })
}

export async function updateBlock(id: number, p: BlockParams): Promise<Block> {
  const cur = await getBlock(id)
  validateBlockParams(p)
  await assertTaskExists(p.task_id)
  const row: BlockRow = { ...cur, ...p, task_id: p.task_id ?? null, color: p.color ?? '', note: p.note ?? '', updated_at: nowString() }
  await db.blocks.put(row)
  return rowToBlock(row)
}

export async function deleteBlock(id: number): Promise<void> {
  await getBlock(id)
  await db.blocks.delete(id)
}

