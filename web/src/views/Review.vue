<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { MdEditor } from 'md-editor-v3'
import 'md-editor-v3/lib/style.css'
import { ElMessage, ElMessageBox } from 'element-plus'
import * as api from '@/api/reviews'
import { today, weekStartOf, addDays } from '@/utils/date'
import type { Review, Snapshot } from '@/types/api'

// md-editor-v3 默认语言即中文，无需配置

const weekStart = ref(weekStartOf(today()))
const reviews = ref<Review[]>([])
const reviewMap = computed(() => new Map(reviews.value.map((r) => [r.week_start, r])))
const title = ref('')
const content = ref('')
const snapshot = ref<Snapshot | null>(null)
const loading = ref(false)
const dirty = ref(false)
const lastSavedAt = ref<string | null>(null)

// 近 12 周列表（含未写周）
const weekItems = computed(() =>
  Array.from({ length: 12 }, (_, i) => {
    const ws = addDays(weekStartOf(today()), -7 * (11 - i))
    const r = reviewMap.value.get(ws)
    return { weekStart: ws, weekEnd: addDays(ws, 6), review: r }
  }),
)

onMounted(async () => {
  await Promise.all([loadList(), openWeek(weekStart.value)])
})

async function loadList() {
  const res = await api.listReviews()
  reviews.value = res.data
}

let draftTimer: number | undefined
watch([title, content], () => {
  dirty.value = true
  window.clearTimeout(draftTimer)
  draftTimer = window.setTimeout(() => {
    if (dirty.value) saveDraft()
  }, 800)
})

const draftKey = (ws: string) => `sp-review-draft-${ws}`

function saveDraft() {
  if (!title.value.trim() && !content.value.trim()) return
  localStorage.setItem(draftKey(weekStart.value), JSON.stringify({ title: title.value, content: content.value }))
}
function loadDraft() {
  const raw = localStorage.getItem(draftKey(weekStart.value))
  if (raw) {
    try {
      const d = JSON.parse(raw) as { title: string; content: string }
      title.value = d.title
      content.value = d.content
      return true
    } catch {
      /* 忽略损坏草稿 */
    }
  }
  return false
}
function clearDraft() {
  localStorage.removeItem(draftKey(weekStart.value))
}

async function openWeek(ws: string) {
  weekStart.value = ws
  loading.value = true
  dirty.value = false
  lastSavedAt.value = null
  try {
    try {
      const res = await api.getReview(ws)
      const r = res.data
      title.value = r.title
      content.value = r.content_md
      snapshot.value = r.stats_snapshot
      lastSavedAt.value = r.updated_at
    } catch {
      // 404：新周，优先草稿
      snapshot.value = null
      if (!loadDraft()) {
        title.value = ''
        content.value = ''
      }
    }
  } finally {
    loading.value = false
  }
}
watch(weekStart, (ws) => openWeek(ws))

// ---- 快照渲染与模板 ----
const pct = (r: number | null) => (r === null ? '—' : `${Math.round(r * 100)}%`)
const minutesFmt = (m: number) => (m >= 60 ? `${(m / 60).toFixed(1)}h` : `${m}m`)

function renderSnapshotSection(s: Snapshot): string {
  const lines = [
    '## 本周数据快照',
    `> 生成于 ${s.generated_at}（保存后冻结，可随时手动刷新）`,
    '',
    `- 任务完成率：${pct(s.task_completion_rate)}（${s.task_done}/${s.task_total}）`,
    `- 目标达成率：${pct(s.goal_achievement_rate)}（${s.goal_done}/${s.goal_total}）`,
    `- 打卡：${s.checkin_total}/${s.checkin_days_total} 天`,
    `- 日程占用：${minutesFmt(s.scheduled_minutes)}（已完成 ${minutesFmt(s.completed_block_minutes)}）`,
    `- 逾期任务：${s.overdue_count}`,
  ]
  if (s.unfinished_tasks.length) lines.push(`- 未完成：${s.unfinished_tasks.join('、')}`)
  if (s.by_tag.length) lines.push(`- 标签分布：${s.by_tag.map((x) => `${x.label} ${x.done}/${x.total}`).join('，')}`)
  lines.push('')
  return lines.join('\n')
}

function buildTemplate(s: Snapshot, ws: string): string {
  return [
    `# ${ws} 周复盘（${ws} ~ ${addDays(ws, 6)}）`,
    '',
    renderSnapshotSection(s),
    '## Keep（继续保持）',
    '- ',
    '',
    '## Problem（存在问题）',
    '- ',
    '',
    '## Try（下周尝试）',
    '- ',
    '',
    '## 下周计划',
    '- ',
    '',
  ].join('\n')
}

async function applyTemplate() {
  const res = await api.generateSnapshot(weekStart.value)
  snapshot.value = res.data
  content.value = buildTemplate(res.data, weekStart.value)
  if (!title.value) title.value = `${weekStart.value} 周复盘`
  dirty.value = true
  ElMessage.success('模板已生成，记得保存')
}

async function refreshSnapshot() {
  const res = await api.generateSnapshot(weekStart.value)
  snapshot.value = res.data
  // 替换内容中的快照段；若不存在则插入到开头
  const fresh = renderSnapshotSection(res.data)
  const heading = /^## 本周数据快照[\s\S]*?(?=^## |\s*$)/m
  if (heading.test(content.value)) {
    content.value = content.value.replace(heading, fresh + '\n')
  } else {
    content.value = fresh + '\n' + content.value
  }
  dirty.value = true
  ElMessage.success('快照已刷新，保存后生效')
}

async function save() {
  if (!title.value.trim() && !content.value.trim()) {
    ElMessage.warning('还没有内容可保存')
    return
  }
  await api.putReview(weekStart.value, {
    title: title.value.trim() || `${weekStart.value} 周复盘`,
    content_md: content.value,
    stats_snapshot: snapshot.value,
  })
  dirty.value = false
  lastSavedAt.value = new Date().toLocaleString()
  clearDraft()
  await loadList()
  ElMessage.success('复盘已保存')
}

async function remove() {
  await ElMessageBox.confirm('删除该周复盘？（任务/目标数据不受影响）', '删除复盘', { type: 'warning' })
  await api.deleteReview(weekStart.value)
  clearDraft()
  title.value = ''
  content.value = ''
  snapshot.value = null
  await loadList()
  ElMessage.success('已删除')
}

const badgePct = (r: Review | undefined) => {
  const rate = r?.stats_snapshot?.task_completion_rate
  return rate === null || rate === undefined ? null : Math.round(rate * 100)
}

const weekLabel = (ws: string) => `${ws.slice(5).replace('-', '/')} ~ ${addDays(ws, 6).slice(5).replace('-', '/')}`
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
      <span class="week-title">{{ weekLabel(weekStart) }}</span>
      <span class="spacer" />
      <el-button @click="applyTemplate"><el-icon><MagicStick /></el-icon>从模板创建</el-button>
      <el-button @click="refreshSnapshot" :disabled="!snapshot"><el-icon><Refresh /></el-icon>刷新快照</el-button>
      <el-button v-if="reviewMap.get(weekStart)" type="danger" plain @click="remove">删除复盘</el-button>
      <el-button type="primary" @click="save">保存</el-button>
    </div>

    <el-row :gutter="16" v-loading="loading">
      <el-col :xs="24" :md="5">
        <el-card class="week-list">
          <template #header>近 12 周</template>
          <div
            v-for="it in weekItems"
            :key="it.weekStart"
            class="week-item"
            :class="{ active: it.weekStart === weekStart, today: it.weekStart === weekStartOf(today()) }"
            @click="openWeek(it.weekStart)"
          >
            <div class="wi-head">
              <span class="wi-label">{{ weekLabel(it.weekStart) }}</span>
              <el-tag v-if="badgePct(it.review) !== null" size="small" :type="badgePct(it.review)! >= 80 ? 'success' : badgePct(it.review)! >= 50 ? 'warning' : 'info'" effect="plain">
                {{ badgePct(it.review) }}%
              </el-tag>
              <el-tag v-else-if="it.review" size="small" type="info">已写</el-tag>
              <el-tag v-else size="small" type="info" effect="plain">未写</el-tag>
            </div>
            <div v-if="it.review?.title" class="wi-title">{{ it.review.title }}</div>
          </div>
        </el-card>
      </el-col>
      <el-col :xs="24" :md="19">
        <el-card>
          <div class="editor-head">
            <el-input v-model="title" placeholder="复盘标题" class="title-input" />
            <div class="save-state">
              <el-tag v-if="dirty" size="small" type="warning">未保存</el-tag>
              <el-tag v-else-if="lastSavedAt" size="small" type="success">已保存 · {{ lastSavedAt }}</el-tag>
              <el-tag v-else size="small" type="info">新复盘</el-tag>
            </div>
          </div>
          <MdEditor v-model="content" class="md-editor" placeholder="开始写复盘：Keep / Problem / Try…" />
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
.week-list {
  min-height: 520px;
}
.week-item {
  padding: 8px 10px;
  border-radius: 8px;
  cursor: pointer;
  margin-bottom: 2px;
}
.week-item:hover {
  background: var(--el-fill-color-light);
}
.week-item.active {
  background: rgba(76, 110, 245, 0.12);
}
.week-item.today .wi-label {
  font-weight: 700;
  color: var(--el-color-primary);
}
.wi-head {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
}
.wi-label {
  flex: 1;
  color: var(--el-text-color-secondary);
  font-variant-numeric: tabular-nums;
}
.wi-title {
  font-size: 12px;
  color: var(--el-text-color-secondary);
  margin-top: 2px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.editor-head {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 10px;
}
.title-input {
  max-width: 380px;
}
.save-state {
  flex: 1;
}
.md-editor {
  height: 520px;
}
</style>