<script setup lang="ts">
import { computed, onMounted, onUnmounted, reactive, ref, watch } from 'vue'
import dayjs from 'dayjs'
import { ElMessage, ElMessageBox } from 'element-plus'
import * as api from '@/api/blocks'
import { useTaskStore } from '@/stores/tasks'
import { useSettingsStore } from '@/stores/settings'
import BlockEditor from '@/components/BlockEditor.vue'
import { today, addDays, weekStartOf, weekDays, weekLabels, formatDate } from '@/utils/date'
import type { Block } from '@/types/api'

const HOUR_H = 48 // 每小时像素
const settings = useSettingsStore()
const store = useTaskStore()

const date = ref(today())
const view = ref<'day' | 'week'>('day')
const blocks = ref<Block[]>([])
const loading = ref(false)
const now = ref(dayjs())
const timelineEl = ref<HTMLElement | null>(null)

let timer: number | undefined
onMounted(async () => {
  await load()
  timer = window.setInterval(() => (now.value = dayjs()), 30000)
  scrollToNow()
})
onUnmounted(() => window.clearInterval(timer))
watch([date, view], load)

async function load() {
  loading.value = true
  try {
    const from = view.value === 'day' ? date.value : weekStartOf(date.value)
    const to = view.value === 'day' ? date.value : addDays(weekStartOf(date.value), 6)
    const res = await api.listBlocks(from, to)
    blocks.value = res.data
    await store.fetchAll()
  } finally {
    loading.value = false
  }
}

// ---- 编辑器 ----
const editorOpen = ref(false)
const editing = ref<Block | null>(null)
const draftStart = ref(9 * 60)

function openCreate(startMin?: number) {
  editing.value = null
  draftStart.value = startMin ?? Math.round(nowMin() / 30) * 30
  editorOpen.value = true
}
function openEdit(b: Block) {
  editing.value = b
  editorOpen.value = true
}
async function removeBlock(b: Block) {
  await ElMessageBox.confirm('删除该时间块？', `删除「${b.title}」`, { type: 'warning' })
  await api.deleteBlock(b.id)
  await load()
}

// ---- 时间工具 ----
function nowMin(): number {
  return now.value.hour() * 60 + now.value.minute()
}
function topOf(b: Block): number {
  return (b.start_min / 60) * HOUR_H
}
function heightOf(b: Block): number {
  return ((b.end_min - b.start_min) / 60) * HOUR_H
}

const dayBlocks = computed(() =>
  [...blocks.value].filter((b) => b.block_date === date.value).sort((a, b) => a.start_min - b.start_min),
)

// 重叠检测（B-05）：交集块加警示边框
const overlapIds = computed(() => {
  const set = new Set<number>()
  const list = dayBlocks.value
  for (let i = 0; i < list.length; i++) {
    for (let j = i + 1; j < list.length; j++) {
      if (list[j].start_min < list[i].end_min) {
        set.add(list[i].id)
        set.add(list[j].id)
      } else break
    }
  }
  return set
})

const taskDone = computed(() => {
  const m = new Map<number, boolean>()
  for (const t of store.tasks) m.set(t.id, t.status === 'done')
  return m
})

// ---- 点击空白创建（B-02） ----
function onGridClick(e: MouseEvent) {
  const el = timelineEl.value
  if (!el || e.target !== e.currentTarget) return
  const rect = el.getBoundingClientRect()
  const y = e.clientY - rect.top + el.scrollTop
  openCreate(Math.round((y / HOUR_H) * 60 / 30) * 30)
}

// ---- 拖动平移（B-06，Pointer 事件） ----
const drag = reactive({ active: false, moved: false, block: null as Block | null, startY: 0, startMin: 0, top: 0 })

function onBlockPointerDown(e: PointerEvent, b: Block) {
  if (e.button !== 0) return
  drag.active = true
  drag.moved = false
  drag.block = b
  drag.startY = e.clientY
  drag.startMin = b.start_min
  drag.top = topOf(b)
  window.addEventListener('pointermove', onPointerMove)
  window.addEventListener('pointerup', onPointerUp)
}
function onPointerMove(e: PointerEvent) {
  if (!drag.active || !drag.block) return
  const dy = e.clientY - drag.startY
  if (Math.abs(dy) > 3) drag.moved = true
  const dur = drag.block.end_min - drag.block.start_min
  const raw = Math.round((drag.startMin + (dy / HOUR_H) * 60) / 30) * 30
  drag.top = ((Math.max(0, Math.min(1440 - dur, raw))) / 60) * HOUR_H
}
async function onPointerUp() {
  window.removeEventListener('pointermove', onPointerMove)
  window.removeEventListener('pointerup', onPointerUp)
  if (!drag.active || !drag.block) return
  const b = drag.block
  const moved = drag.moved
  const newStart = Math.round((drag.top / HOUR_H) * 60)
  drag.active = false
  drag.block = null
  if (moved && newStart !== b.start_min) {
    const dur = b.end_min - b.start_min
    await api.updateBlock(b.id, {
      block_date: b.block_date,
      start_min: newStart,
      end_min: newStart + dur,
      title: b.title,
      task_id: b.task_id,
      color: b.color,
      note: b.note,
    })
    ElMessage.success(`已移动到 ${fmt(newStart)}`)
    await load()
  }
}
function fmt(m: number): string {
  const h = Math.floor(m / 60)
  const mm = m % 60
  return `${String(h).padStart(2, '0')}:${String(mm).padStart(2, '0')}`
}

// ---- 拖拽缩放（P1：块上下边缘改起止时间，30 分钟吸附） ----
const resize = reactive({ active: false, block: null as Block | null, edge: '' as 'top' | 'bottom', startY: 0, suppress: false })

function onResizePointerDown(e: PointerEvent, b: Block, edge: 'top' | 'bottom') {
  if (e.button !== 0) return
  resize.active = true
  resize.block = b
  resize.edge = edge
  resize.startY = e.clientY
  resize.suppress = false
  window.addEventListener('pointermove', onResizeMove)
  window.addEventListener('pointerup', onResizeUp)
}
function onResizeMove(e: PointerEvent) {
  if (!resize.active || !resize.block) return
  const b = resize.block
  const dy = e.clientY - resize.startY
  const step = Math.round((dy / HOUR_H) * 60 / 30) * 30
  if (resize.edge === 'top') {
    b.start_min = Math.max(0, Math.min(b.end_min - 30, b.start_min + step))
  } else {
    b.end_min = Math.max(b.start_min + 30, Math.min(1440, b.end_min + step))
  }
}
async function onResizeUp() {
  window.removeEventListener('pointermove', onResizeMove)
  window.removeEventListener('pointerup', onResizeUp)
  if (!resize.active || !resize.block) return
  const b = resize.block
  resize.active = false
  resize.block = null
  resize.suppress = true
  await api.updateBlock(b.id, { block_date: b.block_date, start_min: b.start_min, end_min: b.end_min, title: b.title, task_id: b.task_id, color: b.color, note: b.note })
  ElMessage.success(`已调整到 ${fmt(b.start_min)}–${fmt(b.end_min)}`)
  await load()
}

// ---- 任务拖入时间轴（B-07） ----
let dragTaskId: number | null = null
function onTaskDragStart(e: DragEvent, id: number) {
  dragTaskId = id
  if (e.dataTransfer) e.dataTransfer.effectAllowed = 'copy'
}
function onTimelineDragOver(e: DragEvent) {
  if (dragTaskId !== null) e.preventDefault()
}
async function onTimelineDrop(e: DragEvent) {
  if (dragTaskId === null) return
  const task = store.byId.get(dragTaskId)
  dragTaskId = null
  if (!task) return
  const el = timelineEl.value
  if (!el) return
  const rect = el.getBoundingClientRect()
  const y = e.clientY - rect.top + el.scrollTop
  const start = Math.max(0, Math.min(1440, Math.round((y / HOUR_H) * 60 / 30) * 30))
  const dur = Math.min(task.est_minutes ?? 60, 1440 - start)
  await api.createBlock({ block_date: date.value, start_min: start, end_min: start + dur, title: task.title, task_id: task.id })
  ElMessage.success(`已排入日程：${task.title}`)
  await load()
}

// ---- 侧栏任务 ----
const taskQuery = ref('')
const openTasks = computed(() =>
  store.tasks
    .filter((t) => t.status !== 'done')
    .filter((t) => !taskQuery.value || t.title.includes(taskQuery.value)),
)

// ---- 周视图 ----
const weekCols = computed(() => weekDays(weekStartOf(date.value)).map((d, i) => ({ date: d, label: `周${weekLabels(settings.weekStartDay)[i]}`, blocks: blocks.value.filter((b) => b.block_date === d).sort((a, b) => a.start_min - b.start_min) })))

function scrollToNow() {
  requestAnimationFrame(() => {
    const el = timelineEl.value
    if (el) el.scrollTop = Math.max(0, (nowMin() / 60) * HOUR_H - 260)
  })
}

const hourLabels = Array.from({ length: 24 }, (_, h) => h)
const halfHours = Array.from({ length: 48 }, (_, i) => i)

function blockStyle(b: Block) {
  const done = b.task_id != null && taskDone.value.get(b.task_id)
  return {
    top: `${topOf(b)}px`,
    height: `${heightOf(b)}px`,
    background: done ? '#e6f7ee' : b.color ? `${b.color}1a` : '#eef2ff',
    borderColor: done ? '#2f9e44' : b.color || '#4c6ef5',
  }
}
</script>

<template>
  <div class="page">
    <div class="nav-bar">
      <div class="nav-group">
        <el-button-group>
          <el-button @click="date = addDays(date, -1)"><el-icon><ArrowLeft /></el-icon></el-button>
          <el-button @click="date = today()">今天</el-button>
          <el-button @click="date = addDays(date, 1)"><el-icon><ArrowRight /></el-icon></el-button>
        </el-button-group>
        <el-date-picker v-model="date" type="date" value-format="YYYY-MM-DD" style="width: 150px" />
      </div>
      <el-radio-group v-model="view">
        <el-radio-button value="day">日视图</el-radio-button>
        <el-radio-button value="week">周概览</el-radio-button>
      </el-radio-group>
      <span class="spacer" />
      <el-button type="primary" @click="openCreate()">+ 时间块</el-button>
    </div>

    <div v-if="view === 'day'" class="day-layout" v-loading="loading">
      <el-card class="timeline-card" :body-style="{ padding: 0 }">
        <div class="timeline-scroll" ref="timelineEl" @click="onGridClick" @dragover="onTimelineDragOver" @drop="onTimelineDrop">
          <div class="timeline-inner">
            <!-- 半小时网格 -->
            <div
              v-for="i in halfHours"
              :key="i"
              class="grid-line"
              :class="{ hour: i % 2 === 0 }"
              :style="{ top: `${i * 24}px` }"
            />
            <!-- 小时标签 -->
            <div v-for="h in hourLabels" :key="h" class="hour-label" :style="{ top: `${h * HOUR_H}px` }">
              {{ String(h).padStart(2, '0') }}:00
            </div>
            <!-- 当前时间线 -->
            <div v-if="date === today()" class="now-line" :style="{ top: `${(nowMin() / 60) * HOUR_H}px` }">
              <span class="now-dot" />
              <span class="now-time">{{ fmt(nowMin()) }}</span>
            </div>
            <!-- 时间块 -->
            <div
              v-for="b in dayBlocks"
              :key="b.id"
              class="block"
              :class="{ overlap: overlapIds.has(b.id), dragging: drag.block?.id === b.id, done: b.task_id != null && taskDone.get(b.task_id) }"
              :style="drag.block?.id === b.id ? { top: `${drag.top}px`, height: `${heightOf(b)}px`, background: '#e0e7fe' } : blockStyle(b)"
              @pointerdown="onBlockPointerDown($event, b)"
              @click="!drag.moved && !resize.suppress && openEdit(b)"
            >
              <div v-if="drag.block?.id !== b.id" class="resize-handle top" @pointerdown.stop="onResizePointerDown($event, b, 'top')" />
              <div v-if="drag.block?.id !== b.id" class="resize-handle bottom" @pointerdown.stop="onResizePointerDown($event, b, 'bottom')" />
              <div class="block-title">
                <el-icon v-if="b.task_id != null && taskDone.get(b.task_id)" class="done-icon"><CircleCheckFilled /></el-icon>
                <span class="block-time">{{ fmt(b.start_min) }}–{{ fmt(b.end_min) }}</span>
                <span class="block-name">{{ b.title }}</span>
              </div>
              <el-button class="block-del" link size="small" @click.stop="removeBlock(b)"><el-icon><Close /></el-icon></el-button>
            </div>
          </div>
        </div>
      </el-card>

      <el-card class="task-side">
        <template #header>
          <div class="side-header">
            <span>拖任务到时间轴</span>
            <el-input v-model="taskQuery" placeholder="搜索" size="small" clearable style="width: 110px" />
          </div>
        </template>
        <el-empty v-if="!openTasks.length" description="暂无待办任务" :image-size="50" />
        <div
          v-for="t in openTasks"
          :key="t.id"
          class="task-chip"
          draggable="true"
          @dragstart="onTaskDragStart($event, t.id)"
          @dragend="dragTaskId = null"
        >
          <span class="chip-title">{{ t.title }}</span>
          <el-tag v-if="t.due_date" size="small" type="warning" effect="plain">{{ formatDate(t.due_date) }}</el-tag>
          <el-tag v-if="t.est_minutes" size="small" type="info" effect="plain">{{ t.est_minutes }}min</el-tag>
        </div>
      </el-card>
    </div>

    <el-card v-else v-loading="loading">
      <div class="week-grid">
        <div v-for="col in weekCols" :key="col.date" class="week-col" @click="date = col.date; view = 'day'">
          <div class="week-head" :class="{ today: col.date === today() }">
            <div>{{ col.label }}</div>
            <div class="week-date">{{ col.date.slice(5).replace('-', '/') }}</div>
          </div>
          <div class="week-body">
            <div
              v-for="b in col.blocks"
              :key="b.id"
              class="week-block"
              :class="{ done: b.task_id != null && taskDone.get(b.task_id) }"
              :style="{ background: b.color ? `${b.color}22` : '#eef2ff', borderColor: b.color || '#4c6ef5', top: `${(b.start_min / 1440) * 100}%`, height: `${((b.end_min - b.start_min) / 1440) * 100}%` }"
              :title="`${fmt(b.start_min)}-${fmt(b.end_min)} ${b.title}`"
            >
              <span class="wb-time">{{ fmt(b.start_min) }}</span>
              <span class="wb-title">{{ b.title }}</span>
            </div>
          </div>
        </div>
      </div>
    </el-card>

    <BlockEditor v-model="editorOpen" :block="editing" :block-date="date" :draft-start-min="draftStart" @saved="load" />
  </div>
</template>

<style scoped>
.nav-bar {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-bottom: 16px;
}
.nav-group {
  display: flex;
  gap: 8px;
  align-items: center;
}
.spacer {
  flex: 1;
}

.day-layout {
  display: flex;
  gap: 16px;
  align-items: flex-start;
  flex-wrap: wrap;
}
.timeline-card {
  flex: 1;
  min-width: 0;
}
.timeline-scroll {
  height: calc(100vh - 170px);
  overflow-y: auto;
  border-radius: 12px;
}
.timeline-inner {
  position: relative;
  height: 1152px;
  margin-left: 52px;
}
.grid-line {
  position: absolute;
  left: 0;
  right: 0;
  border-top: 1px solid rgba(148, 163, 184, 0.14);
}
.grid-line.hour {
  border-top-color: rgba(148, 163, 184, 0.3);
}
.hour-label {
  position: absolute;
  left: -52px;
  width: 44px;
  text-align: right;
  font-size: 11px;
  color: #94a3b8;
  transform: translateY(-6px);
}
.now-line {
  position: absolute;
  left: 0;
  right: 0;
  border-top: 2px solid #e03131;
  z-index: 5;
  pointer-events: none;
}
.now-dot {
  position: absolute;
  left: -5px;
  top: -6px;
  width: 10px;
  height: 10px;
  border-radius: 50%;
  background: #e03131;
}
.now-time {
  position: absolute;
  right: 6px;
  top: -20px;
  font-size: 11px;
  color: #e03131;
  background: rgba(255, 255, 255, 0.9);
  padding: 0 4px;
  border-radius: 4px;
}
.block {
  position: absolute;
  left: 8px;
  right: 8px;
  border: 1px solid;
  border-left-width: 4px;
  border-radius: 8px;
  padding: 6px 8px;
  cursor: grab;
  overflow: hidden;
  box-shadow: 0 2px 8px rgba(30, 41, 82, 0.08);
  user-select: none;
  touch-action: none;
}
.block:hover {
  box-shadow: 0 4px 14px rgba(30, 41, 82, 0.16);
}
.block.overlap {
  outline: 2px dashed rgba(224, 49, 49, 0.55);
}
.block.done .block-name {
  text-decoration: line-through;
  color: #2f9e44;
}
.block.dragging {
  opacity: 0.85;
  z-index: 10;
}
.block-title {
  display: flex;
  align-items: center;
  gap: 6px;
  overflow: hidden;
}
.block-time {
  font-size: 11px;
  color: var(--el-text-color-secondary);
  white-space: nowrap;
}
.block-name {
  font-size: 13px;
  font-weight: 600;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.done-icon {
  color: #2f9e44;
}
.resize-handle {
  position: absolute;
  left: 4px;
  right: 4px;
  height: 6px;
  cursor: ns-resize;
  z-index: 3;
  opacity: 0;
}
.block:hover .resize-handle {
  opacity: 1;
}
.resize-handle.top {
  top: -3px;
}
.resize-handle.bottom {
  bottom: -3px;
}
.block-del {
  position: absolute;
  top: 2px;
  right: 2px;
  display: none;
}
.block:hover .resize-handle {
  position: absolute;
  left: 4px;
  right: 4px;
  height: 6px;
  cursor: ns-resize;
  z-index: 3;
  opacity: 0;
}
.block:hover .resize-handle {
  opacity: 1;
}
.resize-handle.top {
  top: -3px;
}
.resize-handle.bottom {
  bottom: -3px;
}
.block-del {
  display: inline-flex;
}

.task-side {
  width: 250px;
  flex-shrink: 0;
  flex-grow: 1;
  min-width: 220px;
}
.side-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-weight: 600;
}
.task-chip {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 7px 8px;
  border: 1px solid rgba(148, 163, 184, 0.3);
  border-radius: 8px;
  margin-bottom: 6px;
  cursor: grab;
  background: rgba(255, 255, 255, 0.7);
}
.task-chip:hover {
  border-color: var(--el-color-primary);
}
.chip-title {
  flex: 1;
  font-size: 13px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.week-grid {
  display: grid;
  grid-template-columns: repeat(7, 1fr);
  gap: 8px;
}
.week-col {
  border: 1px solid rgba(148, 163, 184, 0.22);
  border-radius: 10px;
  overflow: hidden;
  cursor: pointer;
}
.week-col:hover {
  border-color: var(--el-color-primary);
}
.week-head {
  text-align: center;
  padding: 8px 0 6px;
  background: rgba(148, 163, 184, 0.08);
  font-size: 12px;
  color: var(--el-text-color-secondary);
}
.week-head.today {
  background: rgba(76, 110, 245, 0.12);
  color: var(--el-color-primary);
  font-weight: 700;
}
.week-date {
  font-size: 11px;
}
.week-body {
  position: relative;
  height: 480px;
}
.week-block {
  position: absolute;
  left: 3px;
  right: 3px;
  border: 1px solid;
  border-left-width: 3px;
  border-radius: 5px;
  padding: 2px 4px;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  font-size: 10px;
}
.wb-time {
  color: var(--el-text-color-secondary);
}
.wb-title {
  font-weight: 600;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.week-block.done .wb-title {
  text-decoration: line-through;
}
</style>