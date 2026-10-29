import { defineStore } from 'pinia'

// 全局弹层状态：设置页与 AI 对话页可以从任意位置打开
export const useUiStore = defineStore('ui', {
  state: () => ({
    settingsOpen: false,
    aiOpen: false,
  }),
  actions: {
    openSettings() {
      this.settingsOpen = true
    },
    closeSettings() {
      this.settingsOpen = false
    },
    openAi() {
      this.aiOpen = true
    },
    closeAi() {
      this.aiOpen = false
    },
  },
})
