<script setup lang="ts">
import { reactive, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { useTaskStore } from '@/stores/tasks'
import { listGoals } from '@/api/goals'
import type { Goal, Task, TaskInput } from '@/types/api'

const props = defineProps<{ modelValue: boolean; task?: Task | null; parentId?: number | null }>()
const emit = defineEmits<{ (e: 'update:modelValue', v: boolean): void; (e: 'saved'): void }>()

const store = useTaskStore()
const goals = ref<Goal[]>([])

const form = reactive({
  title: '',
  note: '',
  priority: 1,
  due_date: null as string | null,
  est_minutes: null as number | null,
  tags: [] as string[],
  parent_id: null as number | null,
  goal_id: null as number | null,
})

watch(
  () => props.modelValue,
  async (open) => {
    if (!open) return
    const g = (await listGoals({ status: 'active' })).data
    goals.value = g.filter((x) => x.status === 'active')
    Object.assign(form, {
      title: props.task?.title ?? '',
      note: props.task?.note ?? '',
      priority: props.task?.priority ?? 1,
      due_date: props.task?.due_date ?? null,
      est_minutes: props.task?.est_minutes ?? null,
      tags: props.task?.tags ?? [],
      parent_id: props.task?.parent_id ?? props.parentId ?? null,
      goal_id: props.task?.goal_id ?? null,
    })
  },
)

// 父任务候选：排除自身与后代（防环在前后端双重保证）
function buildTree(parentId: number | null): { value: number; label: string; children: unknown[] }[] {
  const list = store.childrenMap.get(parentId) ?? []
  const excluded = new Set<number>()
  if (props.task) {
    const walk = (id: number) => {
      for (const t of store.tasks) {
        if (t.parent_id === id) {
          excluded.add(t.id)
          walk(t.id)
        }
      }
    }
    walk(props.task.id)
  }
  return list
    .filter((t) => t.id !== props.task?.id && !excluded.has(t.id))
    .map((t) => ({ value: t.id, label: t.title, children: buildTree(t.id) }))
}

async function save() {
  if (!form.title.trim()) {
    ElMessage.warning('请输入标题')
    return
  }
  const input: TaskInput = {
    title: form.title.trim(),
    note: form.note,
    priority: form.priority,
    due_date: form.due_date || null,
    est_minutes: form.est_minutes,
    tags: form.tags,
    parent_id: form.parent_id,
    goal_id: form.goal_id,
  }
  if (props.task) {
    await store.update(props.task.id, input)
  } else {
    await store.create(input)
  }
  emit('update:modelValue', false)
  emit('saved')
}
</script>

<template>
  <el-dialog
    :model-value="modelValue"
    :title="task ? '编辑任务' : '新建任务'"
    width="520px"
    @update:model-value="emit('update:modelValue', $event)"
  >
    <el-form label-width="80px">
      <el-form-item label="标题" required>
        <el-input v-model="form.title" placeholder="任务标题" @keyup.enter="save" />
      </el-form-item>
      <el-form-item label="父任务">
        <el-tree-select
          v-model="form.parent_id"
          :data="buildTree(null)"
          check-strictly
          clearable
          placeholder="无（顶层任务）"
          style="width: 100%"
        />
      </el-form-item>
      <el-form-item label="关联目标">
        <el-select v-model="form.goal_id" clearable placeholder="不关联" style="width: 100%">
          <el-option v-for="g in goals" :key="g.id" :label="g.title" :value="g.id" />
        </el-select>
      </el-form-item>
      <el-form-item label="优先级">
        <el-radio-group v-model="form.priority">
          <el-radio-button :value="0">P0</el-radio-button>
          <el-radio-button :value="1">P1</el-radio-button>
          <el-radio-button :value="2">P2</el-radio-button>
          <el-radio-button :value="3">P3</el-radio-button>
        </el-radio-group>
      </el-form-item>
      <el-form-item label="截止日">
        <el-date-picker v-model="form.due_date" type="date" value-format="YYYY-MM-DD" placeholder="选择日期" style="width: 100%" />
      </el-form-item>
      <el-form-item label="预估时长">
        <el-input-number v-model="form.est_minutes" :min="1" :step="15" placeholder="分钟" />
      </el-form-item>
      <el-form-item label="标签">
        <el-select v-model="form.tags" multiple filterable allow-create default-first-option placeholder="输入后回车创建" style="width: 100%" />
      </el-form-item>
      <el-form-item label="备注">
        <el-input v-model="form.note" type="textarea" :rows="2" />
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="emit('update:modelValue', false)">取消</el-button>
      <el-button type="primary" @click="save">保存</el-button>
    </template>
  </el-dialog>
</template>