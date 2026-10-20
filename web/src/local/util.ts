// 深拷贝为纯 JSON 对象：Vue reactive/ref 产生的 Proxy 无法被 IndexedDB 结构化克隆，
// 所有写入 IndexedDB 的入参都必须先过 plain()。
export function plain<T>(v: T): T {
  return JSON.parse(JSON.stringify(v)) as T
}
