import Dexie, { type Table } from 'dexie'

// 与 Go 版 SQLite schema（000001_init.up.sql）一一对应的浏览器数据库
// 类型复用 types/api.ts 的 DTO 语义，保证字段一致。

export interface SettingRow {
  key: string
  value: string
}

export interface GoalRow {
  id?: number
  title: string
  goal_type: 'day' | 'week'
  goal_mode: 'daily' | 'once'
  date: string | null
  week_start: string | null
  checkin_days: number
  status: 'active' | 'done' | 'archived'
  note: string
  sort: number
  completed_at: string | null
  created_at: string
  updated_at: string
}

export interface GoalCheckinRow {
  id?: number
  goal_id: number
  checkin_date: string
  created_at: string
}

export interface TaskRow {
  id?: number
  title: string
  note: string
  parent_id: number | null
  goal_id: number | null
  status: 'todo' | 'doing' | 'done'
  priority: number
  due_date: string | null
  est_minutes: number | null
  tags: string[]
  sort: number
  completed_at: string | null
  created_at: string
  updated_at: string
}

export interface TaskEventRow {
  id?: number
  task_id: number
  event_type: 'created' | 'done' | 'reopen'
  occurred_date: string
  occurred_at: string
}

export interface BlockRow {
  id?: number
  block_date: string
  start_min: number
  end_min: number
  title: string
  task_id: number | null
  color: string
  note: string
  created_at: string
  updated_at: string
}

export interface ReviewRow {
  id?: number
  week_start: string
  title: string
  content_md: string
  stats_snapshot: unknown
  created_at: string
  updated_at: string
}

class StudyPilotDB extends Dexie {
  settings!: Table<SettingRow, string>
  goals!: Table<GoalRow, number>
  goal_checkins!: Table<GoalCheckinRow, number>
  tasks!: Table<TaskRow, number>
  task_events!: Table<TaskEventRow, number>
  blocks!: Table<BlockRow, number>
  reviews!: Table<ReviewRow, number>

  constructor() {
    super('studypilot')
    this.version(1).stores({
      settings: 'key',
      goals: '++id, goal_type, date, week_start, status, sort',
      goal_checkins: '++id, goal_id, checkin_date, [goal_id+checkin_date]',
      tasks: '++id, parent_id, goal_id, status, due_date, sort',
      task_events: '++id, task_id, event_type, occurred_date',
      blocks: '++id, block_date, task_id',
      reviews: '++id, week_start',
    })
  }
}

export const db = new StudyPilotDB()
