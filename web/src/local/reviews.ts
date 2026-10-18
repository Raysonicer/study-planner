import { db, type ReviewRow } from '@/db/db'
import { errInvalid, errNotFound } from './errors'
import { nowString, validDate, isWeekStart } from './time'
import { getWeekStartDay } from './settings'
import type { Review, Snapshot } from '@/types/api'

// ===== 与 Go 版 service/reviews.go 对齐 =====

export interface ReviewParams {
  title?: string
  content_md?: string
  stats_snapshot?: Snapshot | null
}

function rowToReview(r: ReviewRow): Review {
  return { ...r, id: r.id!, stats_snapshot: (r.stats_snapshot ?? null) as Snapshot | null }
}

export async function listReviews(): Promise<Review[]> {
  const rows = await db.reviews.toArray()
  rows.sort((a, b) => (a.week_start > b.week_start ? -1 : 1))
  return rows.map(rowToReview)
}

export async function getReview(weekStart: string): Promise<Review> {
  const r = await db.reviews.where('week_start').equals(weekStart).first()
  if (!r) throw errNotFound()
  return rowToReview(r)
}

// UpsertReview 按周创建或更新复盘（快照随保存冻结）。
export async function upsertReview(weekStart: string, p: ReviewParams): Promise<Review> {
  if (!validDate(weekStart) || !isWeekStart(weekStart, await getWeekStartDay())) {
    throw errInvalid('week_start 必须是周一日期')
  }
  const now = nowString()
  const existing = await db.reviews.where('week_start').equals(weekStart).first()
  if (existing) {
    const row: ReviewRow = { ...existing }
    if (p.title !== undefined) row.title = p.title
    if (p.content_md !== undefined) row.content_md = p.content_md
    if (p.stats_snapshot !== undefined) row.stats_snapshot = p.stats_snapshot
    row.updated_at = now
    await db.reviews.put(row)
    return rowToReview(row)
  }
  const row: ReviewRow = {
    week_start: weekStart,
    title: p.title ?? '',
    content_md: p.content_md ?? '',
    stats_snapshot: p.stats_snapshot ?? {},
    created_at: now,
    updated_at: now,
  }
  const id = await db.reviews.add(row)
  return rowToReview({ ...row, id })
}

export async function deleteReview(weekStart: string): Promise<void> {
  await getReview(weekStart)
  await db.reviews.where('week_start').equals(weekStart).delete()
}
