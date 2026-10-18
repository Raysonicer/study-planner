import { local } from './client'

export interface Health {
  status: string
  version: string
}

export function getHealth() {
  return local<Health>(async () => ({ status: 'ok', version: '0.2.0-local' }))
}
