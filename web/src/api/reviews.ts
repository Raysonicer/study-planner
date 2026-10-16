import { http } from './client'
import type { Review, Snapshot } from '@/types/api'

export function listReviews() {
  return http.get<Review[]>('/reviews')
}

export function getReview(weekStart: string) {
  return http.get<Review>(`/reviews/${weekStart}`)
}

export function putReview(weekStart: string, input: { title?: string; content_md?: string; stats_snapshot?: Snapshot | null }) {
  return http.put<Review>(`/reviews/${weekStart}`, input)
}

export function deleteReview(weekStart: string) {
  return http.delete(`/reviews/${weekStart}`)
}

export function generateSnapshot(weekStart: string) {
  return http.post<Snapshot>(`/reviews/${weekStart}/snapshot`)
}