<script lang="ts" setup>
import { computed, nextTick, onBeforeUnmount, ref } from 'vue'
import {
  NAlert,
  NButton,
  NCollapse,
  NCollapseItem,
  NIcon,
  NInput,
  useMessage,
} from 'naive-ui'
import {
  PaperPlaneOutline,
  PersonOutline,
  SparklesOutline,
  StopOutline,
} from '@vicons/ionicons5'
import { useSettingsStore } from '../../stores/settings'
import { useTabsStore } from '../../stores/tabs'
import { useUiStore } from '../../stores/ui'
import { lang, t } from '../../i18n'
import { cancelLLMChat, llmChat, parseRaw } from '../../api/repeater'
import { copyText } from '../../utils/clipboard'
import { extractHttpRequest } from '../../utils/httpRequest'
import type { ChatMessage } from '../../types'

interface UiMessage extends ChatMessage {
  id: number
}

const settings = useSettingsStore()
const tabs = useTabsStore()
const ui = useUiStore()
const message = useMessage()

const messages = ref<UiMessage[]>([])
const input = ref('')
const sending = ref(false)
const cancelling = ref(false)
const cancelRequested = ref(false)
const listRef = ref<HTMLDivElement | null>(null)
let idCounter = 0

const currentRequest = computed(() => tabs.activeTab?.requestText ?? '')
const currentTarget = computed(() => {
  const target = tabs.activeTab?.target
  if (!target?.host) return ''
  const port = target.port > 0 ? `:${target.port}` : ''
  return `${target.scheme}://${target.host}${port}`
})

const suggestions = computed(() => [
  t('ai.suggestJsonPost'),
  t('ai.suggestPayload'),
  t('ai.suggestExplain'),
])

function nextId(): number {
  idCounter += 1
  return idCounter
}

function scrollToBottom() {
  void nextTick(() => {
    const el = listRef.value
    if (el) el.scrollTop = el.scrollHeight
  })
}

function hasApplicableRequest(item: UiMessage): boolean {
  return extractHttpRequest(item.content) !== null
}

function openSettings() {
  ui.openSettings()
}

function clearChat() {
  messages.value = []
}

function useSuggestion(text: string) {
  input.value = text
}

function onKeydown(event: KeyboardEvent) {
  const modifier = event.metaKey || event.ctrlKey
  if (event.key !== 'Enter' || !modifier || event.shiftKey || event.altKey) return
  if (event.isComposing || event.keyCode === 229) return
  event.preventDefault()
  void send()
}

async function send() {
  const text = input.value.trim()
  if (!text || sending.value) return
  if (!settings.loaded) await settings.load()
  if (!settings.llmConfigured) {
    message.warning(t('ai.notConfigured'))
    return
  }

  const tab = tabs.activeTab
  messages.value.push({ id: nextId(), role: 'user', content: text })
  input.value = ''
  scrollToBottom()

  const history: ChatMessage[] = messages.value.map((item) => ({
    role: item.role,
    content: item.content,
  }))

  sending.value = true
  cancelRequested.value = false
  cancelling.value = false
  try {
    const res = await llmChat({
      messages: history,
      requestText: tab?.requestText ?? '',
      scheme: tab?.target.scheme ?? 'http',
      host: tab?.target.host ?? '',
      port: tab?.target.port ?? 0,
      lang: lang.value,
    })
    messages.value.push({ id: nextId(), role: 'assistant', content: res.content })
  } catch (e) {
    if (cancelRequested.value) {
      message.info(t('ai.cancelled'))
    } else {
      const msg = e instanceof Error ? e.message : String(e)
      message.error(t('ai.error', { msg }))
    }
  } finally {
    sending.value = false
    cancelling.value = false
    cancelRequested.value = false
    scrollToBottom()
  }
}

async function cancelSend() {
  if (!sending.value || cancelling.value) return
  cancelling.value = true
  cancelRequested.value = true
  try {
    await cancelLLMChat()
  } catch {
    // ignore
  }
}

async function applyMessage(item: UiMessage) {
  const text = extractHttpRequest(item.content)
  if (!text) {
    message.warning(t('ai.noRequest'))
    return
  }
  try {
    const parsed = await parseRaw(text, lang.value)
    const current = tabs.activeTab?.target
    if (parsed.target.host) {
      if (current && current.host === parsed.target.host && parsed.target.port === 0) {
        tabs.updateTarget({ scheme: current.scheme, host: parsed.target.host, port: current.port })
      } else {
        tabs.updateTarget({
          scheme: parsed.target.scheme,
          host: parsed.target.host,
          port: parsed.target.port,
        })
      }
    }
    tabs.setRequestText(text)
    message.success(t('ai.applied'))
  } catch (e) {
    const msg = e instanceof Error ? e.message : String(e)
    message.error(t('ai.applyFailed', { msg }))
  }
}

async function copyMessage(item: UiMessage) {
  try {
    await copyText(item.content)
    message.success(t('ai.copied'))
  } catch (e) {
    const msg = e instanceof Error ? e.message : String(e)
    message.error(t('ctx.copyFailed', { msg }))
  }
}

onBeforeUnmount(() => {
  if (sending.value) void cancelLLMChat()
})
</script>

<template>
  <div class="ai-panel">
    <div class="ai-head">
      <span class="ai-title">{{ t('ai.title') }}</span>
      <span v-if="settings.llm.model" class="ai-model mono">{{ settings.llm.model }}</span>
    </div>
    <p class="ai-subtitle">{{ t('ai.subtitle') }}</p>

    <n-alert v-if="!settings.llmConfigured" type="warning" :bordered="false" class="alert">
      <div class="alert-row">
        <span>{{ t('ai.notConfigured') }}</span>
        <n-button size="tiny" type="primary" secondary @click="openSettings">
          {{ t('ai.openSettings') }}
        </n-button>
      </div>
    </n-alert>

    <n-collapse v-if="currentRequest" class="context">
      <n-collapse-item :title="t('ai.currentRequest')" name="request">
        <div v-if="currentTarget" class="context-target mono">{{ currentTarget }}</div>
        <pre class="context-pre mono">{{ currentRequest }}</pre>
      </n-collapse-item>
    </n-collapse>

    <div ref="listRef" class="chat-list">
      <div v-if="!messages.length" class="chat-empty">
        <div class="empty-icon">
          <n-icon :component="SparklesOutline" :size="20" />
        </div>
        <div class="empty-text">{{ t('ai.empty') }}</div>
        <div class="empty-chips">
          <button
            v-for="suggestion in suggestions"
            :key="suggestion"
            class="suggest-chip"
            type="button"
            @click="useSuggestion(suggestion)"
          >
            {{ suggestion }}
          </button>
        </div>
      </div>

      <div v-for="item in messages" :key="item.id" class="chat-item" :class="item.role">
        <div class="chat-avatar" :class="item.role">
          <n-icon :component="item.role === 'user' ? PersonOutline : SparklesOutline" :size="12" />
        </div>
        <div class="chat-main">
          <div class="chat-role">{{ item.role === 'user' ? t('ai.you') : t('ai.assistant') }}</div>
          <div class="chat-bubble">{{ item.content }}</div>
          <div v-if="item.role === 'assistant'" class="chat-actions">
            <n-button
              v-if="hasApplicableRequest(item)"
              text
              size="tiny"
              type="primary"
              @click="applyMessage(item)"
            >
              {{ t('ai.apply') }}
            </n-button>
            <n-button text size="tiny" @click="copyMessage(item)">{{ t('ai.copy') }}</n-button>
          </div>
        </div>
      </div>

      <div v-if="sending" class="chat-item assistant">
        <div class="chat-avatar assistant">
          <n-icon :component="SparklesOutline" :size="12" />
        </div>
        <div class="chat-main">
          <div class="chat-role">{{ t('ai.assistant') }}</div>
          <div class="chat-bubble thinking">
            <span class="thinking-dot"></span>
            <span class="thinking-dot"></span>
            <span class="thinking-dot"></span>
            <span class="thinking-text">{{ t('ai.thinking') }}</span>
          </div>
        </div>
      </div>
    </div>

    <div class="composer">
      <n-input
        :value="input"
        type="textarea"
        :placeholder="t('ai.placeholder')"
        :autosize="{ minRows: 2, maxRows: 6 }"
        :disabled="sending"
        @update:value="(value: string) => (input = value)"
        @keydown="onKeydown"
      />
      <div class="composer-bar">
        <span class="hint">{{ t('ai.inputHint') }}</span>
        <n-button
          v-if="sending"
          type="error"
          secondary
          size="small"
          :loading="cancelling"
          @click="cancelSend"
        >
          <template #icon>
            <n-icon :component="StopOutline" />
          </template>
          {{ t('ai.stop') }}
        </n-button>
        <n-button
          v-else
          type="primary"
          size="small"
          :disabled="!settings.llmConfigured || !input.trim()"
          @click="send"
        >
          <template #icon>
            <n-icon :component="PaperPlaneOutline" />
          </template>
          {{ t('ai.send') }}
        </n-button>
      </div>
    </div>

    <div class="ai-footer">
      <n-button quaternary size="tiny" :disabled="!messages.length || sending" @click="clearChat">
        {{ t('ai.clear') }}
      </n-button>
      <n-button quaternary size="tiny" @click="openSettings">{{ t('ai.config') }}</n-button>
      <span class="hint footer-hint">{{ t('ai.requestHint') }}</span>
    </div>
  </div>
</template>

<style scoped>
.ai-panel {
  height: 100%;
  min-height: 0;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.ai-head {
  flex: none;
  display: flex;
  align-items: baseline;
  gap: 8px;
  min-width: 0;
}

.ai-title {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-strong);
}

.ai-model {
  padding: 1px 6px;
  border: 1px solid var(--border);
  border-radius: 9px;
  font-size: 10px;
  color: var(--text-muted);
}

.ai-subtitle {
  flex: none;
  margin: 0;
  font-size: 11px;
  line-height: 1.4;
  color: var(--text-dim);
}

.alert {
  flex: none;
}

.alert-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.context {
  flex: none;
}

.context-target {
  margin-bottom: 4px;
  font-size: 11px;
  color: var(--accent-soft);
}

.context-pre {
  max-height: 120px;
  margin: 0;
  padding: 6px 8px;
  overflow: auto;
  border: 1px solid var(--border-soft);
  border-radius: 6px;
  background: var(--input-bg);
  font-size: 11px;
  line-height: 1.45;
  white-space: pre-wrap;
}

.chat-list {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 8px;
  border: 1px solid var(--border-soft);
  border-radius: 8px;
  background: var(--input-bg);
}

.chat-empty {
  margin: auto;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
  padding: 8px;
  text-align: center;
}

.empty-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  border: 1px solid var(--accent-border);
  border-radius: 50%;
  color: var(--accent);
}

.empty-text {
  font-size: 11.5px;
  line-height: 1.6;
  color: var(--text-muted);
}

.empty-chips {
  display: flex;
  flex-wrap: wrap;
  justify-content: center;
  gap: 6px;
}

.suggest-chip {
  padding: 3px 8px;
  border: 1px solid var(--border);
  border-radius: 12px;
  background: transparent;
  color: var(--text-muted);
  font-size: 10.5px;
  font-family: inherit;
  cursor: pointer;
}

.suggest-chip:hover {
  border-color: var(--accent-border);
  background: var(--hover-bg-soft);
}

.chat-item {
  display: flex;
  gap: 6px;
}

.chat-item.user {
  flex-direction: row-reverse;
}

.chat-avatar {
  flex: none;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 22px;
  height: 22px;
  margin-top: 14px;
  border: 1px solid var(--border);
  border-radius: 50%;
  color: var(--text-muted);
}

.chat-avatar.user {
  border-color: var(--accent-border);
  color: var(--accent);
}

.chat-main {
  display: flex;
  flex-direction: column;
  gap: 3px;
  min-width: 0;
  max-width: calc(100% - 30px);
}

.chat-item.user .chat-main {
  align-items: flex-end;
}

.chat-role {
  font-size: 10px;
  color: var(--text-dim);
}

.chat-bubble {
  padding: 6px 9px;
  border: 1px solid var(--border-soft);
  border-radius: 8px;
  background: var(--bar-bg);
  font-size: 12px;
  line-height: 1.55;
  white-space: pre-wrap;
  word-break: break-word;
  user-select: text;
}

.chat-item.user .chat-bubble {
  border-color: var(--accent-border);
  background: var(--accent-border);
}

.chat-bubble.thinking {
  display: flex;
  align-items: center;
  gap: 4px;
  color: var(--text-muted);
}

.thinking-dot {
  width: 4px;
  height: 4px;
  border-radius: 50%;
  background: var(--accent);
  animation: thinking-bounce 1.1s infinite ease-in-out;
}

.thinking-dot:nth-child(2) {
  animation-delay: 0.15s;
}

.thinking-dot:nth-child(3) {
  animation-delay: 0.3s;
}

@keyframes thinking-bounce {
  0%,
  80%,
  100% {
    opacity: 0.25;
  }
  40% {
    opacity: 1;
  }
}

.chat-actions {
  display: flex;
  gap: 8px;
}

.composer {
  flex: none;
  display: flex;
  flex-direction: column;
  gap: 6px;
  padding: 6px;
  border: 1px solid var(--border);
  border-radius: 8px;
  background: var(--bar-bg);
}

.composer-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.hint {
  font-size: 10.5px;
  color: var(--text-dim);
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.ai-footer {
  flex: none;
  display: flex;
  align-items: center;
  gap: 6px;
  flex-wrap: wrap;
}

.footer-hint {
  margin-left: auto;
}
</style>
