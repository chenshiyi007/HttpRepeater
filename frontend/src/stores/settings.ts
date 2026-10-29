import { defineStore } from 'pinia'
import { loadSettings, saveSettings } from '../api/repeater'
import { lang, setLang, type Lang } from '../i18n'
import type { AppSettings, LLMSettings, ReasoningEffort } from '../types'

export type ThemeMode = 'dark' | 'light'

const THEME_STORAGE_KEY = 'httpRepeater.theme'

// 默认下游代理地址（Clash / mihomo 常见本地端口）
export const DEFAULT_PROXY_URL = 'http://127.0.0.1:7890'

const DEFAULT_LLM: LLMSettings = {
  baseUrl: 'https://api.openai.com/v1',
  apiKey: '',
  model: '',
  apiType: 'completions',
  reasoningEffort: '',
  systemPrompt: '',
}

function defaultLLM(): LLMSettings {
  return { ...DEFAULT_LLM }
}

function readStoredTheme(): ThemeMode {
  try {
    return localStorage.getItem(THEME_STORAGE_KEY) === 'light' ? 'light' : 'dark'
  } catch {
    return 'dark'
  }
}

function writeStoredTheme(mode: ThemeMode) {
  try {
    localStorage.setItem(THEME_STORAGE_KEY, mode)
  } catch {
    // 存储不可用（隐私模式等）不影响本次会话切换
  }
}

function applyTheme(mode: ThemeMode) {
  document.body.dataset.theme = mode
}

function normalizeLang(value: unknown): Lang {
  return value === 'en' ? 'en' : 'zh'
}

function normalizeTheme(value: unknown): ThemeMode {
  return value === 'light' ? 'light' : 'dark'
}

function normalizeApiType(value: unknown): LLMSettings['apiType'] {
  return value === 'responses' ? 'responses' : 'completions'
}

function normalizeEffort(value: unknown): ReasoningEffort {
  return value === 'minimal' || value === 'low' || value === 'medium' || value === 'high' ? value : ''
}

export const useSettingsStore = defineStore('settings', {
  state: () => ({
    theme: readStoredTheme(),
    language: lang.value,
    defaultProxyUrl: DEFAULT_PROXY_URL,
    llm: defaultLLM(),
    configPath: '~/.httpRepeater/setting.json',
    defaultSystemPrompt: '',
    loaded: false,
    loading: false,
    loadError: '',
    saveError: '',
  }),
  getters: {
    llmConfigured(state): boolean {
      return !!state.llm.baseUrl.trim() && !!state.llm.model.trim()
    },
  },
  actions: {
    initTheme() {
      applyTheme(this.theme)
    },
    async load() {
      if (this.loading || this.loaded) return
      this.loading = true
      try {
        const res = await loadSettings()
        this.configPath = res.path || this.configPath
        this.defaultSystemPrompt = res.defaultSystemPrompt || ''
        this.loadError = res.error || ''

        const s = res.settings
        this.defaultProxyUrl = (s.defaultProxyUrl || '').trim() || DEFAULT_PROXY_URL
        this.llm = {
          ...defaultLLM(),
          ...s.llm,
          apiType: normalizeApiType(s.llm?.apiType),
          reasoningEffort: normalizeEffort(s.llm?.reasoningEffort),
          systemPrompt: s.llm?.systemPrompt || this.defaultSystemPrompt,
        }
        if (res.exists) {
          this.language = normalizeLang(s.language)
          this.theme = normalizeTheme(s.theme)
        } else {
          // 首次运行：沿用浏览器语言和本地主题
          this.language = lang.value
          this.theme = readStoredTheme()
        }
        setLang(this.language)
        applyTheme(this.theme)
        writeStoredTheme(this.theme)
      } catch (e) {
        this.loadError = e instanceof Error ? e.message : String(e)
        setLang(this.language)
        applyTheme(this.theme)
      } finally {
        this.loading = false
        this.loaded = true
      }
    },
    async persist() {
      if (!this.loaded) return
      const payload: AppSettings = {
        language: this.language,
        theme: this.theme,
        defaultProxyUrl: this.defaultProxyUrl,
        llm: { ...this.llm },
      }
      await saveSettings(payload)
    },
    async setTheme(mode: ThemeMode) {
      this.theme = normalizeTheme(mode)
      applyTheme(this.theme)
      writeStoredTheme(this.theme)
      try {
        await this.persist()
      } catch (e) {
        this.saveError = e instanceof Error ? e.message : String(e)
      }
    },
    async setLanguage(value: Lang) {
      this.language = normalizeLang(value)
      setLang(this.language)
      try {
        await this.persist()
      } catch (e) {
        this.saveError = e instanceof Error ? e.message : String(e)
      }
    },
    toggleTheme() {
      void this.setTheme(this.theme === 'dark' ? 'light' : 'dark')
    },
    // 设置页保存：本地立即生效，再写配置文件；失败时抛给调用方提示
    async saveAll(next: AppSettings): Promise<void> {
      this.language = normalizeLang(next.language)
      this.theme = normalizeTheme(next.theme)
      this.defaultProxyUrl = (next.defaultProxyUrl || '').trim() || DEFAULT_PROXY_URL
      this.llm = {
        baseUrl: next.llm.baseUrl.trim(),
        apiKey: next.llm.apiKey.trim(),
        model: next.llm.model.trim(),
        apiType: normalizeApiType(next.llm.apiType),
        reasoningEffort: normalizeEffort(next.llm.reasoningEffort),
        systemPrompt: next.llm.systemPrompt.trim() || this.defaultSystemPrompt,
      }
      setLang(this.language)
      applyTheme(this.theme)
      writeStoredTheme(this.theme)
      this.saveError = ''
      try {
        await this.persist()
      } catch (e) {
        this.saveError = e instanceof Error ? e.message : String(e)
        throw e
      }
    },
  },
})
