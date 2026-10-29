import { ref } from 'vue'

export type Lang = 'zh' | 'en'

const LANG_STORAGE_KEY = 'httpRepeater.lang'

const zh = {
  'app.subtitle': 'HTTP 请求重放工具',
  'app.switchToLight': '切换到浅色主题',
  'app.switchToDark': '切换到深色主题',
  'app.switchLang': 'Switch to English',
  'app.byAuthor': '作者 {name}',
  'app.settings': '设置',

  'tab.defaultTitle': '请求 {n}',
  'tab.close': '关闭标签 ({mod}W)',
  'tab.new': '新建标签 ({mod}T)',

  'target.label': '目标',
  'target.hostPlaceholder': 'example.com / 127.0.0.1',
  'target.portPlaceholder': '默认',
  'target.parseTitle': '从请求报文解析目标',
  'target.parse': '解析',
  'target.proxyToggleTitle': '下游代理开关（仅影响当前标签，下一次发送生效）',
  'target.proxyLabel': '代理: {state}',
  'target.proxyOn': '开',
  'target.proxyOff': '关',
  'target.proxySettingsTitle': '设置当前标签的代理地址',
  'target.proxyPopTitle': '当前标签代理地址',
  'target.proxyReset': '恢复默认（{url}）',
  'target.proxyHint': 'HTTP 代理；HTTPS 目标自动走 CONNECT 隧道；地址可带 user:pass 认证',
  'target.followHost': 'Host 跟随',
  'target.autoCL': '自动 CL',
  'target.send': '发送',
  'target.cancel': '取消',

  'request.title': '请求',
  'request.templates': '模板',
  'request.importTitle': '从剪贴板导入 curl 命令 / fetch 代码',
  'request.import': '导入',
  'request.importedHostMismatch': '导入报文含自定义 Host，已关闭「Host 跟随」',
  'request.imported': '已从 {source} 导入',
  'request.importFailed': '导入失败：{msg}',
  'request.clipboardEmpty': '剪贴板为空或无法读取',
  'request.ai': 'AI',
  'request.aiTitle': '和 AI 对话，修改当前请求或编写 payload',

  'settings.title': '设置',
  'settings.sectionGeneral': '外观与语言',
  'settings.language': '界面语言',
  'settings.langZh': '中文',
  'settings.langEn': 'English',
  'settings.theme': '主题',
  'settings.themeDark': '深色',
  'settings.themeLight': '浅色',
  'settings.sectionProxy': '默认代理',
  'settings.defaultProxy': '默认代理地址',
  'settings.defaultProxyHint': '新建标签页和「恢复默认」使用此地址，仅影响下游 HTTP 代理。',
  'settings.sectionLLM': 'LLM 配置',
  'settings.baseUrl': 'Base URL',
  'settings.baseUrlHint': '可填根地址、/v1 地址或完整 endpoint',
  'settings.apiKey': 'API Key',
  'settings.apiKeyHint': '仅保存在本机配置文件（权限 0600）',
  'settings.model': '模型',
  'settings.apiType': 'API 类型',
  'settings.apiTypeResponses': 'Responses API',
  'settings.apiTypeCompletions': 'Chat Completions',
  'settings.reasoning': '思考强度',
  'settings.reasoningDefault': '默认（不发送）',
  'settings.reasoningMinimal': 'minimal',
  'settings.reasoningLow': 'low',
  'settings.reasoningMedium': 'medium',
  'settings.reasoningHigh': 'high',
  'settings.systemPrompt': '系统提示词',
  'settings.systemPromptHint': '用于指导模型如何修改请求并输出原始 HTTP 报文',
  'settings.resetPrompt': '恢复默认提示词',
  'settings.save': '保存',
  'settings.cancel': '取消',
  'settings.saved': '设置已保存',
  'settings.saveFailed': '保存失败：{msg}',
  'settings.configPath': '配置文件：{path}',
  'settings.loadError': '读取配置时出现问题，当前使用默认值：{msg}',
  'settings.loading': '正在读取配置…',

  'ai.title': 'AI 助手',
  'ai.subtitle': '基于当前请求对话，结果可一键覆盖回编辑器',
  'ai.placeholder': '例如：把这个请求改成 JSON POST，并补充 Content-Type',
  'ai.send': '发送',
  'ai.thinking': '思考中…',
  'ai.empty': '让 AI 帮你修改请求、补充参数或编写 payload。当前编辑器里的请求会随每条消息一起发送。',
  'ai.currentRequest': '当前请求',
  'ai.requestHint': 'AI 输出不会自动发包。',
  'ai.apply': '应用到请求',
  'ai.applied': '已覆盖到当前请求（未发送）',
  'ai.applyFailed': '应用失败：{msg}',
  'ai.noRequest': '这条回复里没有可应用的 HTTP 请求',
  'ai.copy': '复制',
  'ai.copied': '已复制',
  'ai.clear': '清空对话',
  'ai.config': 'LLM 设置',
  'ai.notConfigured': '还没有配置 LLM。请先在设置中填写 Base URL 和模型（本地模型可不填 API Key）。',
  'ai.openSettings': '打开设置',
  'ai.error': '请求失败：{msg}',
  'ai.you': '你',
  'ai.assistant': 'AI',
  'ai.close': '关闭',
  'ai.inputHint': 'Ctrl/⌘+Enter 发送，Enter 换行',
  'ai.stop': '停止',
  'ai.cancelled': '已取消',
  'ai.resize': '拖动调整大小',
  'ai.suggestJsonPost': '把请求改成 JSON POST',
  'ai.suggestPayload': '帮我写一个测试 payload',
  'ai.suggestExplain': '解释这个请求的问题',

  'ctx.copyUrl': '复制 URL',
  'ctx.copyCurl': '复制为 curl',
  'ctx.convertToPost': '转换为 POST',
  'ctx.convertToGet': '转换为 GET',
  'ctx.convertUnsupported': '转换请求方法（仅 GET/POST）',
  'ctx.copied': '已复制',
  'ctx.copyUrlDone': 'URL 已复制',
  'ctx.copyCurlDone': 'curl 命令已复制',
  'ctx.copyFailed': '复制失败：{msg}',
  'ctx.convertFailed': '转换失败：{msg}',
  'ctx.converted': '已转换为 {method}',

  'response.title': '响应',
  'response.empty': '发送请求后，响应将展示在这里',
  'response.searchPlaceholder': '搜索响应',
  'response.noMatch': '无匹配',
  'response.prevMatch': '上一个匹配',
  'response.nextMatch': '下一个匹配',

  'history.failed': '失败',
  'history.discardTitle': '放弃未发送的修改？',
  'history.discardContent':
    '编辑器当前内容与所选历史快照不一致。切换后将加载历史快照，未发送的修改会被覆盖。',
  'history.discardConfirm': '继续切换',
  'history.cancel': '取消',
  'history.dirty': '当前编辑内容未发送，与历史快照不一致',
  'history.prev': '上一条历史',
  'history.next': '下一条历史',
  'history.viewList': '查看历史列表',
  'history.label': '历史 {label}',

  'status.noTarget': '未设置目标',
  'status.proxyOff': '代理未启用',
  'status.proxyOn': '代理 {url}',
  'status.history': '历史 {n} 条',

  'error.requestEmpty': '请求内容为空',
  'error.hostRequired': '请先填写目标 Host',
  'error.proxyEmpty': '已启用代理但代理地址为空，请先填写代理地址',
  'error.parseHostMissing': '未能从请求中解析出目标 Host',
  'error.parseSuccess': '解析成功：{url}',
  'error.parseFailed': '解析失败：{msg}',
  'error.sendFailed': '发送失败：{msg}',

  'tpl.get': 'GET 请求',
  'tpl.postForm': 'POST 表单',
  'tpl.postJson': 'POST JSON',
  'tpl.upload': '文件上传 (multipart)',
  'tpl.emptyHost': 'Host 留空',
  'tpl.clear': '清空',
}

export type MessageKey = keyof typeof zh

const en: Record<MessageKey, string> = {
  'app.subtitle': 'HTTP request replay tool',
  'app.switchToLight': 'Switch to light theme',
  'app.switchToDark': 'Switch to dark theme',
  'app.switchLang': '切换到中文',
  'app.byAuthor': 'by {name}',
  'app.settings': 'Settings',

  'tab.defaultTitle': 'Request {n}',
  'tab.close': 'Close tab ({mod}W)',
  'tab.new': 'New tab ({mod}T)',

  'target.label': 'Target',
  'target.hostPlaceholder': 'example.com / 127.0.0.1',
  'target.portPlaceholder': 'Default',
  'target.parseTitle': 'Parse target from request',
  'target.parse': 'Parse',
  'target.proxyToggleTitle': 'Downstream proxy (this tab only; applies to the next send)',
  'target.proxyLabel': 'Proxy: {state}',
  'target.proxyOn': 'on',
  'target.proxyOff': 'off',
  'target.proxySettingsTitle': 'Set the proxy URL for this tab',
  'target.proxyPopTitle': 'Proxy URL for this tab',
  'target.proxyReset': 'Reset to default ({url})',
  'target.proxyHint':
    'HTTP proxy; HTTPS targets use CONNECT tunneling; credentials can be embedded as user:pass',
  'target.followHost': 'Follow Host',
  'target.autoCL': 'Auto CL',
  'target.send': 'Send',
  'target.cancel': 'Cancel',

  'request.title': 'Request',
  'request.templates': 'Templates',
  'request.importTitle': 'Import a curl command or fetch code from the clipboard',
  'request.import': 'Import',
  'request.importedHostMismatch':
    "The imported request has a custom Host; 'Follow Host' was disabled",
  'request.imported': 'Imported from {source}',
  'request.importFailed': 'Import failed: {msg}',
  'request.clipboardEmpty': 'Clipboard is empty or cannot be read',
  'request.ai': 'AI',
  'request.aiTitle': 'Chat with AI to edit the current request or write a payload',

  'settings.title': 'Settings',
  'settings.sectionGeneral': 'Appearance & Language',
  'settings.language': 'Language',
  'settings.langZh': '中文',
  'settings.langEn': 'English',
  'settings.theme': 'Theme',
  'settings.themeDark': 'Dark',
  'settings.themeLight': 'Light',
  'settings.sectionProxy': 'Default Proxy',
  'settings.defaultProxy': 'Default proxy URL',
  'settings.defaultProxyHint': 'Used by new tabs and "Reset to default"; applies to the downstream HTTP proxy.',
  'settings.sectionLLM': 'LLM',
  'settings.baseUrl': 'Base URL',
  'settings.baseUrlHint': 'A root URL, a /v1 URL, or a full endpoint',
  'settings.apiKey': 'API Key',
  'settings.apiKeyHint': 'Stored only in the local config file (mode 0600)',
  'settings.model': 'Model',
  'settings.apiType': 'API type',
  'settings.apiTypeResponses': 'Responses API',
  'settings.apiTypeCompletions': 'Chat Completions',
  'settings.reasoning': 'Reasoning effort',
  'settings.reasoningDefault': 'Default (omit)',
  'settings.reasoningMinimal': 'minimal',
  'settings.reasoningLow': 'low',
  'settings.reasoningMedium': 'medium',
  'settings.reasoningHigh': 'high',
  'settings.systemPrompt': 'System prompt',
  'settings.systemPromptHint': 'Guides the model to edit requests and output raw HTTP',
  'settings.resetPrompt': 'Reset to default',
  'settings.save': 'Save',
  'settings.cancel': 'Cancel',
  'settings.saved': 'Settings saved',
  'settings.saveFailed': 'Save failed: {msg}',
  'settings.configPath': 'Config file: {path}',
  'settings.loadError': 'There was a problem reading the config; defaults are in use: {msg}',
  'settings.loading': 'Loading settings…',

  'ai.title': 'AI assistant',
  'ai.subtitle': 'Chat about the current request; apply the result back to the editor',
  'ai.placeholder': 'e.g. Turn this request into a JSON POST and add Content-Type',
  'ai.send': 'Send',
  'ai.thinking': 'Thinking…',
  'ai.empty': 'Ask AI to edit the request, add parameters, or write a payload. The current request is sent with each message.',
  'ai.currentRequest': 'Current request',
  'ai.requestHint': 'AI output is never sent automatically.',
  'ai.apply': 'Apply to request',
  'ai.applied': 'Applied to the current request (not sent)',
  'ai.applyFailed': 'Apply failed: {msg}',
  'ai.noRequest': 'This reply contains no applicable HTTP request',
  'ai.copy': 'Copy',
  'ai.copied': 'Copied',
  'ai.clear': 'Clear chat',
  'ai.config': 'LLM settings',
  'ai.notConfigured': 'LLM is not configured. Set the base URL and model in Settings first (a local model may not need an API key).',
  'ai.openSettings': 'Open settings',
  'ai.error': 'Request failed: {msg}',
  'ai.you': 'You',
  'ai.assistant': 'AI',
  'ai.close': 'Close',
  'ai.inputHint': 'Ctrl/⌘+Enter to send, Enter for a new line',
  'ai.stop': 'Stop',
  'ai.cancelled': 'Cancelled',
  'ai.resize': 'Drag to resize',
  'ai.suggestJsonPost': 'Turn this into a JSON POST',
  'ai.suggestPayload': 'Write a test payload',
  'ai.suggestExplain': "Explain what's wrong with this request",

  'ctx.copyUrl': 'Copy URL',
  'ctx.copyCurl': 'Copy as curl',
  'ctx.convertToPost': 'Change to POST',
  'ctx.convertToGet': 'Change to GET',
  'ctx.convertUnsupported': 'Change method (GET/POST only)',
  'ctx.copied': 'Copied',
  'ctx.copyUrlDone': 'URL copied',
  'ctx.copyCurlDone': 'curl command copied',
  'ctx.copyFailed': 'Copy failed: {msg}',
  'ctx.convertFailed': 'Conversion failed: {msg}',
  'ctx.converted': 'Changed to {method}',

  'response.title': 'Response',
  'response.empty': 'The response will appear here after you send a request',
  'response.searchPlaceholder': 'Search response',
  'response.noMatch': 'No match',
  'response.prevMatch': 'Previous match',
  'response.nextMatch': 'Next match',

  'history.failed': 'Failed',
  'history.discardTitle': 'Discard unsent changes?',
  'history.discardContent':
    'The editor differs from the selected history snapshot. Switching will load the snapshot and overwrite unsent changes.',
  'history.discardConfirm': 'Switch anyway',
  'history.cancel': 'Cancel',
  'history.dirty': 'Current edits are unsent and differ from the history snapshot',
  'history.prev': 'Previous history',
  'history.next': 'Next history',
  'history.viewList': 'View history list',
  'history.label': 'History {label}',

  'status.noTarget': 'No target',
  'status.proxyOff': 'Proxy off',
  'status.proxyOn': 'Proxy {url}',
  'status.history': 'History: {n}',

  'error.requestEmpty': 'Request is empty',
  'error.hostRequired': 'Enter a target Host first',
  'error.proxyEmpty': 'Proxy is enabled but the proxy URL is empty',
  'error.parseHostMissing': 'Could not parse a target Host from the request',
  'error.parseSuccess': 'Parsed: {url}',
  'error.parseFailed': 'Parse failed: {msg}',
  'error.sendFailed': 'Send failed: {msg}',

  'tpl.get': 'GET request',
  'tpl.postForm': 'POST form',
  'tpl.postJson': 'POST JSON',
  'tpl.upload': 'File upload (multipart)',
  'tpl.emptyHost': 'Empty Host',
  'tpl.clear': 'Clear',
}

const messages: Record<Lang, Record<MessageKey, string>> = { zh, en }

function detectLang(): Lang {
  try {
    const stored = localStorage.getItem(LANG_STORAGE_KEY)
    if (stored === 'zh' || stored === 'en') return stored
  } catch {
    // localStorage 不可用时按浏览器语言判断
  }
  return navigator.language?.toLowerCase().startsWith('zh') ? 'zh' : 'en'
}

export const lang = ref<Lang>(detectLang())

function applyDocumentLang(value: Lang) {
  document.documentElement.lang = value === 'zh' ? 'zh-CN' : 'en'
}

export function setLang(value: Lang) {
  lang.value = value
  applyDocumentLang(value)
  try {
    localStorage.setItem(LANG_STORAGE_KEY, value)
  } catch {
    // 存储不可用不影响本次会话
  }
}

export function toggleLang() {
  setLang(lang.value === 'zh' ? 'en' : 'zh')
}

export function initLang() {
  applyDocumentLang(lang.value)
}

export function t(key: MessageKey, params?: Record<string, string | number>): string {
  let text = messages[lang.value][key] ?? messages.zh[key] ?? key
  if (params) {
    for (const [name, value] of Object.entries(params)) {
      text = text.replaceAll(`{${name}}`, () => String(value))
    }
  }
  return text
}
