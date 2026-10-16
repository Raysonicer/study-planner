import dayjs from 'dayjs'

export const DATE_FMT = 'YYYY-MM-DD'

export function today(): string {
  return dayjs().format(DATE_FMT)
}

export function formatDate(d: string): string {
  return dayjs(d).format('M月D日')
}

// 返回 d 所在周的周一（周键=周一）
export function mondayOf(d: string): string {
  const dow = (dayjs(d).day() + 6) % 7
  return dayjs(d).subtract(dow, 'day').format(DATE_FMT)
}

// 按设置项 week_start_day 返回 d 所在周的起始日（默认周一）
export function weekStartOf(d: string): string {
  const startDay = readWeekStartDay()
  const dow = (dayjs(d).day() + 7 - startDay) % 7
  return dayjs(d).subtract(dow, 'day').format(DATE_FMT)
}

// 把任意日期对齐到包含它的那个周的起始日（用于周选择器返回值校正）
export function alignWeekStart(d: string): string {
  return weekStartOf(d)
}

function readWeekStartDay(): number {
  try {
    const raw = localStorage.getItem('sp-week-start-day')
    const n = raw ? parseInt(raw, 10) : 1
    return n >= 0 && n <= 6 ? n : 1
  } catch {
    return 1
  }
}

// 供设置页同步写本地缓存（避免 utils 直接依赖 store 的循环引用）
export function cacheWeekStartDay(day: number) {
  localStorage.setItem('sp-week-start-day', String(day))
}

export function addDays(d: string, n: number): string {
  return dayjs(d).add(n, 'day').format(DATE_FMT)
}

export function weekDays(weekStart: string): string[] {
  return Array.from({ length: 7 }, (_, i) => addDays(weekStart, i))
}

// 周一..周日 的短标签
export const WEEKDAY_LABELS = ['一', '二', '三', '四', '五', '六', '日']

// 以指定起始日生成本周 7 天标签（如周日起始 → 日一二三四五六）
export function weekLabels(startDay = 1): string[] {
  return Array.from({ length: 7 }, (_, i) => WEEKDAY_LABELS[(startDay + i) % 7])
}