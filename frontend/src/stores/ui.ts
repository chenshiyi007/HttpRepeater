import { defineStore } from 'pinia'

export type SideRailTab = 'ai' | 'decode' | 'time'

const RAIL_WIDTH_KEY = 'httpRepeater.sideRailWidth'

function readRailWidth(): number {
  try {
    const n = Number(localStorage.getItem(RAIL_WIDTH_KEY))
    if (Number.isFinite(n)) return Math.min(560, Math.max(280, Math.round(n)))
  } catch {
    // ignore
  }
  return 360
}

function writeRailWidth(width: number) {
  try {
    localStorage.setItem(RAIL_WIDTH_KEY, String(width))
  } catch {
    // ignore
  }
}

// 全局弹层 / 右侧边栏（AI、解码、时间）
export const useUiStore = defineStore('ui', {
  state: () => ({
    settingsOpen: false,
    sideRailOpen: true,
    sideRailTab: 'ai' as SideRailTab,
    sideRailWidth: readRailWidth(),
    decoderInput: '',
  }),
  getters: {
    sideRailCollapsed(state): boolean {
      return !state.sideRailOpen
    },
  },
  actions: {
    openSettings() {
      this.settingsOpen = true
    },
    closeSettings() {
      this.settingsOpen = false
    },
    openSideRail(tab: SideRailTab) {
      this.sideRailTab = tab
      this.sideRailOpen = true
    },
    openAi() {
      this.openSideRail('ai')
    },
    toggleSideRail() {
      this.sideRailOpen = !this.sideRailOpen
    },
    closeSideRail() {
      this.sideRailOpen = false
    },
    setSideRailTab(tab: SideRailTab) {
      this.sideRailTab = tab
      if (!this.sideRailOpen) this.sideRailOpen = true
    },
    setSideRailWidth(width: number) {
      this.sideRailWidth = Math.min(560, Math.max(280, Math.round(width)))
      writeRailWidth(this.sideRailWidth)
    },
    sendToDecoder(text: string) {
      this.decoderInput = text
      this.openSideRail('decode')
    },
  },
})
