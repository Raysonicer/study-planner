import dayjs from 'dayjs'

// 与 Go 版 service.go 的日期工具对齐（本地时区、YYYY-MM-DD）
export const DATE_FMT = 'YYYY-MM-DD'
export const TIME_FMT = 'YYYY-MM-DD HH:mm:ss'

export function today(): string {
  return dayjs().format(DATE_FMT)
}

export function nowString(): string {
  return dayjs().format(TIME_FMT)
}

export function validDate(s: string): boolean {
  if (!/^\d{4}-\d{2}-\d{2}$/.test(s)) return false
  return dayjs(s, DATE_FMT, true).isValid()
}

export function addDays(d: string, n: number): string {
  return dayjs(d).add(n, 'day').format(DATE_FMT)
}

export function weekEndDate(weekStart: string): string {
  return addDays(weekStart, 6)
}

export function weekdayOf(d: string): number {
  // 0=周日 … 6=周六（与 Go time.Weekday 对齐）
  return dayjs(d).day()
}

// 周起始日必须是设置项 week_start_day（默认周一=1）
export function isWeekStart(v: string, weekStartDay: number): boolean {
  if (!validDate(v)) return false
  const day = weekStartDay >= 0 && weekStartDay <= 6 ? weekStartDay : 1
  return weekdayOf(v) === day
}

