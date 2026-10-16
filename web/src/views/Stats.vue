<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import * as api from '@/api/stats'
import * as goalsApi from '@/api/goals'
import { useTaskStore } from '@/stores/tasks'
import { useSettingsStore } from '@/stores/settings'
import { useChart } from '@/utils/chart'
import { today, weekStartOf, addDays, formatDate } from '@/utils/date'
import type { Dashboard, Heatmap, StatsDaily, StatsWeekly } from '@/types/api'

const router = useRouter()
const store = useTaskStore()
const settings = useSettingsStore()

const weekStart = ref(weekStartOf(today()))
const weekly = ref<StatsWeekly | null>(null)
const heatmap = ref<Heatmap | null>(null)
const loading = ref(false)

const trendRef = ref<HTMLElement | null>(null)
const byGoalRef = ref<HTMLElement | null>(null)
const byTagRef = ref<HTMLElement | null>(null)
const heatRef = ref<HTMLElement | null>(null)
const trendChart = useChart(trendRef)
const byGoalChart = useChart(byGoalRef)
const byTagChart = useChart(byTagRef)
const heatChart = useChart(heatRef)

onMounted(async () => {
  await Promise.all([settings.load(), load()])
})
watch(weekStart, load)

async function load() {
  loading.value = true
  try {
    const ws = weekStart.value
    const hmStart = addDays(weekStartOf(today()), -14 * 7)
    const hmEnd = addDays(weekStartOf(today()), 6)
    const [w, hm] = await Promise.all([api.getStatsWeekly(ws), api.getStatsHeatmap(hmStart, hmEnd)])
    weekly.value = w.data
    heatmap.value = hm.data
    await store.fetchAll()
    renderCharts()
  } finally {
    loading.value = false
  }
}

const pct = (r: number | null) => (r === null ? null : Math.round(r * 100))
const minutesFmt = (m: number) => (m >= 60 ? `${(m / 60).toFixed(1)}h` : `${m}m`)

function renderCharts() {
  const w = weekly.value
  if (!w) return
  // 趋势折线
  trendChart.setOption({
    tooltip: { trigger: 'axis', valueFormatter: (v) => `${v}%` },
    legend: { data: ['任务完成率', '目标达成率'], bottom: 0 },
    grid: { left: 40, right: 16, top: 20, bottom: 44 },
    xAxis: { type: 'category', data: w.trend.map((t) => t.week_start.slice(5)) },
    yAxis: { type: 'value', max: 100, axisLabel: { formatter: '{value}%' } },
    series: [
      {
        name: '任务完成率', type: 'line', smooth: true, connectNulls: false,
        data: w.trend.map((t) => (t.task_rate === null ? null : +(t.task_rate * 100).toFixed(1))),
        areaStyle: { opacity: 0.12 }, itemStyle: { color: '#4c6ef5' },
      },
      {
        name: '目标达成率', type: 'line', smooth: true, connectNulls: false,
        data: w.trend.map((t) => (t.goal_rate === null ? null : +(t.goal_rate * 100).toFixed(1))),
        areaStyle: { opacity: 0.12 }, itemStyle: { color: '#2f9e44' },
      },
    ],
  })
  // 分布：按目标 / 按标签
  const dist = (items: { label: string; done: number; total: number }[], target: ReturnType<typeof useChart>) => {
    const labels = items.map((i) => i.label)
    const done = items.map((i) => i.done)
    const rest = items.map((i) => i.total - i.done)
    target.setOption({
      tooltip: { trigger: 'axis', axisPointer: { type: 'shadow' } },
      legend: { data: ['已完成', '未完成'], bottom: 0 },
      grid: { left: 6, right: 10, top: 10, bottom: 40, containLabel: true },
      xAxis: { type: 'value' },
      yAxis: { type: 'category', data: labels, axisLabel: { width: 90, overflow: 'truncate' } },
      series: [
        { name: '已完成', type: 'bar', stack: 'x', itemStyle: { color: '#40c463' }, data: done },
        { name: '未完成', type: 'bar', stack: 'x', itemStyle: { color: 'rgba(148,163,184,.35)' }, data: rest },
      ],
    })
  }
  dist(w.by_goal, byGoalChart)
  dist(w.by_tag, byTagChart)
}

function renderHeatmap() {
  const hm = heatmap.value
  if (!hm) return
  const cells = hm.cells.filter((c) => c.score !== null).map((c) => [c.date, +(c.score! * 100).toFixed(1)])
  heatChart.setOption({
    tooltip: {
      formatter: (p: unknown) => {
        const data = (p as { data: [string, number] }).data
        return `${data[0]} 综合完成度 ${data[1]}%`
      },
      backgroundColor: 'rgba(24,28,45,.9)',
      textStyle: { color: '#fff' },
    },
    visualMap: {
      min: 0, max: 100, show: false,
      inRange: { color: ['#ebedf0', '#c6e48b', '#7bc96f', '#239a3b', '#196127'] },
    },
    calendar: {
      range: [hm.start, hm.end],
      cellSize: [15, 15],
      top: 26, left: 64, right: 8,
      dayLabel: { firstDay: 1, nameMap: ['日', '一', '二', '三', '四', '五', '六'], fontSize: 10, color: '#94a3b8' },
      monthLabel: { nameMap: 'cn', fontSize: 11, color: '#94a3b8' },
      yearLabel: { show: false },
      itemStyle: { borderWidth: 2, borderColor: 'transparent', borderRadius: 3 },
      splitLine: { show: false },
    },
    series: [{ type: 'heatmap', coordinateSystem: 'calendar', data: cells }],
  })
  heatChart.on('click', (params: any) => {
    const data = params?.data as [string, number] | undefined
    if (data?.[0]) openDrill(String(data[0]))
  })
}
watch(heatmap, renderHeatmap)
watch(() => settings.heatmapTaskWeight, () => load(), { flush: 'post' })
watch(() => settings.heatmapGoalWeight, () => load(), { flush: 'post' })

// ---- 下钻 ----
const drillOpen = ref(false)
const drillDate = ref('')
const drillDaily = ref<StatsDaily | null>(null)
const drillDash = ref<Dashboard | null>(null)
const drillChecked = ref<Set<number>>(new Set())

async function openDrill(date: string) {
  drillDate.value = date
  drillOpen.value = true
  const [d, dash, cks] = await Promise.all([
    api.getStatsDaily(date),
    api.getDashboard(date),
    goalsApi.listCheckins(date, date),
  ])
  drillDaily.value = d.data
  drillDash.value = dash.data
  drillChecked.value = new Set(cks.data.map((c) => c.goal_id))
}

async function toggleDrillTask(id: number) {
  const t = store.byId.get(id)
  if (t) await store.toggleDone(t)
  await openDrill(drillDate.value)
}

async function toggleUnfinished(id: number) {
  const t = store.byId.get(id)
  if (t) await store.toggleDone(t)
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
      <span class="spacer" />
      <el-popover placement="bottom-end" :width="280" trigger="click">
        <template #reference>
          <el-button><el-icon><Operation /></el-icon>热力图权重</el-button>
        </template>
        <div class="weight-item">
          <span>任务权重 {{ settings.heatmapTaskWeight.toFixed(1) }}</span>
          <el-slider :model-value="settings.heatmapTaskWeight" :min="0" :max="1" :step="0.1" @change="(v: number) => settings.set('heatmap_task_weight', String(Number(v)))" />
        </div>
        <div class="weight-item">
          <span>目标权重 {{ settings.heatmapGoalWeight.toFixed(1) }}</span>
          <el-slider :model-value="settings.heatmapGoalWeight" :min="0" :max="1" :step="0.1" @change="(v: number) => settings.set('heatmap_goal_weight', String(Number(v)))" />
        </div>
        <el-text size="small" type="info">综合完成度 = 可用口径的加权平均</el-text>
      </el-popover>
    </div>

    <el-row :gutter="16" v-loading="loading" class="stat-row">
      <el-col :xs="24" :sm="12" :md="6">
        <el-card class="stat-card">
          <div class="stat-label">本周任务完成率</div>
          <el-progress type="circle" :percentage="pct(weekly?.task_completion_rate ?? null) ?? 0" :width="84" :format="() => (pct(weekly?.task_completion_rate ?? null) === null ? '—' : `${pct(weekly!.task_completion_rate)}%`)" />
          <div class="stat-sub">{{ weekly?.task_done ?? 0 }} / {{ weekly?.task_total ?? 0 }} 完成</div>
        </el-card>
      </el-col>
      <el-col :xs="24" :sm="12" :md="6">
        <el-card class="stat-card">
          <div class="stat-label">周目标达成率</div>
          <el-progress type="circle" :percentage="pct(weekly?.goal_achievement_rate ?? null) ?? 0" :width="84" color="#2f9e44" :format="() => (pct(weekly?.goal_achievement_rate ?? null) === null ? '—' : `${pct(weekly!.goal_achievement_rate)}%`)" />
          <div class="stat-sub">{{ weekly?.goal_done ?? 0 }} / {{ weekly?.goal_total ?? 0 }} 达成</div>
        </el-card>
      </el-col>
      <el-col :xs="24" :sm="12" :md="6">
        <el-card class="stat-card">
          <div class="stat-label">本周日程占用</div>
          <div class="stat-big">{{ minutesFmt(weekly?.scheduled_minutes ?? 0) }}</div>
          <div class="stat-sub">已完成块 {{ minutesFmt(weekly?.completed_block_minutes ?? 0) }}</div>
        </el-card>
      </el-col>
      <el-col :xs="24" :sm="12" :md="6">
        <el-card class="stat-card" :class="{ danger: (weekly?.overdue_count ?? 0) > 0 }">
          <div class="stat-label">本周逾期</div>
          <div class="stat-big" :style="{ color: (weekly?.overdue_count ?? 0) > 0 ? '#e03131' : 'inherit' }">{{ weekly?.overdue_count ?? 0 }}</div>
          <div class="stat-sub">{{ (weekly?.overdue_count ?? 0) > 0 ? '尽快清理' : '没有逾期' }}</div>
        </el-card>
      </el-col>
    </el-row>

    <el-card v-loading="loading" class="chart-card">
      <template #header>完成度热力图（点击格子下钻到日详情）</template>
      <div ref="heatRef" class="heat-box" />
    </el-card>

    <el-row :gutter="16" v-loading="loading">
      <el-col :xs="24" :md="16">
        <el-card class="chart-card">
          <template #header>近 8 周完成率趋势</template>
          <div ref="trendRef" class="chart-box" />
        </el-card>
      </el-col>
      <el-col :xs="24" :md="8">
        <el-card class="chart-card">
          <template #header>本周未完成任务</template>
          <el-empty v-if="!weekly?.unfinished.length" description="全部完成 🎉" :image-size="50" />
          <div v-for="t in weekly?.unfinished ?? []" :key="t.id" class="row">
            <el-checkbox :model-value="false" @change="toggleUnfinished(t.id)" />
            <span class="row-title" @click="router.push('/tasks')">{{ t.title }}</span>
            <el-tag v-if="t.due_date" size="small" type="warning" effect="plain">{{ formatDate(t.due_date) }}</el-tag>
          </div>
        </el-card>
      </el-col>
    </el-row>

    <el-row :gutter="16" v-loading="loading">
      <el-col :xs="24" :md="12">
        <el-card class="chart-card">
          <template #header>任务分布 · 按目标</template>
          <el-empty v-if="!weekly?.by_goal.length" description="本周没有关联目标的任务" :image-size="50" />
          <div ref="byGoalRef" v-else class="chart-box" />
        </el-card>
      </el-col>
      <el-col :xs="24" :md="12">
        <el-card class="chart-card">
          <template #header>任务分布 · 按标签</template>
          <el-empty v-if="!weekly?.by_tag.length" description="本周任务没有标签" :image-size="50" />
          <div ref="byTagRef" v-else class="chart-box" />
        </el-card>
      </el-col>
    </el-row>

    <el-dialog v-model="drillOpen" :title="`${drillDate} · 日详情`" width="620px">
      <el-descriptions :column="2" border v-if="drillDaily">
        <el-descriptions-item label="任务完成率">{{ pct(drillDaily.task_completion_rate) === null ? '—' : pct(drillDaily.task_completion_rate) + '%' }}</el-descriptions-item>
        <el-descriptions-item label="目标达成率">{{ pct(drillDaily.goal_achievement_rate) === null ? '—' : pct(drillDaily.goal_achievement_rate) + '%' }}</el-descriptions-item>
        <el-descriptions-item label="任务">{{ drillDaily.task_done }}/{{ drillDaily.task_total }}</el-descriptions-item>
        <el-descriptions-item label="目标">{{ drillDaily.goal_done }}/{{ drillDaily.goal_total }}</el-descriptions-item>
        <el-descriptions-item label="日程占用">{{ minutesFmt(drillDaily.scheduled_minutes) }}</el-descriptions-item>
        <el-descriptions-item label="逾期">{{ drillDaily.overdue_count }}</el-descriptions-item>
      </el-descriptions>
      <div class="drill-section">
        <div class="drill-title">日目标</div>
        <el-empty v-if="!drillDash?.day_goals.length" description="无" :image-size="40" />
        <div v-for="g in drillDash?.day_goals ?? []" :key="g.id" class="row">
          <el-icon :color="drillChecked.has(g.id) ? '#2f9e44' : '#94a3b8'">
            <CircleCheckFilled v-if="drillChecked.has(g.id)" />
            <Circle v-else />
          </el-icon>
          <span class="row-title" :class="{ done: drillChecked.has(g.id) }">{{ g.title }}</span>
        </div>
      </div>
      <div class="drill-section">
        <div class="drill-title">到期任务</div>
        <el-empty v-if="!drillDash?.tasks_due.length" description="无" :image-size="40" />
        <div v-for="t in drillDash?.tasks_due ?? []" :key="t.id" class="row">
          <el-checkbox :model-value="t.status === 'done'" @change="toggleDrillTask(t.id)" />
          <span class="row-title" :class="{ done: t.status === 'done' }">{{ t.title }}</span>
        </div>
      </div>
    </el-dialog>
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
.chart-card {
  margin-bottom: 16px;
}
.chart-box {
  height: 260px;
}
.heat-box {
  height: 170px;
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
.row-title.done {
  text-decoration: line-through;
  color: var(--el-text-color-secondary);
}
.weight-item {
  margin-bottom: 8px;
  font-size: 13px;
}
.drill-section {
  margin-top: 14px;
}
.drill-title {
  font-weight: 700;
  margin-bottom: 6px;
}
</style>