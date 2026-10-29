// 前端数据模型

export type Scheme = 'http' | 'https'

export interface Target {
  scheme: Scheme // 默认 'http'
  host: string
  port: number // 0 = 协议默认
  useProxy: boolean // 每标签独立
  proxyURL: string // 每标签独立，默认 'http://127.0.0.1:7890'
}

export interface ResponseResult {
  statusLine: string // 'HTTP/1.1 200 OK'
  statusCode: number // 0 = 传输层错误
  headersRaw: string // 原始头文本（保序）
  bodyRaw: string // base64（二进制安全）
  bodyDecoded: string | null // gzip/br 解码后文本（可为 null）
  bytes: number
  durationMs: number
  error: string // '' 表示成功
}

export interface HistoryItem {
  id: string
  seq: number // 标签内自增
  requestText: string // 发送时请求快照（LF 展示格式）
  requestSentText?: string // 实际写出的报文（Host / Content-Length 重写后）
  target: Target
  response: ResponseResult | null
  sentAt: number // epoch ms
}

export interface RepeaterTab {
  id: string
  title: string // '请求 3'
  requestText: string // 当前编辑器内容
  target: Target
  followHostHeader: boolean // Host 头跟随目标，默认 true
  autoContentLength: boolean // 默认 true
  histories: HistoryItem[]
  activeHistoryIndex: number // -1 = 最新编辑态
  updatedAt: number
}

// ---- Go 绑定交互类型（对应 internal/repeater 的 JSON 结构）----

export interface ParsedTarget {
  scheme: Scheme
  host: string
  port: number
}

export interface ParsedRequest {
  method: string
  path: string
  version: string
  headers: [string, string][]
  body: string
  target: ParsedTarget
  isBareUrl: boolean
  skeleton: string
}

export interface SendRequest {
  requestText: string
  scheme: string
  host: string
  port: number
  useProxy: boolean
  proxyURL: string
  followHostHeader: boolean
  autoContentLength: boolean
  verifyTLS: boolean
  timeoutMs: number
  lang: 'zh' | 'en'
}

export interface SendResult {
  requestSent: string
  response: ResponseResult | null
}

// curl 命令 / fetch 代码导入结果
export interface ImportedRequest {
  source: 'curl' | 'fetch'
  requestText: string
  target: ParsedTarget
  hostHeaderMismatch: boolean
  warnings: string[] | null
}
