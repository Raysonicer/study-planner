import { local } from './client'
import * as svc from '@/local/blocks'
import type { Block } from '@/types/api'

export interface BlockInput {
  block_date: string
  start_min: number
  end_min: number
  title: string
  task_id?: number | null
  color?: string
  note?: string
}

export function listBlocks(from: string, to: string) {
  return local<Block[]>(() => svc.listBlocks(from, to))
}

export function createBlock(input: BlockInput) {
  return local<Block>(() => svc.createBlock(input))
}

export function updateBlock(id: number, input: BlockInput) {
  return local<Block>(() => svc.updateBlock(id, input))
}

export function deleteBlock(id: number) {
  return local<void>(() => svc.deleteBlock(id))
}
