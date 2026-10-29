<script lang="ts" setup>
import { computed, ref } from 'vue'
import { NRadioButton, NRadioGroup } from 'naive-ui'
import { useTabsStore } from '../stores/tabs'
import HistoryNav from './HistoryNav.vue'
import { t } from '../i18n'

const store = useTabsStore()
const mode = ref<'pretty' | 'raw'>('pretty')

const historyItem = computed(() => {
  const tab = store.activeTab
  if (!tab || tab.activeHistoryIndex < 0) return null
  return tab.histories[tab.activeHistoryIndex] ?? null
})

const response = computed(() => historyItem.value?.response ?? null)

const headerLines = computed(() => {
  const raw = response.value?.headersRaw ?? ''
  return raw ? raw.split(/\r?\n/).filter((l) => l.trim() !== '') : []
})

const bodyText = computed(() => response.value?.bodyDecoded ?? '')

const summary = computed(() => {
  const r = response.value
  if (!r) return ''
  if (r.error) return r.error
  return `${r.statusLine} · ${r.durationMs} ms · ${formatBytes(r.bytes)}`
})

function formatBytes(n: number): string {
  if (n < 1024) return `${n} B`
  if (n < 1024 * 1024) return `${(n / 1024).toFixed(1)} KB`
  return `${(n / (1024 * 1024)).toFixed(2)} MB`
}

const prettyBody = computed(() => {
  const text = bodyText.value
  if (!text) return ''
  const contentType = response.value?.headersRaw.toLowerCase() ?? ''
  const looksJson = contentType.includes('json') || /^\s*[[{]/.test(text)
  if (looksJson) {
    try {
      return JSON.stringify(JSON.parse(text), null, 2)
    } catch {
      return text
    }
  }
  return text
})

const prettyText = computed(() => {
  const r = response.value
  if (!r) return ''
  if (r.error) return r.error
  const head = [r.statusLine, ...headerLines.value].join('\n')
  return `${head}\n\n${prettyBody.value}`
})

const rawText = computed(() => {
  const r = response.value
  if (!r) return ''
  if (r.error) return r.error
  const head = [r.statusLine, ...headerLines.value].join('\n')
  return `${head}\n\n${bodyText.value}`
})
</script>

<template>
  <div class="panel">
    <div class="panel-header">
      <span class="panel-title">{{ t('response.title') }}</span>

      <HistoryNav />

      <template v-if="response">
        <n-radio-group v-model:value="mode" size="small">
          <n-radio-button value="pretty">Pretty</n-radio-button>
          <n-radio-button value="raw">Raw</n-radio-button>
        </n-radio-group>

        <span class="status" :class="{ 'status-error': response.error }" :title="summary">
          {{ summary }}
        </span>
      </template>
    </div>

    <div class="panel-body">
      <div v-if="!response" class="panel-empty mono">{{ t('response.empty') }}</div>
      <pre
        v-else
        class="response-pre mono"
        :class="{ 'error-text': response.error }"
      >{{ mode === 'pretty' ? prettyText : rawText }}</pre>
    </div>
  </div>
</template>

<style scoped>
.panel {
  height: 100%;
  display: flex;
  flex-direction: column;
  background: var(--app-bg);
  overflow: hidden;
}

.panel-header {
  flex: none;
  height: 32px;
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 0 12px;
  border-bottom: 1px solid var(--border-soft);
  background: var(--bar-bg);
}

.panel-title {
  font-size: 12.5px;
  font-weight: 600;
  color: var(--text);
  letter-spacing: 1px;
  flex: none;
}

.status {
  flex: 1;
  min-width: 0;
  text-align: right;
  font-size: 12px;
  color: var(--accent-soft);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.status-error {
  color: var(--error);
}

.panel-body {
  flex: 1;
  min-height: 0;
  display: flex;
  align-items: stretch;
  justify-content: center;
}

.panel-empty {
  align-self: center;
  font-size: 12px;
  color: var(--text-dim);
}

.response-pre {
  width: 100%;
  height: 100%;
  margin: 0;
  padding: 10px 12px;
  overflow: auto;
  font-size: 12px;
  line-height: 1.6;
  color: var(--text-body);
  user-select: text;
  white-space: pre;
}

.error-text {
  color: var(--error);
}
</style>
