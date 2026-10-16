<script setup lang="ts">
import { reactive, ref, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { useTaskStore } from '@/stores/tasks'
import { listGoals } from '@/api/goals'
import type { Goal } from '@/types/api'

const props = defineProps<{ modelValue: boolean; parentId?: number | null }>()
const emit = defineEmits<{ (e: 'update:modelValue', v: boolean): void; (e: 'saved'): void }>()

const store = useTaskStore()
const goals = ref<Goal[]>([])
const form = reactive({ parent_id: null as number | null, goal_id: null as number | null, text: '' })

watch(
  () => props.modelValue,
  async (open) => {
    if (!open) return
    goals.value = (await listGoals({ status: 'active' })).data
    form.parent_id = props.parentId ?? null
    form.goal_id = null
    form.text = ''
  },
)

function buildTree(parentId: number | null): { value: number; label: string; children: unknown[] }[] {
  return (store.childrenMap.get(parentId) ?? []).map((t) => ({
    value: t.id,
    label: t.title,
    children: buildTree(t.id),
  }))
}

async function submit() {
  const lines = form.text.split('\n')
  const valid = lines.filter((l) => l.trim())
  if (!valid.length) {
    ElMessage.warning('请输入至少一行任务')
    return
  }
  await store.batch(form.parent_id, form.goal_id, valid)
  emit('update:modelValue', false)
  emit('saved')
}
</script>

<template>
  <el-dialog :model-value="modelValue" title="批量分解子任务" width="520px" @update:model-value="emit('update:modelValue', $event)">
    <el-form label-width="80px">
      <el-form-item label="父任务">
        <el-tree-select v-model="form.parent_id" :data="buildTree(null)" check-strictly clearable placeholder="无（顶层任务）" style="width: 100%" />
      </el-form-item>
      <el-form-item label="关联目标">
        <el-select v-model="form.goal_id" clearable placeholder="不关联" style="width: 100%">
          <el-option v-for="g in goals" :key="g.id" :label="g.title" :value="g.id" />
        </el-select>
      </el-form-item>
      <el-form-item label="任务列表" required>
        <el-input v-model="form.text" type="textarea" :rows="8" placeholder="每行一个子任务，空行自动忽略" />
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="emit('update:modelValue', false)">取消</el-button>
      <el-button type="primary" @click="submit">生成子任务</el-button>
    </template>
  </el-dialog>
</template>