import { http } from './client'
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
  return http.get<Task[]>('/tasks', { params })
}

export function createTask(input: TaskInput) {
  return http.post<Task>('/tasks', input)
}

export function updateTask(id: number, input: TaskInput) {
  return http.put<Task>(`/tasks/${id}`, input)
}

export function deleteTask(id: number) {
  return http.delete(`/tasks/${id}`)
}

export function completeTask(id: number) {
  return http.post<Task>(`/tasks/${id}/complete`)
}

export function reopenTask(id: number) {
  return http.post<Task>(`/tasks/${id}/reopen`)
}

export function batchTasks(parentId: number | null, goalId: number | null, lines: string[]) {
  return http.post<Task[]>('/tasks/batch', { parent_id: parentId, goal_id: goalId, lines })
}

export function reorderTasks(parentId: number | null, orderedIds: number[]) {
  return http.patch('/tasks/reorder', { parent_id: parentId, ordered_ids: orderedIds })
}