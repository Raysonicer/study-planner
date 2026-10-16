import { http } from './client'
import type { Dashboard, Heatmap, StatsDaily, StatsWeekly } from '@/types/api'

export function getDashboard(date: string) {
  return http.get<Dashboard>('/dashboard', { params: { date } })
}

export function getStatsDaily(date: string) {
  return http.get<StatsDaily>('/stats/daily', { params: { date } })
}

export function getStatsWeekly(weekStart: string) {
  return http.get<StatsWeekly>('/stats/weekly', { params: { week_start: weekStart } })
}

export function getStatsHeatmap(start: string, end: string) {
  return http.get<Heatmap>('/stats/heatmap', { params: { start, end } })
}