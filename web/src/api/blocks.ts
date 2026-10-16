import { http } from './client'
import type { Block } from '@/types/api'

export function listBlocks(from: string, to: string) {
  return http.get<Block[]>('/blocks', { params: { from, to } })
}

export function createBlock(input: {
  block_date: string
  start_min: number
  end_min: number
  title: string
  task_id?: number | null
  color?: string
  note?: string
}) {
  return http.post<Block>('/blocks', input)
}

export function updateBlock(
  id: number,
  input: {
    block_date: string
    start_min: number
    end_min: number
    title: string
    task_id?: number | null
    color?: string
    note?: string
  },
) {
  return http.put<Block>(`/blocks/${id}`, input)
}

export function deleteBlock(id: number) {
  return http.delete(`/blocks/${id}`)
}