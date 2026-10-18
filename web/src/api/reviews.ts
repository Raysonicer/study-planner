import { local } from './client'
import * as svc from '@/local/reviews'
import * as statsSvc from '@/local/stats'
import type { Review, Snapshot } from '@/types/api'

export function listReviews() {
  return local<Review[]>(() => svc.listReviews())
}

export function getReview(weekStart: string) {
  // 404 表示该周还没有复盘，属预期流程，不弹错误提示
  return local<Review>(() => svc.getReview(weekStart), { skipErrorToast: true })
}

export function putReview(weekStart: string, input: { title?: string; content_md?: string; stats_snapshot?: Snapshot | null }) {
  return local<Review>(() => svc.upsertReview(weekStart, input))
}

export function deleteReview(weekStart: string) {
  return local<void>(() => svc.deleteReview(weekStart))
}

export function generateSnapshot(weekStart: string) {
  return local<Snapshot>(() => statsSvc.computeSnapshot(weekStart))
}
