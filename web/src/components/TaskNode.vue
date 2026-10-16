<script setup lang="ts">
import { computed, ref } from 'vue'
import { useTaskStore } from '@/stores/tasks'
import type { Task } from '@/types/api'
import { formatDate } from '@/utils/date'

defineOptions({ name: 'TaskNode' })

const props = defineProps<{
  task: Task
  siblings: Task[]
  depth: number
  defaultExpandDepth?: number
  visibleIds?: Set<number>
}>()
const emit = defineEmits<{ (e: 'edit', task: Task): void; (e: 'add-child', task: Task): void }>()

const store = useTaskStore()
const expanded = ref(props.depth < (props.defaultExpandDepth ?? 2))

const children = computed(() =>
  (store.childrenMap.get(props.task.id) ?? []).filter((t) => !props.visibleIds || props.visibleIds.has(t.id)),
)
const progress = computed(() => store.progress(props.task.id))
const overdue = computed(() => store.isOverdue(props.task))
const allDone = computed(
  () => progress.value.total > 0 && progress.value.done === progress.value.total && props.task.status !== 'done',
)

// 同层级拖拽排序（原生 HTML5 DnD）
let dragId: number | null = null
function onDragStart(e: DragEvent) {
  dragId = props.task.id
  if (e.dataTransfer) e.dataTransfer.effectAllowed = 'move'
}
function onDragOver(e: DragEvent) {
  const src = dragId !== null ? store.byId.get(dragId) : undefined
  if (src && src.parent_id === props.task.parent_id && dragId !== props.task.id) {
    e.preventDefault()
  }
}
function onDrop() {
  if (dragId === null) return
  const from = props.siblings.findIndex((t) => t.id === dragId)
  const to = props.siblings.findIndex((t) => t.id === props.task.id)
  if (from < 0 || to < 0 || from === to) return
  const ordered = [...props.siblings]
  const [moved] = ordered.splice(from, 1)
  ordered.splice(to, 0, moved)
  void store.reorder(props.task.parent_id, ordered.map((t) => t.id))
  dragId = null
}
</script>

<template>
  <div>
    <div
      class="task-row"
      :class="{ 'is-done': task.status === 'done' }"
      :style="{ paddingLeft: `${depth * 22}px` }"
      draggable="true"
      @dragstart="onDragStart"
      @dragover="onDragOver"
      @drop="onDrop"
    >
      <el-icon v-if="children.length" class="arrow" @click="expanded = !expanded">
        <ArrowDown v-if="expanded" />
        <ArrowRight v-else />
      </el-icon>
      <span v-else class="arrow" />
      <el-checkbox :model-value="task.status === 'done'" @change="store.toggleDone(task)" />
      <span class="title" :title="task.note" @click="expanded = !expanded">{{ task.title }}</span>
      <el-tag v-if="progress.total" size="small" type="info" effect="plain">
        {{ progress.done }}/{{ progress.total }}
      </el-tag>
      <el-tag v-if="allDone" size="small" type="success" class="clickable" @click="store.toggleDone(task)">
        一键完成
      </el-tag>
      <el-tag v-if="overdue" size="small" type="danger">逾期</el-tag>
      <el-tag v-else-if="task.due_date" size="small" type="warning" effect="plain">
        {{ formatDate(task.due_date) }}
      </el-tag>
      <el-tag v-if="task.priority === 0" size="small" type="danger" effect="dark">P0</el-tag>
      <el-tag v-for="tag in task.tags" :key="tag" size="small" effect="plain">{{ tag }}</el-tag>
      <span class="actions">
        <el-button link size="small" type="primary" @click="emit('add-child', task)">子任务</el-button>
        <el-button link size="small" @click="emit('edit', task)">编辑</el-button>
        <el-button link size="small" type="danger" @click="store.remove(task.id)">删除</el-button>
      </span>
    </div>
    <el-progress
      v-if="expanded && progress.total"
      class="mini-progress"
      :percentage="Math.round((progress.done / progress.total) * 100)"
      :show-text="false"
      :stroke-width="3"
    />
    <template v-if="expanded">
      <TaskNode v-for="c in children" :key="c.id" :task="c" :siblings="children" :depth="depth + 1" @edit="emit('edit', $event)" @add-child="emit('add-child', $event)" />
    </template>
  </div>
</template>

<style scoped>
.task-row {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 8px;
  border-radius: 4px;
  cursor: default;
}
.task-row:hover {
  background: var(--el-fill-color-light);
}
.task-row.is-done .title {
  text-decoration: line-through;
  color: var(--el-text-color-secondary);
}
.arrow {
  width: 16px;
  cursor: pointer;
  color: var(--el-text-color-secondary);
}
.title {
  flex: 1;
  cursor: pointer;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.actions {
  display: none;
  gap: 4px;
}
.task-row:hover .actions {
  display: inline-flex;
}
.mini-progress {
  margin: 2px 0 2px 60px;
  max-width: 240px;
}
.clickable {
  cursor: pointer;
}
</style>