import { http } from './client'
import type { Goal, GoalCheckin, GoalInput } from '@/types/api'

export function listGoals(params: {
  goal_type?: string
  date?: string
  week_start?: string
  status?: string
} = {}) {
  return http.get<Goal[]>('/goals', { params })
}

export function createGoal(input: GoalInput) {
  return http.post<Goal>('/goals', input)
}

export function updateGoal(id: number, input: GoalInput) {
  return http.put<Goal>(`/goals/${id}`, input)
}

export function deleteGoal(id: number) {
  return http.delete(`/goals/${id}`)
}

export function checkin(id: number, checkinDate?: string) {
  return http.post<GoalCheckin>(`/goals/${id}/checkin`, checkinDate ? { checkin_date: checkinDate } : {})
}

export function uncheckin(id: number, checkinDate?: string) {
  return http.delete(`/goals/${id}/checkin`, { params: { checkin_date: checkinDate } })
}

export function listCheckins(from: string, to: string) {
  return http.get<GoalCheckin[]>('/checkins', { params: { from, to } })
}