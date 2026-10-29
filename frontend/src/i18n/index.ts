import { ref } from 'vue'

export type Lang = 'zh' | 'en'

const LANG_STORAGE_KEY = 'httpRepeater.lang'

const zh = {
  'app.subtitle': 'HTTP 请求重放工具',
  'app.switchToLight': '切换到浅色主题',
  'app.switchToDark': '切换到深色主题',
  'app.switchLang': 'Switch to English',
  'app.byAuthor': '作者 {name}',

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

  'request.title': '请求',
  'request.templates': '模板',
  'request.importTitle': '从剪贴板导入 curl 命令 / fetch 代码',
  'request.import': '导入',
  'request.importedHostMismatch': '导入报文含自定义 Host，已关闭「Host 跟随」',
  'request.imported': '已从 {source} 导入',
  'request.importFailed': '导入失败：{msg}',
  'request.clipboardEmpty': '剪贴板为空或无法读取',

  'response.title': '响应',
  'response.empty': '发送请求后，响应将展示在这里',

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

  'request.title': 'Request',
  'request.templates': 'Templates',
  'request.importTitle': 'Import a curl command or fetch code from the clipboard',
  'request.import': 'Import',
  'request.importedHostMismatch':
    "The imported request has a custom Host; 'Follow Host' was disabled",
  'request.imported': 'Imported from {source}',
  'request.importFailed': 'Import failed: {msg}',
  'request.clipboardEmpty': 'Clipboard is empty or cannot be read',

  'response.title': 'Response',
  'response.empty': 'The response will appear here after you send a request',

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
