<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import * as api from '@/api/stats'
import * as goalsApi from '@/api/goals'
import { useTaskStore } from '@/stores/tasks'
import { useSettingsStore } from '@/stores/settings'
import { today, addDays, formatDate } from '@/utils/date'
import { withUndo } from '@/utils/undo'
import type { Dashboard } from '@/types/api'

const router = useRouter()
const store = useTaskStore()
const date = ref(today())
const dash = ref<Dashboard | null>(null)
const checkedGoalIds = ref<Set<number>>(new Set())
const loading = ref(false)

async function load() {
  loading.value = true
  try {
    const [d, cks] = await Promise.all([api.getDashboard(date.value), goalsApi.listCheckins(date.value, date.value)])
    dash.value = d.data
    checkedGoalIds.value = new Set(cks.data.map((c) => c.goal_id))
    maybeNotify()
  } finally {
    loading.value = false
  }
}

// 逾期桌面通知（设置项 notify_overdue）
async function maybeNotify() {
  if (date.value !== today()) return
  if (!useSettingsStore().notifyEnabled) return
  const count = dash.value?.overdue_count ?? 0
  if (count <= 0 || !('Notification' in window)) return
  if (Notification.permission === 'default') await Notification.requestPermission()
  if (Notification.permission === 'granted') {
    new Notification('StudyPilot · 逾期提醒', { body: `你有 ${count} 个逾期任务，尽快处理` })
  }
}
onMounted(load)
watch(date, load)

const taskRate = computed(() => dash.value?.task_completion_rate ?? null)
const goalRate = computed(() => dash.value?.goal_achievement_rate ?? null)
const ratePct = (r: number | null) => (r === null ? null : Math.round(r * 100))
const minutesFmt = (m: number) => (m >= 60 ? `${Math.floor(m / 60)}h${m % 60 ? ' ' + (m % 60) + 'm' : ''}` : `${m}m`)
const isMissed = (goalDate: string) => goalDate < today()

async function toggleGoal(goalId: number, goalDate: string, title: string, checked: boolean) {
  if (checked) {
    await goalsApi.uncheckin(goalId, goalDate)
  } else {
    await goalsApi.checkin(goalId)
    withUndo(`已达成「${title}」`, '撤销', () => {
      void goalsApi.uncheckin(goalId, goalDate)
    })
  }
  await load()
}

async function toggleTask(id: number) {
  const t = store.byId.get(id) ?? dash.value?.tasks_due.find((x) => x.id === id)
  if (t) await store.toggleDone(t)
  await load()
}
</script>

<template>
  <div class="page">
    <div class="nav-bar">
      <el-button-group>
        <el-button @click="date = addDays(date, -1)"><el-icon><ArrowLeft /></el-icon></el-button>
        <el-button @click="date = today()">今天</el-button>
        <el-button @click="date = addDays(date, 1)"><el-icon><ArrowRight /></el-icon></el-button>
      </el-button-group>
      <el-date-picker v-model="date" type="date" value-format="YYYY-MM-DD" style="width: 150px" />
      <span class="date-title">{{ date === today() ? '今日面板' : formatDate(date) + ' 面板' }}</span>
    </div>

    <el-row :gutter="16" v-loading="loading" class="stat-row">
      <el-col :xs="24" :sm="12" :md="6">
        <el-card class="stat-card">
          <div class="stat-label">任务完成率</div>
          <el-progress type="circle" :percentage="ratePct(taskRate) ?? 0" :width="88" :format="() => (ratePct(taskRate) === null ? '—' : `${ratePct(taskRate)}%`)" />
          <div class="stat-sub">{{ dash?.tasks_due.length ?? 0 }} 个任务今日到期</div>
        </el-card>
      </el-col>
      <el-col :xs="24" :sm="12" :md="6">
        <el-card class="stat-card">
          <div class="stat-label">目标达成率</div>
          <el-progress type="circle" :percentage="ratePct(goalRate) ?? 0" :width="88" color="#2f9e44" :format="() => (ratePct(goalRate) === null ? '—' : `${ratePct(goalRate)}%`)" />
          <div class="stat-sub">{{ dash?.day_goals.length ?? 0 }} 个日目标</div>
        </el-card>
      </el-col>
      <el-col :xs="24" :sm="12" :md="6">
        <el-card class="stat-card">
          <div class="stat-label">日程占用</div>
          <div class="stat-big">{{ minutesFmt(dash?.scheduled_minutes ?? 0) }}</div>
          <div class="stat-sub">{{ dash?.blocks.length ?? 0 }} 个时间块</div>
        </el-card>
      </el-col>
      <el-col :xs="24" :sm="12" :md="6">
        <el-card class="stat-card" :class="{ danger: (dash?.overdue_count ?? 0) > 0 }">
          <div class="stat-label">逾期任务</div>
          <div class="stat-big" :style="{ color: (dash?.overdue_count ?? 0) > 0 ? '#e03131' : 'inherit' }">{{ dash?.overdue_count ?? 0 }}</div>
          <div class="stat-sub">{{ (dash?.overdue_count ?? 0) > 0 ? '需要尽快处理' : '没有逾期，很棒' }}</div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="16" v-loading="loading">
      <el-col :xs="24" :md="8">
        <el-card>
          <template #header>今日目标</template>
          <el-empty v-if="!dash?.day_goals.length" description="今天还没有目标" :image-size="60" />
          <div v-for="g in dash?.day_goals ?? []" :key="g.id" class="row">
            <el-checkbox
              :model-value="checkedGoalIds.has(g.id)"
              :disabled="!!g.date && g.date > today()"
              @change="toggleGoal(g.id, g.date!, g.title, Boolean($event))"
            />
            <span class="row-title" :class="{ done: checkedGoalIds.has(g.id) }">{{ g.title }}</span>
            <el-tag v-if="isMissed(g.date!) && !checkedGoalIds.has(g.id)" size="small" type="danger">未达成</el-tag>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="24" :md="8">
        <el-card>
          <template #header>今日任务</template>
          <el-empty v-if="!dash?.tasks_due.length" description="今天没有到期任务" :image-size="60" />
          <div v-for="t in dash?.tasks_due ?? []" :key="t.id" class="row">
            <el-checkbox :model-value="t.status === 'done'" @change="toggleTask(t.id)" />
            <span class="row-title" :class="{ done: t.status === 'done' }" @click="router.push('/tasks')">{{ t.title }}</span>
            <el-tag v-if="t.priority === 0" size="small" type="danger" effect="dark">P0</el-tag>
            <el-tag v-for="tag in t.tags" :key="tag" size="small" effect="plain">{{ tag }}</el-tag>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="24" :md="8">
        <el-card>
          <template #header>今日日程</template>
          <el-empty v-if="!dash?.blocks.length" description="还没有安排时间块" :image-size="60" />
          <div v-for="b in dash?.blocks ?? []" :key="b.id" class="row">
            <el-icon :color="b.done ? '#2f9e44' : '#4c6ef5'">
              <CircleCheckFilled v-if="b.done" />
              <Clock v-else />
            </el-icon>
            <span class="time-text">{{ Math.floor(b.start_min / 60).toString().padStart(2, '0') }}:{{ String(b.start_min % 60).padStart(2, '0') }}–{{ Math.floor(b.end_min / 60).toString().padStart(2, '0') }}:{{ String(b.end_min % 60).padStart(2, '0') }}</span>
            <span class="row-title" :class="{ done: b.done }">{{ b.title }}</span>
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<style scoped>
.nav-bar {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
}
.date-title {
  font-size: 16px;
  font-weight: 700;
}
.stat-row {
  margin-bottom: 16px;
}
.stat-card {
  text-align: center;
}
.stat-label {
  font-size: 13px;
  color: var(--el-text-color-secondary);
  margin-bottom: 10px;
}
.stat-big {
  font-size: 32px;
  font-weight: 800;
  margin: 14px 0 6px;
}
.stat-sub {
  font-size: 12px;
  color: var(--el-text-color-secondary);
}
.stat-card.danger {
  border-color: rgba(224, 49, 49, 0.4);
}
.row {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 2px;
}
.row-title {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  cursor: pointer;
}
.row-title.done {
  text-decoration: line-through;
  color: var(--el-text-color-secondary);
}
.time-text {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  font-variant-numeric: tabular-nums;
}
</style>