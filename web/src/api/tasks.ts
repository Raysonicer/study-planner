import { local } from './client'
import * as svc from '@/local/tasks'
import type { Task, TaskInput } from '@/types/api'

export function listTasks(params: {
  status?: string
  goal_id?: number | null
  parent_id?: number | null
  tag?: string
  due_from?: string
  due_to?: string
  q?: string
} = {}) {
  return local<Task[]>(() => svc.listTasks(params))
}

export function createTask(input: TaskInput) {
  return local<Task>(() => svc.createTask(input))
}

export function updateTask(id: number, input: TaskInput) {
  return local<Task>(() => svc.updateTask(id, input))
}

export function deleteTask(id: number) {
  return local<void>(() => svc.deleteTask(id))
}

export function completeTask(id: number) {
  return local<Task>(() => svc.completeTask(id))
}

export function reopenTask(id: number) {
  return local<Task>(() => svc.reopenTask(id))
}

export function batchTasks(parentId: number | null, goalId: number | null, lines: string[]) {
  return local<Task[]>(() => svc.batchTasks(parentId, goalId, lines))
}

export function reorderTasks(parentId: number | null, orderedIds: number[]) {
  return local<void>(() => svc.reorderTasks(parentId, orderedIds))
}
