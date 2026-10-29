import { defineStore } from 'pinia'
import type { HistoryItem, RepeaterTab, ResponseResult, SendRequest, Target } from '../types'
import { cancelSend, sendRequest } from '../api/repeater'
import { lang, t } from '../i18n'
import { DEFAULT_PROXY_URL, useSettingsStore } from './settings'

let uidCounter = 0

function nextTabId(): string {
  uidCounter += 1
  return `tab-${Date.now().toString(36)}-${uidCounter}`
}

let historyCounter = 0

function nextHistoryId(): string {
  historyCounter += 1
  return `hist-${Date.now().toString(36)}-${historyCounter}`
}

function createTab(seq: number, proxyURL: string): RepeaterTab {
  return {
    id: nextTabId(),
    title: t('tab.defaultTitle', { n: seq }),
    requestText: '',
    target: {
      scheme: 'http',
      host: '',
      port: 0,
      useProxy: false,
      proxyURL: proxyURL || DEFAULT_PROXY_URL,
    },
    followHostHeader: true,
    autoContentLength: true,
    histories: [],
    activeHistoryIndex: -1,
    updatedAt: Date.now(),
  }
}

// 标签页状态：增/删/改名/切换 + 目标编辑 + 发送
export const useTabsStore = defineStore('tabs', {
  state: () => {
    const first = createTab(1, DEFAULT_PROXY_URL)
    return {
      tabs: [first] as RepeaterTab[],
      activeTabId: first.id,
      tabSeq: 1,
      sending: false,
    }
  },
  getters: {
    activeTab(state): RepeaterTab | undefined {
      return state.tabs.find((t) => t.id === state.activeTabId)
    },
    // 当前选中的历史（响应面板与状态栏共用）
    activeHistory(state): HistoryItem | null {
      const tab = state.tabs.find((t) => t.id === state.activeTabId)
      if (!tab || tab.activeHistoryIndex < 0) return null
      return tab.histories[tab.activeHistoryIndex] ?? null
    },
    // 编辑器内容与所选历史快照不一致（未发送的修改）
    historyDirty(state): boolean {
      const tab = state.tabs.find((t) => t.id === state.activeTabId)
      if (!tab || tab.activeHistoryIndex < 0) return false
      const h = tab.histories[tab.activeHistoryIndex]
      return !!h && tab.requestText !== h.requestText
    },
  },
  actions: {
    addTab() {
      this.tabSeq += 1
      const settings = useSettingsStore()
      const tab = createTab(this.tabSeq, settings.defaultProxyUrl || DEFAULT_PROXY_URL)
      this.tabs.push(tab)
      this.activeTabId = tab.id
    },
    closeTab(id: string) {
      if (this.tabs.length <= 1) return // 至少保留一个标签
      const idx = this.tabs.findIndex((t) => t.id === id)
      if (idx === -1) return
      this.tabs.splice(idx, 1)
      if (this.activeTabId === id) {
        const next = this.tabs[Math.min(idx, this.tabs.length - 1)]
        this.activeTabId = next.id
      }
    },
    renameTab(id: string, title: string) {
      const trimmed = title.trim()
      if (!trimmed) return
      const tab = this.tabs.find((t) => t.id === id)
      if (tab) {
        tab.title = trimmed
        tab.updatedAt = Date.now()
      }
    },
    setActive(id: string) {
      this.activeTabId = id
    },
    setRequestText(text: string) {
      const tab = this.activeTab
      if (tab) {
        tab.requestText = text
        tab.updatedAt = Date.now()
      }
    },
    updateTarget(partial: Partial<Target>) {
      const tab = this.activeTab
      if (tab) {
        Object.assign(tab.target, partial)
        tab.updatedAt = Date.now()
      }
    },
    setFollowHostHeader(value: boolean) {
      const tab = this.activeTab
      if (tab) {
        tab.followHostHeader = value
        tab.updatedAt = Date.now()
      }
    },
    setAutoContentLength(value: boolean) {
      const tab = this.activeTab
      if (tab) {
        tab.autoContentLength = value
        tab.updatedAt = Date.now()
      }
    },
    // 默认代理变化时，同步仍在使用旧默认值的标签
    applyDefaultProxy(oldURL: string, newURL: string) {
      const previous = (oldURL || '').trim()
      const next = (newURL || '').trim()
      if (!next || previous === next) return
      for (const tab of this.tabs) {
        if ((tab.target.proxyURL || '').trim() === previous) {
          tab.target.proxyURL = next
        }
      }
    },
    // 加载某条历史快照到编辑器（Burp 行为：切换历史 = 载入该请求，可改后重发）
    loadHistory(index: number) {
      const tab = this.activeTab
      if (!tab) return
      const h = tab.histories[index]
      if (!h) return
      tab.activeHistoryIndex = index
      tab.requestText = h.requestText
      tab.updatedAt = Date.now()
    },

    // 取消正在发送的请求；Go 侧关闭连接，sendActiveTab 会收到「已取消」结果并入历史
    async cancelSending(): Promise<void> {
      if (!this.sending) return
      try {
        await cancelSend()
      } catch {
        // 请求可能已经结束，忽略取消失败
      }
    },
    // 发送当前标签的请求；返回 error 表示未发出（网络错误会记入历史，不算 error）
    async sendActiveTab(): Promise<{ error?: string }> {      const tab = this.activeTab
      if (!tab || this.sending) return {}
      if (!tab.requestText.trim()) return { error: t('error.requestEmpty') }
      if (!tab.target.host) return { error: t('error.hostRequired') }

      this.sending = true
      try {
        if (tab.target.useProxy && !tab.target.proxyURL.trim()) {
          return { error: t('error.proxyEmpty') }
        }
        const payload: SendRequest = {
          requestText: tab.requestText,
          scheme: tab.target.scheme,
          host: tab.target.host,
          port: tab.target.port,
          useProxy: tab.target.useProxy,
          proxyURL: tab.target.proxyURL,
          followHostHeader: tab.followHostHeader,
          autoContentLength: tab.autoContentLength,
          verifyTLS: false, // 默认忽略证书校验
          timeoutMs: 30000,
          lang: lang.value,
        }
        const res = await sendRequest(payload)
        const response: ResponseResult | null = res.response ?? null
        const history: HistoryItem = {
          id: nextHistoryId(),
          seq: tab.histories.length + 1,
          requestText: tab.requestText,
          requestSentText: res.requestSent,
          target: { ...tab.target },
          response,
          sentAt: Date.now(),
        }
        tab.histories.push(history)
        tab.activeHistoryIndex = tab.histories.length - 1
        tab.updatedAt = Date.now()
        return {}
      } catch (e) {
        const msg = e instanceof Error ? e.message : String(e)
        return { error: t('error.sendFailed', { msg }) }
      } finally {
        this.sending = false
      }
    },
  },
})
