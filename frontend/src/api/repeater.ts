import {
  BuildRequestCurl as BuildRequestCurlBinding,
  BuildRequestURL as BuildRequestURLBinding,
  CancelLLMChat as CancelLLMChatBinding,
  CancelSend as CancelSendBinding,
  ConvertRequestMethod as ConvertRequestMethodBinding,
  ImportRequest as ImportRequestBinding,
  LLMChat as LLMChatBinding,
  LoadSettings as LoadSettingsBinding,
  ParseRaw as ParseRawBinding,
  SaveSettings as SaveSettingsBinding,
  Send as SendBinding,
} from '../../wailsjs/go/repeater/Service'
import type {
  AppSettings,
  ChatRequest,
  ChatResponse,
  ImportedRequest,
  ParsedRequest,
  RequestToolInput,
  RequestToolResult,
  SendRequest,
  SendResult,
  SettingsLoadResult,
} from '../types'
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

// 取消正在发送的 HTTP 请求（Go 侧关闭连接，结果标记为「已取消」）
export function cancelSend(): Promise<void> {
  return CancelSendBinding()
}

// ---- 设置（~/.httpRepeater/setting.json）----

export function loadSettings(): Promise<SettingsLoadResult> {
  return LoadSettingsBinding() as unknown as Promise<SettingsLoadResult>
}

export function saveSettings(settings: AppSettings): Promise<void> {
  const payload = settings as unknown as Parameters<typeof SaveSettingsBinding>[0]
  return SaveSettingsBinding(payload)
}

// ---- AI 对话 ----

export function llmChat(req: ChatRequest): Promise<ChatResponse> {
  const payload = req as unknown as Parameters<typeof LLMChatBinding>[0]
  return LLMChatBinding(payload) as unknown as Promise<ChatResponse>
}

// 取消正在进行的 LLM 请求（Go 侧会中断 HTTP 上下文）
export function cancelLLMChat(): Promise<void> {
  return CancelLLMChatBinding()
}

// ---- 右键菜单工具 ----

export function buildRequestURL(req: RequestToolInput): Promise<RequestToolResult> {
  const payload = req as unknown as Parameters<typeof BuildRequestURLBinding>[0]
  return BuildRequestURLBinding(payload) as unknown as Promise<RequestToolResult>
}

export function buildRequestCurl(req: RequestToolInput): Promise<RequestToolResult> {
  const payload = req as unknown as Parameters<typeof BuildRequestCurlBinding>[0]
  return BuildRequestCurlBinding(payload) as unknown as Promise<RequestToolResult>
}

export function convertRequestMethod(req: RequestToolInput): Promise<RequestToolResult> {
  const payload = req as unknown as Parameters<typeof ConvertRequestMethodBinding>[0]
  return ConvertRequestMethodBinding(payload) as unknown as Promise<RequestToolResult>
}
