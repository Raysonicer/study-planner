import { defineStore } from 'pinia'
import { ElMessage, ElMessageBox } from 'element-plus'
import * as api from '@/api/tasks'
import { useSettingsStore } from './settings'
import type { Task, TaskInput } from '@/types/api'
import { today } from '@/utils/date'
import { withUndo } from '@/utils/undo'

interface State {
  tasks: Task[]
  loading: boolean
}

export const useTaskStore = defineStore('tasks', {
  state: (): State => ({
    tasks: [],
    loading: false,
  }),
  getters: {
    byId: (s) => {
      const m = new Map<number, Task>()
      for (const t of s.tasks) m.set(t.id, t)
      return m
    },
    roots: (s) => s.tasks.filter((t) => !t.parent_id).sort(bySort),
    childrenMap: (s) => {
      const m = new Map<number | null, Task[]>()
      for (const t of s.tasks) {
        const key = t.parent_id
        if (!m.has(key)) m.set(key, [])
        m.get(key)!.push(t)
      }
      for (const list of m.values()) list.sort(bySort)
      return m
    },
    progress: (s) => {
      return (id: number) => {
        const children = s.tasks.filter((t) => t.parent_id === id)
        const done = children.filter((t) => t.status === 'done').length
        return { done, total: children.length }
      }
    },
    isOverdue: () => (t: Task) => !!t.due_date && t.due_date < today() && t.status !== 'done',
  },
  actions: {
    async fetchAll() {
      this.loading = true
      try {
        const res = await api.listTasks()
        this.tasks = res.data
      } finally {
        this.loading = false
      }
    },
    async create(input: TaskInput) {
      const res = await api.createTask(input)
      await this.fetchAll()
      return res.data
    },
    async update(id: number, input: TaskInput) {
      await api.updateTask(id, input)
      await this.fetchAll()
    },
    async remove(id: number) {
      const t = this.byId.get(id)
      const hasChildren = this.tasks.some((x) => x.parent_id === id)
      const tip = hasChildren ? '该任务含子任务，将一并删除' : '删除后不可恢复'
      await ElMessageBox.confirm(tip, `删除「${t?.title}」`, { type: 'warning' })
      await api.deleteTask(id)
      await this.fetchAll()
    },
    async toggleDone(task: Task) {
      const wasDone = task.status === 'done'
      if (task.status === 'done') {
        await api.reopenTask(task.id)
        ElMessage.success('已重开任务')
      } else {
        await api.completeTask(task.id)
        withUndo(`已完成「${task.title}」`, '撤销', async () => {
          await api.reopenTask(task.id)
          await this.fetchAll()
        })
      }
      await this.fetchAll()
      // 自动完成父任务（设置项 auto_complete_parent）
      if (!wasDone && task.parent_id != null && useSettingsStore().autoCompleteParent) {
        const parent = this.byId.get(task.parent_id)
        if (parent && parent.status !== 'done') {
          const siblings = this.tasks.filter((x) => x.parent_id === task.parent_id)
          if (siblings.length > 0 && siblings.every((x) => x.status === 'done')) {
            await api.completeTask(parent.id)
            await this.fetchAll()
            ElMessage.success(`父任务「${parent.title}」已自动完成`)
          }
        }
      }
    },
    async batch(parentId: number | null, goalId: number | null, lines: string[]) {
      const res = await api.batchTasks(parentId, goalId, lines)
      await this.fetchAll()
      ElMessage.success(`已创建 ${res.data.length} 个子任务`)
    },
    async reorder(parentId: number | null, orderedIds: number[]) {
      await api.reorderTasks(parentId, orderedIds)
      await this.fetchAll()
    },
  },
})

export function bySort(a: Task, b: Task) {
  return a.sort - b.sort || a.id - b.id
}