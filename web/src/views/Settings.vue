<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useSettingsStore } from '@/stores/settings'
import { http } from '@/api/client'
import { today, cacheWeekStartDay } from '@/utils/date'

const settings = useSettingsStore()
const fileInput = ref<HTMLInputElement | null>(null)
onMounted(() => settings.load())

async function set(key: string, value: string, tip?: string) {
  await settings.set(key, value)
  if (tip) ElMessage.success(tip)
}

async function setWeekStart(v: string | number | boolean) {
  const day = Number(v)
  cacheWeekStartDay(day)
  await settings.set('week_start_day', String(day))
  ElMessage.success('周起始日已更新，切换周视图生效')
}

function downloadExport() {
  void (async () => {
    const res = await http.get('/export', { responseType: 'blob' })
    const url = URL.createObjectURL(res.data as Blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `studypilot-export-${today()}.json`
    a.click()
    URL.revokeObjectURL(url)
    ElMessage.success('导出完成')
  })()
}

function onFilePicked() {
  const input = fileInput.value
  const file = input?.files?.[0]
  if (!file) return
  const reader = new FileReader()
  reader.onload = () => {
    void (async () => {
      try {
        const data = JSON.parse(String(reader.result))
        await ElMessageBox.confirm('导入将覆盖当前全部数据（建议先导出备份）', '确认导入', { type: 'warning', confirmButtonText: '覆盖导入' })
        await http.post('/export/import', data)
        ElMessage.success('导入完成，数据已替换')
      } catch (e) {
        ElMessage.error('导入失败：文件格式不正确')
      }
    })()
  }
  reader.readAsText(file)
  input.value = ''
}
</script>

<template>
  <div class="page">
    <el-card class="section">
      <template #header>外观</template>
      <el-form label-width="120px">
        <el-form-item label="主题">
          <el-radio-group :model-value="settings.theme" @change="(v: string | number | boolean) => set('theme', String(v), '主题已切换')">
            <el-radio-button value="light">浅色（渐变）</el-radio-button>
            <el-radio-button value="dark">深色（渐变）</el-radio-button>
          </el-radio-group>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card class="section">
      <template #header>统计</template>
      <el-form label-width="120px">
        <el-form-item label="热力图 · 任务权重">
          <el-slider :model-value="settings.heatmapTaskWeight" :min="0" :max="1" :step="0.1" style="max-width: 320px" @change="(v: number) => set('heatmap_task_weight', String(Number(v).toFixed(1)), '任务权重已更新')" />
          <span class="val">{{ settings.heatmapTaskWeight.toFixed(1) }}</span>
        </el-form-item>
        <el-form-item label="热力图 · 目标权重">
          <el-slider :model-value="settings.heatmapGoalWeight" :min="0" :max="1" :step="0.1" style="max-width: 320px" @change="(v: number) => set('heatmap_goal_weight', String(Number(v).toFixed(1)), '目标权重已更新')" />
          <span class="val">{{ settings.heatmapGoalWeight.toFixed(1) }}</span>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card class="section">
      <template #header>任务与提醒</template>
      <el-form label-width="120px">
        <el-form-item label="子任务全完成时自动完成父任务">
          <el-switch :model-value="settings.autoCompleteParent" @change="(v: boolean) => set('auto_complete_parent', String(Boolean(v)), '已' + (v ? '开启' : '关闭'))" />
        </el-form-item>
        <el-form-item label="逾期桌面通知">
          <el-switch :model-value="settings.notifyEnabled" @change="(v: boolean) => set('notify_overdue', String(Boolean(v)), '已' + (v ? '开启' : '关闭'))" />
          <el-text size="small" type="info">打开今日面板时若有逾期任务将弹出系统通知</el-text>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card class="section">
      <template #header>周设置</template>
      <el-form label-width="120px">
        <el-form-item label="周起始日">
          <el-select :model-value="String(settings.weekStartDay)" style="width: 160px" @change="setWeekStart">
            <el-option label="周一（默认）" value="1" />
            <el-option label="周日" value="0" />
            <el-option label="周六" value="6" />
          </el-select>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card class="section">
      <template #header>数据</template>
      <el-form label-width="120px">
        <el-form-item label="存储位置">
          <el-text>data/app.db（SQLite 单文件，启动时自动备份 14 份到 data/backups）</el-text>
        </el-form-item>
        <el-form-item label="备份与迁移">
          <el-button @click="downloadExport"><el-icon><Download /></el-icon>导出 JSON</el-button>
          <el-button @click="fileInput?.click()"><el-icon><Upload /></el-icon>导入 JSON（覆盖）</el-button>
          <input ref="fileInput" type="file" accept=".json,application/json" style="display: none" @change="onFilePicked" />
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<style scoped>
.page {
  max-width: 720px;
}
.section {
  margin-bottom: 16px;
}
.val {
  margin-left: 12px;
  font-weight: 700;
}
</style>