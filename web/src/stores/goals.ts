import { defineStore } from 'pinia'
import * as api from '@/api/goals'
import type { Goal, GoalInput } from '@/types/api'
import { addDays } from '@/utils/date'

interface State {
  dayGoals: Goal[]
  dayDate: string
  weekGoals: Goal[]
  weekStart: string
  checked: Record<number, string[]> // goal_id -> checkin 日期列表
  loading: boolean
}

export const useGoalStore = defineStore('goals', {
  state: (): State => ({
    dayGoals: [],
    dayDate: '',
    weekGoals: [],
    weekStart: '',
    checked: {},
    loading: false,
  }),
  getters: {
    isChecked: (s) => (goalId: number, date: string) => (s.checked[goalId] ?? []).includes(date),
    dayRate(s): number | null {
      const active = s.dayGoals.filter((g) => g.status !== 'archived')
      if (active.length === 0) return null
      const done = active.filter((g) => (s.checked[g.id] ?? []).includes(g.date ?? '')).length
      return done / active.length
    },
    weekRate(s): number | null {
      const active = s.weekGoals.filter((g) => g.status !== 'archived')
      if (active.length === 0) return null
      let sum = 0
      for (const g of active) {
        if (g.goal_mode === 'once') {
          sum += g.status === 'done' ? 1 : 0
        } else {
          const n = (s.checked[g.id] ?? []).length
          sum += Math.min(n / (g.checkin_days || 1), 1)
        }
      }
      return sum / active.length
    },
  },
  actions: {
    async loadDay(date: string) {
      this.loading = true
      try {
        this.dayDate = date
        const [goals, checkins] = await Promise.all([
          api.listGoals({ goal_type: 'day', date }),
          api.listCheckins(date, date),
        ])
        this.dayGoals = goals.data
        this.applyCheckins(checkins.data)
      } finally {
        this.loading = false
      }
    },
    async loadWeek(ws: string) {
      this.loading = true
      try {
        this.weekStart = ws
        const [goals, checkins] = await Promise.all([
          api.listGoals({ goal_type: 'week', week_start: ws }),
          api.listCheckins(ws, addDays(ws, 6)),
        ])
        this.weekGoals = goals.data
        this.applyCheckins(checkins.data)
      } finally {
        this.loading = false
      }
    },
    applyCheckins(list: { goal_id: number; checkin_date: string }[]) {
      const m: Record<number, string[]> = {}
      for (const ck of list) {
        ;(m[ck.goal_id] ??= []).push(ck.checkin_date)
      }
      this.checked = m
    },
    async createGoal(input: GoalInput) {
      await api.createGoal(input)
      await this.reload()
    },
    async updateGoal(id: number, input: GoalInput) {
      await api.updateGoal(id, input)
      await this.reload()
    },
    async removeGoal(id: number) {
      await api.deleteGoal(id)
      await this.reload()
    },
    async toggleCheckin(goal: Goal, date: string) {
      const checked = this.isChecked(goal.id, date)
      if (checked) {
        await api.uncheckin(goal.id, date)
      } else {
        await api.checkin(goal.id, date)
      }
      const list = this.checked[goal.id] ?? []
      this.checked[goal.id] = checked ? list.filter((d) => d !== date) : [...list, date]
    },
    async reload() {
      await Promise.all([this.loadDay(this.dayDate), this.loadWeek(this.weekStart)])
    },
  },
})