import {
  ImportRequest as ImportRequestBinding,
  ParseRaw as ParseRawBinding,
  Send as SendBinding,
} from '../../wailsjs/go/repeater/Service'
import type { ImportedRequest, ParsedRequest, SendRequest, SendResult } from '../types'
import type { Lang } from '../i18n'

// 单一边界：隔离 Wails 生成的绑定类型与本地模型
export function parseRaw(raw: string, lang: Lang): Promise<ParsedRequest> {
  return ParseRawBinding(raw, lang) as unknown as Promise<ParsedRequest>
}

// curl / fetch → 可编辑 HTTP 报文；不是这两种格式时 reject，调用方回退 parseRaw
export function importRequest(text: string, lang: Lang): Promise<ImportedRequest> {
  return ImportRequestBinding(text, lang) as unknown as Promise<ImportedRequest>
}

export function sendRequest(req: SendRequest): Promise<SendResult> {
  const payload = req as unknown as Parameters<typeof SendBinding>[0]
  return SendBinding(payload) as unknown as Promise<SendResult>
}
