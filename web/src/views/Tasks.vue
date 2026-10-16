<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useTaskStore } from '@/stores/tasks'
import { listGoals } from '@/api/goals'
import TaskNode from '@/components/TaskNode.vue'
import TaskEditor from '@/components/TaskEditor.vue'
import BatchTasksDialog from '@/components/BatchTasksDialog.vue'
import type { Goal, Task } from '@/types/api'

const store = useTaskStore()
const goals = ref<Goal[]>([])

const q = ref('')
const status = ref('')
const goalFilter = ref<number | null>(null)
const tag = ref('')

const editorOpen = ref(false)
const editing = ref<Task | null>(null)
const parentForNew = ref<number | null>(null)
const batchOpen = ref(false)
const expandDepth = ref(2)
const treeKey = ref(0)

onMounted(async () => {
  await store.fetchAll()
  goals.value = (await listGoals({ status: 'active' })).data
})

// 可见集合：自身匹配或存在匹配后代的节点均保留（树语义过滤）
const visibleSet = computed(() => {
  const set = new Set<number>()
  const match = (t: Task) =>
    (!status.value || t.status === status.value) &&
    (goalFilter.value === null || t.goal_id === goalFilter.value) &&
    (!tag.value || t.tags.includes(tag.value)) &&
    (!q.value || t.title.includes(q.value))
  const walk = (id: number): boolean => {
    const t = store.byId.get(id)!
    let ok = match(t)
    for (const c of store.childrenMap.get(id) ?? []) {
      if (walk(c.id)) ok = true
    }
    if (ok) set.add(id)
    return ok
  }
  for (const r of store.roots) walk(r.id)
  return set
})

const filteredRoots = computed(() => store.roots.filter((r) => visibleSet.value.has(r.id)))

function toggleExpandAll() {
  expandDepth.value = expandDepth.value === 2 ? 99 : 2
  treeKey.value++
}

function openCreate(parentId: number | null = null) {
  editing.value = null
  parentForNew.value = parentId
  editorOpen.value = true
}

function onEdit(task: Task) {
  editing.value = task
  parentForNew.value = null
  editorOpen.value = true
}
</script>

<template>
  <div class="page">
    <el-card class="toolbar">
      <el-input v-model="q" placeholder="搜索标题" clearable style="width: 200px" />
      <el-select v-model="status" placeholder="状态" clearable style="width: 120px">
        <el-option label="待办" value="todo" />
        <el-option label="进行中" value="doing" />
        <el-option label="已完成" value="done" />
      </el-select>
      <el-select v-model="goalFilter" placeholder="按目标" clearable style="width: 180px">
        <el-option v-for="g in goals" :key="g.id" :label="g.title" :value="g.id" />
      </el-select>
      <el-input v-model="tag" placeholder="按标签" clearable style="width: 140px" />
      <span class="spacer" />
      <el-button @click="toggleExpandAll">{{ expandDepth === 2 ? '全部展开' : '默认折叠' }}</el-button>
      <el-button @click="batchOpen = true">批量分解</el-button>
      <el-button type="primary" @click="openCreate()">+ 新建任务</el-button>
    </el-card>

    <el-card class="tree-card" v-loading="store.loading">
      <el-empty v-if="!store.loading && !store.tasks.length" description="还没有任务，点击右上角新建" />
      <template v-else>
        <TaskNode
          v-for="r in filteredRoots"
          :key="`${treeKey}-${r.id}`"
          :task="r"
          :siblings="filteredRoots"
          :depth="0"
          :default-expand-depth="expandDepth"
          :visible-ids="visibleSet"
          @edit="onEdit"
          @add-child="(t) => openCreate(t.id)"
        />
        <el-empty v-if="!store.loading && !filteredRoots.length" description="没有符合条件的任务" :image-size="60" />
      </template>
    </el-card>

    <TaskEditor v-model="editorOpen" :task="editing" :parent-id="parentForNew" />
    <BatchTasksDialog v-model="batchOpen" />
  </div>
</template>

<style scoped>
.toolbar {
  display: flex;
  gap: 8px;
  align-items: center;
  margin-bottom: 16px;
}
.toolbar :deep(.el-card__body) {
  display: flex;
  gap: 8px;
  align-items: center;
  flex-wrap: wrap;
}
.spacer {
  flex: 1;
}
.tree-card {
  min-height: 300px;
}
</style>