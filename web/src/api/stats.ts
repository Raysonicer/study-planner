import { local } from './client'
import * as svc from '@/local/stats'
import type { Dashboard, Heatmap, StatsDaily, StatsWeekly } from '@/types/api'

export function getDashboard(date: string) {
  return local<Dashboard>(() => svc.dashboard(date))
}

export function getStatsDaily(date: string) {
  return local<StatsDaily>(() => svc.dailyStats(date))
}

export function getStatsWeekly(weekStart: string) {
  return local<StatsWeekly>(() => svc.weeklyStats(weekStart))
}

export function getStatsHeatmap(start: string, end: string) {
  return local<Heatmap>(() => svc.heatmap(start, end))
}
