import { local } from './client'
import * as svc from '@/local/settings'

export function getSettings() {
  return local<Record<string, string>>(() => svc.allSettings())
}

export function putSettings(values: Record<string, string>) {
  return local<Record<string, string>>(async () => {
    for (const [k, v] of Object.entries(values)) {
      await svc.setSetting(k, v)
    }
    return values
  })
}
