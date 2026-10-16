<script setup lang="ts">
import { reactive, watch } from 'vue'
import { ElMessage } from 'element-plus'
import { useGoalStore } from '@/stores/goals'
import type { Goal, GoalInput, GoalMode, GoalType } from '@/types/api'

const props = defineProps<{
  modelValue: boolean
  goal?: Goal | null
  goalType?: GoalType
  date?: string
  weekStart?: string
}>()
const emit = defineEmits<{ (e: 'update:modelValue', v: boolean): void; (e: 'saved'): void }>()

const store = useGoalStore()
const form = reactive({
  title: '',
  goal_type: 'day' as GoalType,
  goal_mode: 'daily' as GoalMode,
  date: null as string | null,
  week_start: null as string | null,
  checkin_days: 5,
  note: '',
})

watch(
  () => props.modelValue,
  (open) => {
    if (!open) return
    Object.assign(form, {
      title: props.goal?.title ?? '',
      goal_type: props.goal?.goal_type ?? props.goalType ?? 'day',
      goal_mode: props.goal?.goal_mode ?? 'daily',
      date: props.goal?.date ?? props.date ?? null,
      week_start: props.goal?.week_start ?? props.weekStart ?? null,
      checkin_days: props.goal?.checkin_days ?? 5,
      note: props.goal?.note ?? '',
    })
  },
)

async function save() {
  if (!form.title.trim()) {
    ElMessage.warning('请输入目标标题')
    return
  }
  const input: GoalInput = {
    title: form.title.trim(),
    goal_type: form.goal_type,
    goal_mode: form.goal_type === 'week' ? form.goal_mode : 'daily',
    date: form.goal_type === 'day' ? form.date : null,
    week_start: form.goal_type === 'week' ? form.week_start : null,
    checkin_days: form.goal_mode === 'daily' ? form.checkin_days : undefined,
    note: form.note,
  }
  if (!input.date && form.goal_type === 'day') {
    ElMessage.warning('请选择日期')
    return
  }
  if (!input.week_start && form.goal_type === 'week') {
    ElMessage.warning('请选择周')
    return
  }
  if (props.goal) {
    await store.updateGoal(props.goal.id, input)
  } else {
    await store.createGoal(input)
  }
  emit('update:modelValue', false)
  emit('saved')
}
</script>

<template>
  <el-dialog
    :model-value="modelValue"
    :title="goal ? '编辑目标' : '新建目标'"
    width="500px"
    @update:model-value="emit('update:modelValue', $event)"
  >
    <el-form label-width="80px">
      <el-form-item label="类型">
        <el-radio-group v-model="form.goal_type">
          <el-radio-button value="day">日目标</el-radio-button>
          <el-radio-button value="week">周目标</el-radio-button>
        </el-radio-group>
      </el-form-item>
      <el-form-item label="标题" required>
        <el-input v-model="form.title" placeholder="目标标题" @keyup.enter="save" />
      </el-form-item>
      <el-form-item v-if="form.goal_type === 'day'" label="日期" required>
        <el-date-picker v-model="form.date" type="date" value-format="YYYY-MM-DD" style="width: 100%" />
      </el-form-item>
      <template v-else>
        <el-form-item label="所属周" required>
          <el-date-picker v-model="form.week_start" type="week" value-format="YYYY-MM-DD" style="width: 100%" />
        </el-form-item>
        <el-form-item label="达成方式">
          <el-radio-group v-model="form.goal_mode">
            <el-radio-button value="daily">每日打卡</el-radio-button>
            <el-radio-button value="once">一次性</el-radio-button>
          </el-radio-group>
        </el-form-item>
        <el-form-item v-if="form.goal_mode === 'daily'" label="应打卡">
          <el-input-number v-model="form.checkin_days" :min="1" :max="7" /> 天 / 周
        </el-form-item>
      </template>
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