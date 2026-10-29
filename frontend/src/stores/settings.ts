import { defineStore } from 'pinia'

export type ThemeMode = 'dark' | 'light'

const THEME_STORAGE_KEY = 'httpRepeater.theme'

function readStoredTheme(): ThemeMode {
  try {
    return localStorage.getItem(THEME_STORAGE_KEY) === 'light' ? 'light' : 'dark'
  } catch {
    return 'dark'
  }
}

function applyTheme(mode: ThemeMode) {
  document.body.dataset.theme = mode
}

// 应用设置：主题偏好，存 localStorage
export const useSettingsStore = defineStore('settings', {
  state: () => ({
    theme: readStoredTheme(),
  }),
  actions: {
    initTheme() {
      applyTheme(this.theme)
    },
    setTheme(mode: ThemeMode) {
      this.theme = mode
      applyTheme(mode)
      try {
        localStorage.setItem(THEME_STORAGE_KEY, mode)
      } catch {
        // 存储不可用（隐私模式等）不影响本次会话切换
      }
    },
    toggleTheme() {
      this.setTheme(this.theme === 'dark' ? 'light' : 'dark')
    },
  },
})
