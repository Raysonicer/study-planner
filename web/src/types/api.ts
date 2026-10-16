// 与 docs/openapi.yaml 对齐的 DTO

export type GoalType = 'day' | 'week'
export type GoalMode = 'daily' | 'once'
export type GoalStatus = 'active' | 'done' | 'archived'
export type TaskStatus = 'todo' | 'doing' | 'done'

export interface Goal {
  id: number
  title: string
  goal_type: GoalType
  goal_mode: GoalMode
  date: string | null
  week_start: string | null
  checkin_days: number
  status: GoalStatus
  note: string
  sort: number
  completed_at: string | null
  created_at: string
  updated_at: string
}

export interface GoalCheckin {
  id: number
  goal_id: number
  checkin_date: string
  created_at: string
}

export interface GoalInput {
  title: string
  goal_type: GoalType
  goal_mode?: GoalMode
  date?: string | null
  week_start?: string | null
  checkin_days?: number
  status?: GoalStatus
  note?: string
  sort?: number
}

export interface Task {
  id: number
  title: string
  note: string
  parent_id: number | null
  goal_id: number | null
  status: TaskStatus
  priority: number
  due_date: string | null
  est_minutes: number | null
  tags: string[]
  sort: number
  completed_at: string | null
  created_at: string
  updated_at: string
}

export interface Block {
  id: number
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

export interface BlockView extends Block {
  done: boolean
}

export interface TrendPoint {
  week_start: string
  task_rate: number | null
  goal_rate: number | null
}

export interface DistItem {
  label: string
  done: number
  total: number
}

export interface StatsDaily {
  date: string
  task_completion_rate: number | null
  goal_achievement_rate: number | null
  scheduled_minutes: number
  completed_block_minutes: number
  overdue_count: number
  task_total: number
  task_done: number
  goal_total: number
  goal_done: number
}

export interface StatsWeekly {
  week_start: string
  week_end: string
  task_completion_rate: number | null
  goal_achievement_rate: number | null
  task_total: number
  task_done: number
  overdue_count: number
  goal_total: number
  goal_done: number
  scheduled_minutes: number
  completed_block_minutes: number
  trend: TrendPoint[]
  by_goal: DistItem[]
  by_tag: DistItem[]
  unfinished: Task[]
}

export interface HeatmapCell {
  date: string
  score: number | null
  task_rate: number | null
  goal_rate: number | null
  task_total: number
  task_done: number
  goal_total: number
  goal_done: number
}

export interface Heatmap {
  start: string
  end: string
  cells: HeatmapCell[]
}

export interface Dashboard {
  date: string
  task_completion_rate: number | null
  goal_achievement_rate: number | null
  scheduled_minutes: number
  overdue_count: number
  day_goals: Goal[]
  tasks_due: Task[]
  blocks: BlockView[]
}

export interface Snapshot {
  generated_at: string
  week_start: string
  week_end: string
  task_completion_rate: number | null
  goal_achievement_rate: number | null
  task_total: number
  task_done: number
  overdue_count: number
  goal_total: number
  goal_done: number
  checkin_total: number
  checkin_days_total: number
  scheduled_minutes: number
  completed_block_minutes: number
  unfinished_tasks: string[]
  by_tag: DistItem[]
}

export interface Review {
  id: number
  week_start: string
  title: string
  content_md: string
  stats_snapshot: Snapshot | null
  created_at: string
  updated_at: string
}

export interface TaskInput {
  title: string
  note?: string
  parent_id?: number | null
  goal_id?: number | null
  status?: TaskStatus
  priority?: number
  due_date?: string | null
  est_minutes?: number | null
  tags?: string[]
  sort?: number
}