import { local } from './client'
import * as svc from '@/local/goals'
import type { Goal, GoalCheckin, GoalInput } from '@/types/api'

export function listGoals(params: {
  goal_type?: string
  date?: string
  week_start?: string
  status?: string
} = {}) {
  return local<Goal[]>(() => svc.listGoals(params))
}

export function createGoal(input: GoalInput) {
  return local<Goal>(() => svc.createGoal(input))
}

export function updateGoal(id: number, input: GoalInput) {
  return local<Goal>(() => svc.updateGoal(id, input))
}

export function deleteGoal(id: number) {
  return local<void>(() => svc.deleteGoal(id))
}

export function checkin(id: number, checkinDate?: string) {
  return local<GoalCheckin>(() => svc.checkin(id, checkinDate))
}

export function uncheckin(id: number, checkinDate?: string) {
  return local<void>(() => svc.uncheckin(id, checkinDate))
}

export function listCheckins(from: string, to: string) {
  return local<GoalCheckin[]>(() => svc.listCheckins(from, to))
}
