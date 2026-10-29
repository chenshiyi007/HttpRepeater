<script lang="ts" setup>
import { computed, nextTick, ref, watch } from 'vue'
import { NButton, NIcon, NInput, NRadioButton, NRadioGroup } from 'naive-ui'
import { ChevronDownOutline, ChevronUpOutline, SearchOutline } from '@vicons/ionicons5'
import { useTabsStore } from '../stores/tabs'
import HistoryNav from './HistoryNav.vue'
import { t } from '../i18n'

const MAX_MATCHES = 2000

interface MatchRange {
  start: number
  end: number
}

const store = useTabsStore()
const mode = ref<'pretty' | 'raw'>('pretty')
const searchQuery = ref('')
const activeMatch = ref(0)
const preRef = ref<HTMLElement | null>(null)

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

const displayText = computed(() => (mode.value === 'pretty' ? prettyText.value : rawText.value))

function escapeRegExp(value: string): string {
  return value.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
}

function escapeHtml(value: string): string {
  return value
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#39;')
}

const matches = computed<MatchRange[]>(() => {
  const query = searchQuery.value
  const text = displayText.value
  if (!query || !text) return []
  const pattern = new RegExp(escapeRegExp(query), 'gi')
  const ranges: MatchRange[] = []
  for (const match of text.matchAll(pattern)) {
    if (match.index === undefined) continue
    ranges.push({ start: match.index, end: match.index + match[0].length })
    if (ranges.length >= MAX_MATCHES) break
  }
  return ranges
})

const highlightedHtml = computed(() => {
  const text = displayText.value
  if (!text) return ''
  const ranges = matches.value
  if (!ranges.length) return escapeHtml(text)

  const parts: string[] = []
  let position = 0
  ranges.forEach((range, index) => {
    if (range.start > position) {
      parts.push(escapeHtml(text.slice(position, range.start)))
    }
    const active = index === activeMatch.value ? ' class="active"' : ''
    parts.push(
      `<mark data-match="${index}"${active}>${escapeHtml(text.slice(range.start, range.end))}</mark>`,
    )
    position = range.end
  })
  if (position < text.length) {
    parts.push(escapeHtml(text.slice(position)))
  }
  return parts.join('')
})

const matchLabel = computed(() => {
  if (!searchQuery.value) return ''
  const total = matches.value.length
  if (!total) return t('response.noMatch')
  const totalText = total >= MAX_MATCHES ? `${MAX_MATCHES}+` : String(total)
  return `${activeMatch.value + 1}/${totalText}`
})

function scrollToMatch(smooth = false) {
  void nextTick(() => {
    const marks = preRef.value?.querySelectorAll('mark[data-match]')
    const target = marks?.[activeMatch.value]
    target?.scrollIntoView({
      block: 'center',
      inline: 'nearest',
      behavior: smooth ? 'smooth' : 'auto',
    })
  })
}

function goMatch(delta: number) {
  const total = matches.value.length
  if (!total) return
  activeMatch.value = (activeMatch.value + delta + total) % total
  scrollToMatch(true)
}

function onSearchKeydown(event: KeyboardEvent) {
  if (event.key === 'Enter') {
    event.preventDefault()
    goMatch(event.shiftKey ? -1 : 1)
  } else if (event.key === 'Escape') {
    searchQuery.value = ''
  }
}

watch(
  [searchQuery, () => mode.value, () => historyItem.value?.id],
  () => {
    activeMatch.value = 0
    if (searchQuery.value && matches.value.length) {
      scrollToMatch()
    }
  },
)

watch(
  () => matches.value.length,
  (total) => {
    if (activeMatch.value >= total) activeMatch.value = 0
  },
)
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

    <div v-if="response" class="search-bar">
      <n-input
        v-model:value="searchQuery"
        size="tiny"
        clearable
        :placeholder="t('response.searchPlaceholder')"
        class="search-input"
        @keydown="onSearchKeydown"
      >
        <template #prefix>
          <n-icon :component="SearchOutline" :size="13" />
        </template>
      </n-input>
      <span v-if="searchQuery" class="match-count mono" :class="{ none: !matches.length }">
        {{ matchLabel }}
      </span>
      <n-button
        text
        size="tiny"
        :disabled="!matches.length"
        :title="t('response.prevMatch')"
        @click="goMatch(-1)"
      >
        <template #icon>
          <n-icon :component="ChevronUpOutline" />
        </template>
      </n-button>
      <n-button
        text
        size="tiny"
        :disabled="!matches.length"
        :title="t('response.nextMatch')"
        @click="goMatch(1)"
      >
        <template #icon>
          <n-icon :component="ChevronDownOutline" />
        </template>
      </n-button>
    </div>

    <div class="panel-body">
      <div v-if="!response" class="panel-empty mono">{{ t('response.empty') }}</div>
      <pre
        v-else
        ref="preRef"
        class="response-pre mono"
        :class="{ 'error-text': response.error }"
        v-html="highlightedHtml"
      ></pre>
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

.search-bar {
  flex: none;
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 4px 12px;
  border-bottom: 1px solid var(--border-soft);
  background: var(--bar-bg);
}

.search-input {
  width: 220px;
  max-width: 45%;
}

.match-count {
  min-width: 52px;
  text-align: right;
  font-size: 11px;
  color: var(--text-muted);
}

.match-count.none {
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

.response-pre :deep(mark) {
  padding: 0;
  border-radius: 2px;
  background: var(--search-hit);
  color: inherit;
}

.response-pre :deep(mark.active) {
  background: var(--search-hit-active);
  color: var(--search-hit-active-text);
  box-shadow: 0 0 0 1px var(--search-hit-active);
}

.error-text {
  color: var(--error);
}
</style>
