<script lang="ts" setup>
import { computed, reactive, ref, watch } from 'vue'
import {
  NAlert,
  NButton,
  NDivider,
  NInput,
  NModal,
  NSelect,
  useMessage,
  type SelectOption,
} from 'naive-ui'
import { useSettingsStore, DEFAULT_PROXY_URL } from '../stores/settings'
import { useTabsStore } from '../stores/tabs'
import { useUiStore } from '../stores/ui'
import { t } from '../i18n'
import type { AppSettings, ReasoningEffort } from '../types'

const settings = useSettingsStore()
const tabs = useTabsStore()
const ui = useUiStore()
const message = useMessage()

const saving = ref(false)
const saveError = ref('')

const draft = reactive<AppSettings>({
  language: 'zh',
  theme: 'dark',
  defaultProxyUrl: DEFAULT_PROXY_URL,
  llm: {
    baseUrl: '',
    apiKey: '',
    model: '',
    apiType: 'completions',
    reasoningEffort: '',
    systemPrompt: '',
  },
})

const languageOptions = computed<SelectOption[]>(() => [
  { label: t('settings.langZh'), value: 'zh' },
  { label: t('settings.langEn'), value: 'en' },
])
const themeOptions = computed<SelectOption[]>(() => [
  { label: t('settings.themeDark'), value: 'dark' },
  { label: t('settings.themeLight'), value: 'light' },
])
const apiTypeOptions = computed<SelectOption[]>(() => [
  { label: t('settings.apiTypeCompletions'), value: 'completions' },
  { label: t('settings.apiTypeResponses'), value: 'responses' },
])
const reasoningOptions = computed<SelectOption[]>(() => [
  { label: t('settings.reasoningDefault'), value: '' },
  { label: t('settings.reasoningMinimal'), value: 'minimal' },
  { label: t('settings.reasoningLow'), value: 'low' },
  { label: t('settings.reasoningMedium'), value: 'medium' },
  { label: t('settings.reasoningHigh'), value: 'high' },
])

function resetDraft() {
  draft.language = settings.language
  draft.theme = settings.theme
  draft.defaultProxyUrl = settings.defaultProxyUrl
  draft.llm = { ...settings.llm }
}

function onLanguageChange(value: string | number | null) {
  draft.language = value === 'en' ? 'en' : 'zh'
}

function onThemeChange(value: string | number | null) {
  draft.theme = value === 'light' ? 'light' : 'dark'
}

function onApiTypeChange(value: string | number | null) {
  draft.llm.apiType = value === 'responses' ? 'responses' : 'completions'
}

function onReasoningChange(value: string | number | null) {
  draft.llm.reasoningEffort =
    value === 'minimal' || value === 'low' || value === 'medium' || value === 'high'
      ? (value as ReasoningEffort)
      : ''
}

function resetPrompt() {
  draft.llm.systemPrompt = settings.defaultSystemPrompt || ''
}

function close() {
  if (saving.value) return
  ui.closeSettings()
}

async function save() {
  if (saving.value) return
  saving.value = true
  saveError.value = ''
  const previousDefaultProxy = settings.defaultProxyUrl
  try {
    await settings.saveAll({
      language: draft.language,
      theme: draft.theme,
      defaultProxyUrl: draft.defaultProxyUrl,
      llm: { ...draft.llm },
    })
    if (previousDefaultProxy !== settings.defaultProxyUrl) {
      tabs.applyDefaultProxy(previousDefaultProxy, settings.defaultProxyUrl)
    }
    message.success(t('settings.saved'))
    ui.closeSettings()
  } catch (e) {
    const msg = e instanceof Error ? e.message : String(e)
    saveError.value = t('settings.saveFailed', { msg })
    message.error(saveError.value)
  } finally {
    saving.value = false
  }
}

watch(
  () => ui.settingsOpen,
  async (open) => {
    if (!open) return
    if (!settings.loaded) await settings.load()
    resetDraft()
    saveError.value = ''
  },
)
</script>

<template>
  <n-modal
    :show="ui.settingsOpen"
    preset="card"
    :bordered="false"
    :mask-closable="false"
    :auto-focus="false"
    style="width: 720px; max-width: calc(100vw - 48px)"
    :title="t('settings.title')"
    @update:show="(value: boolean) => !value && close()"
  >
    <div class="settings-body">
      <n-alert v-if="settings.loadError" type="warning" :bordered="false" class="alert">
        {{ t('settings.loadError', { msg: settings.loadError }) }}
      </n-alert>
      <n-alert v-if="saveError" type="error" :bordered="false" class="alert">
        {{ saveError }}
      </n-alert>

      <n-divider title-placement="left">{{ t('settings.sectionGeneral') }}</n-divider>
      <div class="form-grid">
        <label class="field-label">{{ t('settings.language') }}</label>
        <n-select
          :value="draft.language"
          :options="languageOptions"
          @update:value="onLanguageChange"
        />

        <label class="field-label">{{ t('settings.theme') }}</label>
        <n-select :value="draft.theme" :options="themeOptions" @update:value="onThemeChange" />
      </div>

      <n-divider title-placement="left">{{ t('settings.sectionProxy') }}</n-divider>
      <div class="form-grid">
        <label class="field-label">{{ t('settings.defaultProxy') }}</label>
        <n-input
          :value="draft.defaultProxyUrl"
          :placeholder="DEFAULT_PROXY_URL"
          @update:value="(value: string) => (draft.defaultProxyUrl = value)"
        />
      </div>
      <div class="field-hint">{{ t('settings.defaultProxyHint') }}</div>

      <n-divider title-placement="left">{{ t('settings.sectionLLM') }}</n-divider>
      <div class="form-grid">
        <label class="field-label">{{ t('settings.baseUrl') }}</label>
        <n-input
          :value="draft.llm.baseUrl"
          placeholder="https://api.openai.com/v1"
          @update:value="(value: string) => (draft.llm.baseUrl = value)"
        />

        <label class="field-label">{{ t('settings.apiKey') }}</label>
        <n-input
          :value="draft.llm.apiKey"
          type="password"
          show-password-on="click"
          placeholder="sk-..."
          @update:value="(value: string) => (draft.llm.apiKey = value)"
        />

        <label class="field-label">{{ t('settings.model') }}</label>
        <n-input
          :value="draft.llm.model"
          placeholder="gpt-4o-mini / deepseek-chat ..."
          @update:value="(value: string) => (draft.llm.model = value)"
        />

        <label class="field-label">{{ t('settings.apiType') }}</label>
        <n-select
          :value="draft.llm.apiType"
          :options="apiTypeOptions"
          @update:value="onApiTypeChange"
        />

        <label class="field-label">{{ t('settings.reasoning') }}</label>
        <n-select
          :value="draft.llm.reasoningEffort"
          :options="reasoningOptions"
          @update:value="onReasoningChange"
        />
      </div>
      <div class="field-hint">
        {{ t('settings.baseUrlHint') }} · {{ t('settings.apiKeyHint') }}
      </div>

      <div class="prompt-block">
        <div class="prompt-head">
          <label class="field-label">{{ t('settings.systemPrompt') }}</label>
          <n-button text size="tiny" @click="resetPrompt">{{ t('settings.resetPrompt') }}</n-button>
        </div>
        <n-input
          :value="draft.llm.systemPrompt"
          type="textarea"
          :autosize="{ minRows: 4, maxRows: 10 }"
          @update:value="(value: string) => (draft.llm.systemPrompt = value)"
        />
        <div class="field-hint">{{ t('settings.systemPromptHint') }}</div>
      </div>
    </div>

    <template #footer>
      <div class="settings-footer">
        <span class="config-path mono">{{ t('settings.configPath', { path: settings.configPath }) }}</span>
        <div class="footer-spacer"></div>
        <n-button size="small" :disabled="saving" @click="close">{{ t('settings.cancel') }}</n-button>
        <n-button size="small" type="primary" :loading="saving" @click="save">
          {{ t('settings.save') }}
        </n-button>
      </div>
    </template>
  </n-modal>
</template>

<style scoped>
.settings-body {
  max-height: calc(100vh - 240px);
  overflow-y: auto;
  padding-right: 4px;
}

.alert {
  margin-bottom: 10px;
}

.form-grid {
  display: grid;
  grid-template-columns: 120px minmax(0, 1fr);
  align-items: center;
  gap: 10px 12px;
  padding: 2px 8px;
}

.field-label {
  font-size: 12px;
  color: var(--text);
}

.field-hint {
  margin: 6px 8px 2px 140px;
  font-size: 11px;
  line-height: 1.5;
  color: var(--text-muted);
}

.prompt-block {
  margin-top: 12px;
}

.prompt-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 6px;
}

.settings-footer {
  display: flex;
  align-items: center;
  gap: 8px;
  min-width: 0;
}

.config-path {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 11px;
  color: var(--text-dim);
}

.footer-spacer {
  flex: 1;
}
</style>
