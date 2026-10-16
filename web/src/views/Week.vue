<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import dayjs from 'dayjs'
import * as api from '@/api/stats'
import * as reviewsApi from '@/api/reviews'
import { useGoalStore } from '@/stores/goals'
import { useTaskStore } from '@/stores/tasks'
import { useSettingsStore } from '@/stores/settings'
import { today, weekStartOf, addDays, weekDays, weekLabels } from '@/utils/date'
import type { Heatmap, Review, StatsWeekly } from '@/types/api'
import type { Goal } from '@/types/api'

const router = useRouter()
const gstore = useGoalStore()
const tstore = useTaskStore()
const settings = useSettingsStore()

const weekStart = ref(weekStartOf(today()))
const weekly = ref<StatsWeekly | null>(null)
const heatmap = ref<Heatmap | null>(null)
const reviews = ref<Review[]>([])
const loading = ref(false)

const weekRatePct = computed(() => (gstore.weekRate === null ? null : Math.round(gstore.weekRate * 100)))
const currentReview = computed(() => reviews.value.find((r) => r.week_start === weekStart.value))
const reviewPct = computed(() => {
  const rate = currentReview.value?.stats_snapshot?.task_completion_rate
  return rate === null || rate === undefined ? null : Math.round(rate * 100)
})

// 一周 7 天 × 当日完成度（数据来自热力图接口）
const weekCells = computed(() => {
  const cells = new Map((heatmap.value?.cells ?? []).map((c) => [c.date, c]))
  return weekDays(weekStart.value).map((date) => ({
    date,
    cell: cells.get(date) ?? null,
    weekday: ['日', '一', '二', '三', '四', '五', '六'][dayjs(date).day()],
    isToday: date === today(),
  }))
})

onMounted(async () => {
  await Promise.all([settings.load(), load()])
})
watch(weekStart, load)

async function load() {
  loading.value = true
  try {
    const ws = weekStart.value
    const [w, hm, revs] = await Promise.all([
      api.getStatsWeekly(ws),
      api.getStatsHeatmap(ws, addDays(ws, 6)),
      reviewsApi.listReviews(),
    ])
    weekly.value = w.data
    heatmap.value = hm.data
    reviews.value = revs.data
    await Promise.all([gstore.loadWeek(ws), tstore.fetchAll()])
  } finally {
    loading.value = false
  }
}

const pct = (r: number | null) => (r === null ? null : Math.round(r * 100))
const minutesFmt = (m: number) => (m >= 60 ? `${(m / 60).toFixed(1)}h` : `${m}m`)
const reviewTimeFmt = (s: string) => dayjs(s).format('M月D日 HH:mm')

async function toggleWeekDay(goal: Goal, date: string) {
  await gstore.toggleCheckin(goal, date)
  await load()
}
async function toggleOnce(goal: Goal) {
  const status = goal.status === 'done' ? 'active' : 'done'
  await gstore.updateGoal(goal.id, {
    title: goal.title,
    goal_type: 'week',
    goal_mode: 'once',
    week_start: goal.week_start,
    status,
  })
  await load()
}
async function toggleUnfinished(id: number) {
  const t = tstore.byId.get(id)
  if (t) await tstore.toggleDone(t)
  await load()
}
</script>

<template>
  <div class="page">
    <div class="nav-bar">
      <el-button-group>
        <el-button @click="weekStart = addDays(weekStart, -7)"><el-icon><ArrowLeft /></el-icon></el-button>
        <el-button @click="weekStart = weekStartOf(today())">本周</el-button>
        <el-button @click="weekStart = addDays(weekStart, 7)"><el-icon><ArrowRight /></el-icon></el-button>
      </el-button-group>
      <el-date-picker v-model="weekStart" type="week" value-format="YYYY-MM-DD" style="width: 150px" />
      <span class="week-title">{{ weekly?.week_start }} ~ {{ weekly?.week_end }}</span>
      <el-tag v-if="weekStart === weekStartOf(today())" size="small" type="success" effect="plain">本周</el-tag>
      <span class="spacer" />
      <el-button @click="router.push('/review')"><el-icon><EditPen /></el-icon>写复盘</el-button>
      <el-button @click="router.push('/stats')"><el-icon><TrendCharts /></el-icon>详细统计</el-button>
    </div>

    <el-row :gutter="16" v-loading="loading" class="stat-row">
      <el-col :xs="12" :md="6">
        <el-card class="stat-card">
          <div class="stat-label">任务完成率</div>
          <el-progress type="circle" :percentage="pct(weekly?.task_completion_rate ?? null) ?? 0" :width="84"
            :format="() => (pct(weekly?.task_completion_rate ?? null) === null ? '—' : `${pct(weekly!.task_completion_rate)}%`)" />
          <div class="stat-sub">{{ weekly?.task_done ?? 0 }} / {{ weekly?.task_total ?? 0 }} 完成</div>
        </el-card>
      </el-col>
      <el-col :xs="12" :md="6">
        <el-card class="stat-card">
          <div class="stat-label">周目标达成率</div>
          <el-progress type="circle" :percentage="pct(weekly?.goal_achievement_rate ?? null) ?? 0" :width="84" color="#2f9e44"
            :format="() => (pct(weekly?.goal_achievement_rate ?? null) === null ? '—' : `${pct(weekly!.goal_achievement_rate)}%`)" />
          <div class="stat-sub">{{ weekly?.goal_done ?? 0 }} / {{ weekly?.goal_total ?? 0 }} 达成</div>
        </el-card>
      </el-col>
      <el-col :xs="12" :md="6">
        <el-card class="stat-card">
          <div class="stat-label">日程占用</div>
          <div class="stat-big">{{ minutesFmt(weekly?.scheduled_minutes ?? 0) }}</div>
          <div class="stat-sub">已完成 {{ minutesFmt(weekly?.completed_block_minutes ?? 0) }}</div>
        </el-card>
      </el-col>
      <el-col :xs="12" :md="6">
        <el-card class="stat-card" :class="{ danger: (weekly?.overdue_count ?? 0) > 0 }">
          <div class="stat-label">逾期任务</div>
          <div class="stat-big" :style="{ color: (weekly?.overdue_count ?? 0) > 0 ? '#e03131' : 'inherit' }">{{ weekly?.overdue_count ?? 0 }}</div>
          <div class="stat-sub">{{ (weekly?.overdue_count ?? 0) > 0 ? '点击下方列表清理' : '没有逾期 🎉' }}</div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="16" v-loading="loading">
      <el-col :xs="24" :md="14">
        <el-card class="block-card">
          <template #header>本周每日完成度</template>
          <div class="day-strip">
            <div v-for="d in weekCells" :key="d.date" class="day-cell" :class="{ today: d.isToday }">
              <div class="day-head">
                <span class="day-week">周{{ d.weekday }}</span>
                <span class="day-date">{{ d.date.slice(5).replace('-', '/') }}</span>
              </div>
              <template v-if="d.cell">
                <div class="day-bar">
                  <span class="bar-label">任务</span>
                  <el-progress :percentage="pct(d.cell.task_rate) ?? 0" :stroke-width="6" :show-text="false"
                    :color="d.cell.task_rate === null ? '#cbd5e1' : '#4c6ef5'" />
                  <span class="bar-num">{{ d.cell.task_done }}/{{ d.cell.task_total }}</span>
                </div>
                <div class="day-bar">
                  <span class="bar-label">目标</span>
                  <el-progress :percentage="pct(d.cell.goal_rate) ?? 0" :stroke-width="6" :show-text="false"
                    :color="d.cell.goal_rate === null ? '#cbd5e1' : '#2f9e44'" />
                  <span class="bar-num">{{ d.cell.goal_done }}/{{ d.cell.goal_total }}</span>
                </div>
              </template>
              <div v-else class="day-empty">—</div>
            </div>
          </div>
          <div class="strip-hint">点击顶部「今日」可管理当天目标与任务</div>
        </el-card>
      </el-col>
      <el-col :xs="24" :md="10">
        <el-card class="block-card">
          <template #header>
            <div class="card-header">
              <span>周目标</span>
              <el-progress v-if="weekRatePct !== null" type="circle" :percentage="weekRatePct" :width="38" />
              <el-button type="primary" size="small" @click="router.push('/goals')">管理</el-button>
            </div>
          </template>
          <el-empty v-if="!gstore.weekGoals.length" description="暂无周目标，去「目标」页创建" :image-size="60">
            <el-button type="primary" @click="router.push('/goals')">去创建</el-button>
          </el-empty>
          <div v-for="g in gstore.weekGoals" :key="g.id" class="goal-row">
            <div class="goal-main">
              <span class="goal-title" :class="{ done: g.status === 'done' }">{{ g.title }}</span>
              <el-tag size="small" :type="g.goal_mode === 'once' ? 'warning' : 'success'" effect="plain">
                {{ g.goal_mode === 'once' ? '一次性' : `打卡 ${gstore.checked[g.id]?.length ?? 0}/${g.checkin_days}` }}
              </el-tag>
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

    <el-row :gutter="16" v-loading="loading" class="row-2">
      <el-col :xs="24" :md="14">
        <el-card class="block-card">
          <template #header>
            <div class="card-header">
              <span>本周未完成任务</span>
              <el-button link type="primary" size="small" @click="router.push('/tasks')">全部任务 →</el-button>
            </div>
          </template>
          <el-empty v-if="!weekly?.unfinished.length" description="全部完成 🎉" :image-size="50" />
          <div v-for="t in weekly?.unfinished ?? []" :key="t.id" class="row">
            <el-checkbox :model-value="false" @change="toggleUnfinished(t.id)" />
            <span class="row-title" @click="router.push('/tasks')">{{ t.title }}</span>
            <el-tag v-if="t.due_date" size="small" type="warning" effect="plain">{{ t.due_date.slice(5).replace('-', '/') }}</el-tag>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="24" :md="10">
        <el-card class="block-card review-card">
          <template #header>本周复盘</template>
          <template v-if="currentReview">
            <div class="review-head">
              <el-tag type="success">已复盘</el-tag>
              <el-tag v-if="reviewPct !== null" size="small" type="info" effect="plain">任务完成率 {{ reviewPct }}%</el-tag>
              <span class="spacer" />
              <el-button size="small" @click="router.push('/review')">打开</el-button>
            </div>
            <div class="review-title">{{ currentReview.title }}</div>
            <div class="review-meta">保存于 {{ reviewTimeFmt(currentReview.updated_at) }} · 数据快照已冻结</div>
          </template>
          <template v-else>
            <el-empty description="本周还没有复盘" :image-size="60">
              <el-button type="primary" @click="router.push('/review')">去写复盘</el-button>
            </el-empty>
          </template>
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
.week-title {
  font-weight: 700;
}
.spacer {
  flex: 1;
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
  font-size: 30px;
  font-weight: 800;
  margin: 12px 0 6px;
}
.stat-sub {
  font-size: 12px;
  color: var(--el-text-color-secondary);
}
.stat-card.danger {
  border-color: rgba(224, 49, 49, 0.4);
}
.block-card {
  margin-bottom: 16px;
}
.row-2 {
  margin-bottom: 8px;
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

.day-strip {
  display: grid;
  grid-template-columns: repeat(7, 1fr);
  gap: 8px;
}
.day-cell {
  border: 1px solid rgba(148, 163, 184, 0.2);
  border-radius: 10px;
  padding: 8px 10px;
  background: rgba(148, 163, 184, 0.06);
}
.day-cell.today {
  border-color: rgba(76, 110, 245, 0.55);
  background: rgba(76, 110, 245, 0.08);
}
.day-head {
  display: flex;
  justify-content: space-between;
  align-items: baseline;
  margin-bottom: 8px;
}
.day-week {
  font-weight: 700;
  font-size: 13px;
}
.day-cell.today .day-week {
  color: #3b5bdb;
}
.day-date {
  font-size: 11px;
  color: var(--el-text-color-secondary);
  font-variant-numeric: tabular-nums;
}
.day-bar {
  display: flex;
  align-items: center;
  gap: 6px;
  margin: 4px 0;
}
.bar-label {
  font-size: 11px;
  color: var(--el-text-color-secondary);
  flex-shrink: 0;
  width: 22px;
}
.bar-num {
  font-size: 11px;
  color: var(--el-text-color-secondary);
  font-variant-numeric: tabular-nums;
  flex-shrink: 0;
}
.day-empty {
  color: var(--el-text-color-secondary);
  font-size: 12px;
  padding: 8px 0;
}
.strip-hint {
  margin-top: 10px;
  font-size: 12px;
  color: var(--el-text-color-secondary);
}
@media (max-width: 900px) {
  .day-strip {
    grid-template-columns: repeat(4, 1fr);
  }
}

.goal-row {
  display: flex;
  flex-direction: column;
  gap: 6px;
  padding: 7px 4px;
  border-bottom: 1px dashed rgba(148, 163, 184, 0.18);
}
.goal-row:last-child {
  border-bottom: none;
}
.goal-main {
  display: flex;
  align-items: center;
  gap: 8px;
}
.goal-title {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.goal-title.done {
  text-decoration: line-through;
  color: var(--el-text-color-secondary);
}
.week-checkins {
  display: flex;
  gap: 6px;
  padding-left: 2px;
}

.row {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 5px 2px;
}
.row-title {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  cursor: pointer;
}

.review-head {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 10px;
}
.review-title {
  font-weight: 600;
  margin-bottom: 6px;
}
.review-meta {
  font-size: 12px;
  color: var(--el-text-color-secondary);
}
</style>
