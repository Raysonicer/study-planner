<script setup lang="ts">
import { reactive, watch } from 'vue'
import { ElMessage } from 'element-plus'
import * as api from '@/api/blocks'
import { useTaskStore } from '@/stores/tasks'
import type { Block } from '@/types/api'

const props = defineProps<{ modelValue: boolean; block?: Block | null; blockDate: string; draftStartMin?: number }>()
const emit = defineEmits<{ (e: 'update:modelValue', v: boolean): void; (e: 'saved'): void }>()

const store = useTaskStore()
const COLORS = [
  { label: '默认', value: '' },
  { label: '靛蓝', value: '#4c6ef5' },
  { label: '橙', value: '#f08c00' },
  { label: '绿', value: '#2f9e44' },
  { label: '红', value: '#e03131' },
  { label: '紫', value: '#7048e8' },
  { label: '青', value: '#0c8599' },
]

const form = reactive({ title: '', start: '09:00', end: '10:00', color: '', task_id: null as number | null, note: '' })

function min2str(m: number): string {
  const h = Math.floor(m / 60)
  const mm = m % 60
  return `${String(h).padStart(2, '0')}:${String(mm).padStart(2, '0')}`
}
function str2min(s: string): number {
  const [h, m] = s.split(':').map(Number)
  return h * 60 + m
}

watch(
  () => props.modelValue,
  (open) => {
    if (!open) return
    const start = props.block?.start_min ?? props.draftStartMin ?? 9 * 60
    const end = props.block?.end_min ?? start + 60
    Object.assign(form, {
      title: props.block?.title ?? '',
      start: min2str(start),
      end: min2str(end),
      color: props.block?.color ?? '',
      task_id: props.block?.task_id ?? null,
      note: props.block?.note ?? '',
    })
  },
)

async function save() {
  if (!form.title.trim()) {
    ElMessage.warning('请输入标题')
    return
  }
  const startMin = str2min(form.start)
  const endMin = str2min(form.end)
  if (endMin <= startMin) {
    ElMessage.warning('结束时间必须晚于开始时间')
    return
  }
  const input = {
    block_date: props.blockDate,
    start_min: startMin,
    end_min: endMin,
    title: form.title.trim(),
    task_id: form.task_id,
    color: form.color,
    note: form.note,
  }
  if (props.block) {
    await api.updateBlock(props.block.id, input)
  } else {
    await api.createBlock(input)
  }
  emit('update:modelValue', false)
  emit('saved')
}
</script>

<template>
  <el-dialog
    :model-value="modelValue"
    :title="block ? '编辑时间块' : '新建时间块'"
    width="480px"
    @update:model-value="emit('update:modelValue', $event)"
  >
    <el-form label-width="80px">
      <el-form-item label="日期">
        <el-text>{{ blockDate }}</el-text>
      </el-form-item>
      <el-form-item label="标题" required>
        <el-input v-model="form.title" placeholder="这段时间做什么？" @keyup.enter="save" />
      </el-form-item>
      <el-form-item label="时间段" required>
        <div class="time-range">
          <el-time-select v-model="form.start" start="00:00" end="23:30" step="00:30" />
          <span class="sep">至</span>
          <el-time-select v-model="form.end" start="00:30" end="24:00" step="00:30" />
        </div>
      </el-form-item>
      <el-form-item label="颜色">
        <div class="colors">
          <div
            v-for="c in COLORS"
            :key="c.value"
            class="swatch"
            :class="{ active: form.color === c.value }"
            :style="{ background: c.value || '#4c6ef5' }"
            :title="c.label"
            @click="form.color = c.value"
          />
        </div>
      </el-form-item>
      <el-form-item label="关联任务">
        <el-select v-model="form.task_id" clearable filterable placeholder="不关联" style="width: 100%">
          <el-option
            v-for="t in store.tasks.filter((x) => x.status !== 'done')"
            :key="t.id"
            :label="t.title"
            :value="t.id"
          />
        </el-select>
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

<style scoped>
.time-range {
  display: flex;
  align-items: center;
  gap: 8px;
}
.sep {
  color: var(--el-text-color-secondary);
}
.colors {
  display: flex;
  gap: 8px;
}
.swatch {
  width: 24px;
  height: 24px;
  border-radius: 50%;
  cursor: pointer;
  border: 2px solid transparent;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.18);
}
.swatch.active {
  border-color: var(--el-color-primary);
  outline: 2px solid rgba(76, 110, 245, 0.25);
}
</style>