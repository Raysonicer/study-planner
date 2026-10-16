import { http } from './client'

export interface Health {
  status: string
  version: string
}

export function getHealth() {
  return http.get<Health>('/healthz')
}