<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { ElMessageBox } from 'element-plus'
import { useGoalStore } from '@/stores/goals'
import { useSettingsStore } from '@/stores/settings'
import { useTaskStore } from '@/stores/tasks'
import GoalEditor from '@/components/GoalEditor.vue'
import { withUndo } from '@/utils/undo'
import { today, weekStartOf, addDays, weekDays, weekLabels } from '@/utils/date'
import type { Goal } from '@/types/api'

const gstore = useGoalStore()
const settings = useSettingsStore()
const tstore = useTaskStore()

const dayDate = ref(today())
const weekStart = ref(weekStartOf(dayDate.value))

const editorOpen = ref(false)
const editing = ref<Goal | null>(null)
const editorType = ref<'day' | 'week'>('day')

async function load() {
  await Promise.all([gstore.loadDay(dayDate.value), gstore.loadWeek(weekStart.value), tstore.fetchAll()])
}
onMounted(load)
watch([dayDate, weekStart], load)

const dayRatePct = computed(() => (gstore.dayRate === null ? null : Math.round(gstore.dayRate * 100)))
const weekRatePct = computed(() => (gstore.weekRate === null ? null : Math.round(gstore.weekRate * 100)))

// 关联任务参考口径（G-11）
const taskRef = computed(() => {
  const m = new Map<number, { done: number; total: number }>()
  for (const t of tstore.tasks) {
    if (t.goal_id == null) continue
    const it = m.get(t.goal_id) ?? { done: 0, total: 0 }
    it.total++
    if (t.status === 'done') it.done++
    m.set(t.goal_id, it)
  }
  return m
})

function isMissed(g: Goal): boolean {
  return g.goal_type === 'day' && !!g.date && g.date < today() && !gstore.isChecked(g.id, g.date)
}

async function toggleDay(goal: Goal) {
  const wasChecked = gstore.isChecked(goal.id, goal.date!)
  await gstore.toggleCheckin(goal, goal.date!)
  if (!wasChecked) {
    withUndo(`已达成「${goal.title}」`, '撤销', () => gstore.toggleCheckin(goal, goal.date!))
  }
}

async function toggleWeekDay(goal: Goal, date: string) {
  const wasChecked = gstore.isChecked(goal.id, date)
  await gstore.toggleCheckin(goal, date)
  if (!wasChecked) {
    withUndo(`「${goal.title}」${date} 已打卡`, '撤销', () => gstore.toggleCheckin(goal, date))
  }
}

async function toggleOnce(goal: Goal) {
  const status = goal.status === 'done' ? 'active' : 'done'
  await gstore.updateGoal(goal.id, { title: goal.title, goal_type: 'week', goal_mode: 'once', week_start: goal.week_start, status })
  withUndo(status === 'done' ? `已完成「${goal.title}」` : `已重开「${goal.title}」`, '撤销', () =>
    gstore.updateGoal(goal.id, { title: goal.title, goal_type: 'week', goal_mode: 'once', week_start: goal.week_start, status: goal.status === 'done' ? 'active' : 'done' }),
  )
}

async function archiveGoal(goal: Goal) {
  await ElMessageBox.confirm('归档后默认视图隐藏，历史统计保留', `归档「${goal.title}」`, { type: 'warning' })
  await gstore.updateGoal(goal.id, { title: goal.title, goal_type: goal.goal_type, goal_mode: goal.goal_mode, date: goal.date, week_start: goal.week_start, checkin_days: goal.checkin_days, status: 'archived' })
}

async function removeGoal(goal: Goal) {
  await ElMessageBox.confirm('打卡记录将一并删除', `删除「${goal.title}」`, { type: 'warning' })
  await gstore.removeGoal(goal.id)
}

function openCreate(type: 'day' | 'week') {
  editing.value = null
  editorType.value = type
  editorOpen.value = true
}
function openEdit(goal: Goal) {
  editing.value = goal
  editorType.value = goal.goal_type
  editorOpen.value = true
}
</script>

<template>
  <div class="page">
    <div class="nav-bar">
      <div class="nav-group">
        <el-button-group>
          <el-button @click="dayDate = addDays(dayDate, -1)"><el-icon><ArrowLeft /></el-icon></el-button>
          <el-button @click="dayDate = today()">今天</el-button>
          <el-button @click="dayDate = addDays(dayDate, 1)"><el-icon><ArrowRight /></el-icon></el-button>
        </el-button-group>
        <el-date-picker v-model="dayDate" type="date" value-format="YYYY-MM-DD" style="width: 150px" />
      </div>
      <div class="nav-group">
        <el-button-group>
          <el-button @click="weekStart = addDays(weekStart, -7)"><el-icon><ArrowLeft /></el-icon></el-button>
          <el-button @click="weekStart = weekStartOf(today())">本周</el-button>
          <el-button @click="weekStart = addDays(weekStart, 7)"><el-icon><ArrowRight /></el-icon></el-button>
        </el-button-group>
        <el-date-picker v-model="weekStart" type="week" value-format="YYYY-MM-DD" style="width: 150px" />
      </div>
    </div>

    <el-row :gutter="16">
      <el-col :xs="24" :md="12">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>日目标 · {{ dayDate }}</span>
              <el-progress v-if="dayRatePct !== null" type="circle" :percentage="dayRatePct" :width="40" />
              <el-button type="primary" size="small" @click="openCreate('day')">+ 日目标</el-button>
            </div>
          </template>
          <el-empty v-if="!gstore.dayGoals.length" description="暂无日目标" :image-size="60" />
          <div v-for="g in gstore.dayGoals" :key="g.id" class="goal-row">
            <el-checkbox
              :model-value="gstore.isChecked(g.id, g.date!)"
              :disabled="!!g.date && g.date > today()"
              @change="toggleDay(g)"
            />
            <span class="goal-title" :class="{ done: gstore.isChecked(g.id, g.date!) }">{{ g.title }}</span>
            <el-tag v-if="isMissed(g)" size="small" type="danger">未达成</el-tag>
            <el-tag v-if="taskRef.get(g.id)?.total" size="small" type="info" effect="plain">
              任务 {{ taskRef.get(g.id)!.done }}/{{ taskRef.get(g.id)!.total }}
            </el-tag>
            <span class="row-actions">
              <el-button link size="small" @click="openEdit(g)">编辑</el-button>
              <el-button link size="small" type="danger" @click="removeGoal(g)">删除</el-button>
            </span>
          </div>
        </el-card>
      </el-col>

      <el-col :xs="24" :md="12">
        <el-card>
          <template #header>
            <div class="card-header">
              <span>周目标 · {{ weekStart }} 起</span>
              <el-progress v-if="weekRatePct !== null" type="circle" :percentage="weekRatePct" :width="40" />
              <el-button type="primary" size="small" @click="openCreate('week')">+ 周目标</el-button>
            </div>
          </template>
          <el-empty v-if="!gstore.weekGoals.length" description="暂无周目标" :image-size="60" />
          <div v-for="g in gstore.weekGoals" :key="g.id" class="goal-row week">
            <div class="week-goal-main">
              <span class="goal-title" :class="{ done: g.status === 'done' }">{{ g.title }}</span>
              <el-tag size="small" :type="g.goal_mode === 'once' ? 'warning' : 'success'">
                {{ g.goal_mode === 'once' ? '一次性' : `打卡 ${gstore.checked[g.id]?.length ?? 0}/${g.checkin_days}` }}
              </el-tag>
              <el-tag v-if="taskRef.get(g.id)?.total" size="small" type="info" effect="plain">
                任务 {{ taskRef.get(g.id)!.done }}/{{ taskRef.get(g.id)!.total }}
              </el-tag>
              <span class="row-actions">
                <el-button link size="small" @click="openEdit(g)">编辑</el-button>
                <el-button link size="small" @click="archiveGoal(g)">归档</el-button>
                <el-button link size="small" type="danger" @click="removeGoal(g)">删除</el-button>
              </span>
            </div>
            <div v-if="g.goal_mode === 'daily'" class="week-checkins">
              <el-check-tag
                v-for="(d, i) in weekDays(g.week_start!)"
                :key="d"
                :checked="gstore.isChecked(g.id, d)"
                @change="toggleWeekDay(g, d)"
              >
                {{ weekLabels(settings.weekStartDay)[i] }}
              </el-check-tag>
            </div>
            <el-switch v-else :model-value="g.status === 'done'" @change="toggleOnce(g)" />
          </div>
        </el-card>
      </el-col>
    </el-row>

    <GoalEditor
      v-model="editorOpen"
      :goal="editing"
      :goal-type="editorType"
      :date="dayDate"
      :week-start="weekStart"
    />
  </div>
</template>

<style scoped>
.nav-bar {
  display: flex;
  gap: 24px;
  margin-bottom: 16px;
  align-items: center;
}
.nav-group {
  display: flex;
  gap: 8px;
  align-items: center;
}
.card-header {
  display: flex;
  align-items: center;
  gap: 12px;
}
.card-header > span:first-child {
  flex: 1;
  font-weight: 600;
}
.goal-row {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 4px;
}
.goal-row.week {
  flex-direction: column;
  align-items: stretch;
}
.week-goal-main {
  display: flex;
  align-items: center;
  gap: 8px;
}
.goal-title {
  flex: 1;
}
.goal-title.done {
  text-decoration: line-through;
  color: var(--el-text-color-secondary);
}
.row-actions {
  display: none;
  gap: 2px;
}
.goal-row:hover .row-actions {
  display: inline-flex;
}
.week-checkins {
  display: flex;
  gap: 6px;
  margin-top: 6px;
  padding-left: 4px;
}
</style>