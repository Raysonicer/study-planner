import { db, type SettingRow } from '@/db/db'

// 与 Go 版 store 包的 settings 读写对齐
export async function allSettings(): Promise<Record<string, string>> {
  const rows = await db.settings.toArray()
  const m: Record<string, string> = {}
  for (const r of rows) m[r.key] = r.value
  return m
}

export async function setSetting(key: string, value: string): Promise<void> {
  await db.settings.put({ key, value } satisfies SettingRow)
}

export async function getSettingString(key: string, def = ''): Promise<string> {
  const r = await db.settings.get(key)
  return r?.value ?? def
}

export async function getSettingInt(key: string, def: number): Promise<number> {
  const raw = await getSettingString(key, String(def))
  const n = parseInt(raw, 10)
  return Number.isFinite(n) ? n : def
}

export async function getSettingFloat(key: string, def: number): Promise<number> {
  const raw = await getSettingString(key, String(def))
  const n = parseFloat(raw)
  return Number.isFinite(n) ? n : def
}

export async function getWeekStartDay(): Promise<number> {
  const n = await getSettingInt('week_start_day', 1)
  return n >= 0 && n <= 6 ? n : 1
}
