import * as echarts from 'echarts/core'
import { BarChart, HeatmapChart, LineChart } from 'echarts/charts'
import { CalendarComponent, GridComponent, LegendComponent, TooltipComponent, VisualMapComponent } from 'echarts/components'
import { CanvasRenderer } from 'echarts/renderers'
import type { ComposeOption } from 'echarts/core'
import type { BarSeriesOption, HeatmapSeriesOption, LineSeriesOption } from 'echarts/charts'
import type { CalendarComponentOption, GridComponentOption, LegendComponentOption, TooltipComponentOption, VisualMapComponentOption } from 'echarts/components'
import type { Ref } from 'vue'
import { onUnmounted } from 'vue'

echarts.use([LineChart, BarChart, HeatmapChart, CalendarComponent, GridComponent, LegendComponent, TooltipComponent, VisualMapComponent, CanvasRenderer])

export type ChartOption = ComposeOption<
  | LineSeriesOption
  | BarSeriesOption
  | HeatmapSeriesOption
  | GridComponentOption
  | LegendComponentOption
  | TooltipComponentOption
  | VisualMapComponentOption
  | CalendarComponentOption
>

// ECharts 生命周期封装：init/resize/dispose
export function useChart(el: Ref<HTMLElement | null>) {
  let chart: echarts.ECharts | null = null
  let observer: ResizeObserver | null = null

  function ensure(): echarts.ECharts | null {
    if (!el.value) return null
    if (!chart) {
      chart = echarts.init(el.value)
      observer = new ResizeObserver(() => chart?.resize())
      observer.observe(el.value)
    }
    return chart
  }

  function setOption(option: ChartOption) {
    ensure()?.setOption(option, true)
  }

  function on(event: string, handler: (...args: any[]) => void) {
    ensure()?.on(event, handler as never)
  }

  onUnmounted(() => {
    observer?.disconnect()
    chart?.dispose()
    chart = null
  })

  return { setOption, on }
}