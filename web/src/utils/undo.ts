import { h } from 'vue'
import { ElMessage } from 'element-plus'

// 完成/撤销类操作统一提示：5 秒内可撤销（全局交互规范）
export function withUndo(message: string, undoLabel: string, undoFn: () => void | Promise<void>) {
  const actionBtn = h(
    'button',
    {
      class: 'el-message__undo-btn',
      onClick: () => {
        void undoFn()
      },
    },
    undoLabel,
  )
  // element-plus 类型未暴露 action 选项，实际运行时支持
  ElMessage({ message, type: 'success', duration: 5000, showClose: true, action: actionBtn } as never)
}