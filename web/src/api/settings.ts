import { http } from './client'

export function getSettings() {
  return http.get<Record<string, string>>('/settings')
}

export function putSettings(values: Record<string, string>) {
  return http.put<Record<string, string>>('/settings', values)
}