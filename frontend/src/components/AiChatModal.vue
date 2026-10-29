<script lang="ts" setup>
import { computed, nextTick, onBeforeUnmount, reactive, ref } from 'vue'
import {
  NAlert,
  NButton,
  NCollapse,
  NCollapseItem,
  NIcon,
  NInput,
  NModal,
  useMessage,
} from 'naive-ui'
import {
  PaperPlaneOutline,
  PersonOutline,
  SparklesOutline,
  StopOutline,
} from '@vicons/ionicons5'
import { useSettingsStore } from '../stores/settings'
import { useTabsStore } from '../stores/tabs'
import { useUiStore } from '../stores/ui'
import { lang, t } from '../i18n'
import { cancelLLMChat, llmChat, parseRaw } from '../api/repeater'
import { copyText } from '../utils/clipboard'
import { extractHttpRequest } from '../utils/httpRequest'
import type { ChatMessage } from '../types'

interface UiMessage extends ChatMessage {
  id: number
}

const AI_SIZE_KEY = 'httpRepeater.aiModalSize'
const DEFAULT_WIDTH = 640
const DEFAULT_HEIGHT = 620

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

function clamp(value: number, min: number, max: number): number {
  return Math.min(max, Math.max(min, value))
}

function readStoredSize(): { width: number; height: number } {
  try {
    const raw = localStorage.getItem(AI_SIZE_KEY)
    if (!raw) return { width: DEFAULT_WIDTH, height: DEFAULT_HEIGHT }
    const parsed = JSON.parse(raw) as { width?: number; height?: number }
    return {
      width: clamp(Number(parsed.width) || DEFAULT_WIDTH, 460, 1400),
      height: clamp(Number(parsed.height) || DEFAULT_HEIGHT, 400, 1200),
    }
  } catch {
    return { width: DEFAULT_WIDTH, height: DEFAULT_HEIGHT }
  }
}

const modalSize = reactive(readStoredSize())

const modalStyle = computed(
  () =>
    `width: ${modalSize.width}px; height: ${modalSize.height}px;` +
    ' max-width: calc(100vw - 40px); max-height: calc(100vh - 40px);' +
    ' display: flex; flex-direction: column; overflow: hidden;',
)

const modalContentStyle =
  'flex: 1 1 auto; min-height: 0; display: flex; flex-direction: column;'

// ---- 拖动右下角调整对话框大小 ----
let resizing = false
let resizeStartX = 0
let resizeStartY = 0
let resizeStartW = 0
let resizeStartH = 0

function onResizeMove(event: PointerEvent) {
  if (!resizing) return
  const maxW = Math.min(1400, window.innerWidth - 40)
  const maxH = window.innerHeight - 40
  // 弹层居中，左右/上下对称伸缩，所以鼠标位移要乘 2
  const nextW = resizeStartW + (event.clientX - resizeStartX) * 2
  const nextH = resizeStartH + (event.clientY - resizeStartY) * 2
  modalSize.width = Math.round(clamp(nextW, 460, maxW))
  modalSize.height = Math.round(clamp(nextH, 400, maxH))
}

function onResizeEnd() {
  if (!resizing) return
  resizing = false
  window.removeEventListener('pointermove', onResizeMove)
  window.removeEventListener('pointerup', onResizeEnd)
  document.body.style.cursor = ''
  try {
    localStorage.setItem(
      AI_SIZE_KEY,
      JSON.stringify({ width: modalSize.width, height: modalSize.height }),
    )
  } catch {
    // 存储不可用不影响本次会话
  }
}

function onResizeStart(event: PointerEvent) {
  event.preventDefault()
  resizing = true
  resizeStartX = event.clientX
  resizeStartY = event.clientY
  resizeStartW = modalSize.width
  resizeStartH = modalSize.height
  document.body.style.cursor = 'nwse-resize'
  window.addEventListener('pointermove', onResizeMove)
  window.addEventListener('pointerup', onResizeEnd)
}

onBeforeUnmount(() => {
  window.removeEventListener('pointermove', onResizeMove)
  window.removeEventListener('pointerup', onResizeEnd)
  document.body.style.cursor = ''
})

// ---- 对话 ----
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

function close() {
  ui.closeAi()
}

function openSettings() {
  ui.closeAi()
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
  // 输入法组合过程中的 Enter 只用于选词
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
    // 请求可能已经结束，忽略取消失败
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
      // 原始报文通常只带 Host（没有 scheme）；同 Host 且未写端口时沿用当前目标，
      // 避免把 https 目标误判成 http。
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
</script>

<template>
  <n-modal
    :show="ui.aiOpen"
    preset="card"
    class="ai-modal"
    :bordered="false"
    :mask-closable="false"
    :auto-focus="false"
    :style="modalStyle"
    :content-style="modalContentStyle"
    @update:show="(value: boolean) => !value && close()"
  >
    <template #header>
      <div class="ai-header">
        <span class="ai-title">{{ t('ai.title') }}</span>
        <span v-if="settings.llm.model" class="ai-model mono">{{ settings.llm.model }}</span>
        <span class="ai-subtitle">{{ t('ai.subtitle') }}</span>
      </div>
    </template>

    <div class="ai-body">
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
            <n-icon :component="SparklesOutline" :size="22" />
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

        <div
          v-for="item in messages"
          :key="item.id"
          class="chat-item"
          :class="item.role"
        >
          <div class="chat-avatar" :class="item.role">
            <n-icon
              :component="item.role === 'user' ? PersonOutline : SparklesOutline"
              :size="13"
            />
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
            <n-icon :component="SparklesOutline" :size="13" />
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
          :autosize="{ minRows: 2, maxRows: 7 }"
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
    </div>

    <template #footer>
      <div class="ai-footer">
        <n-button quaternary size="tiny" :disabled="!messages.length || sending" @click="clearChat">
          {{ t('ai.clear') }}
        </n-button>
        <n-button quaternary size="tiny" @click="openSettings">{{ t('ai.config') }}</n-button>
        <div class="footer-spacer"></div>
        <span class="hint footer-hint">{{ t('ai.requestHint') }}</span>
        <span class="resize-grip" :title="t('ai.resize')" @pointerdown="onResizeStart"></span>
        <n-button size="tiny" @click="close">{{ t('ai.close') }}</n-button>
      </div>
    </template>
  </n-modal>
</template>

<style scoped>
.ai-header {
  display: flex;
  align-items: baseline;
  gap: 8px;
  min-width: 0;
}

.ai-title {
  flex: none;
  font-size: 14px;
  font-weight: 600;
  color: var(--text-strong);
}

.ai-model {
  flex: none;
  padding: 1px 6px;
  border: 1px solid var(--border);
  border-radius: 9px;
  font-size: 10.5px;
  line-height: 15px;
  color: var(--text-muted);
}

.ai-subtitle {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-size: 11.5px;
  color: var(--text-dim);
}

.ai-body {
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.alert {
  flex: none;
}

.alert-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
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
  max-height: 160px;
  margin: 0;
  padding: 8px 10px;
  overflow: auto;
  border: 1px solid var(--border-soft);
  border-radius: 6px;
  background: var(--input-bg);
  font-size: 11.5px;
  line-height: 1.5;
  color: var(--text-body);
  white-space: pre;
}

.chat-list {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  display: flex;
  flex-direction: column;
  gap: 14px;
  padding: 12px;
  border: 1px solid var(--border-soft);
  border-radius: 10px;
  background: var(--input-bg);
}

.chat-empty {
  margin: auto;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 10px;
  max-width: 440px;
  padding: 12px;
  text-align: center;
}

.empty-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 44px;
  height: 44px;
  border: 1px solid var(--accent-border);
  border-radius: 50%;
  background: var(--accent-border);
  color: var(--accent);
}

.empty-text {
  font-size: 12px;
  line-height: 1.7;
  color: var(--text-muted);
}

.empty-chips {
  display: flex;
  flex-wrap: wrap;
  justify-content: center;
  gap: 6px;
}

.suggest-chip {
  padding: 4px 9px;
  border: 1px solid var(--border);
  border-radius: 12px;
  background: transparent;
  color: var(--text-muted);
  font-size: 11px;
  font-family: inherit;
  cursor: pointer;
  transition: background-color 0.15s ease, color 0.15s ease, border-color 0.15s ease;
}

.suggest-chip:hover {
  border-color: var(--accent-border);
  background: var(--hover-bg-soft);
  color: var(--text);
}

.chat-item {
  display: flex;
  gap: 8px;
  max-width: 100%;
}

.chat-item.user {
  flex-direction: row-reverse;
}

.chat-avatar {
  flex: none;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  margin-top: 16px;
  border: 1px solid var(--border);
  border-radius: 50%;
  background: var(--bar-bg);
  color: var(--text-muted);
}

.chat-avatar.user {
  border-color: var(--accent-border);
  background: var(--accent-border);
  color: var(--accent);
}

.chat-avatar.assistant {
  color: var(--accent);
}

.chat-main {
  display: flex;
  flex-direction: column;
  gap: 4px;
  min-width: 0;
  max-width: calc(100% - 34px);
}

.chat-item.user .chat-main {
  align-items: flex-end;
}

.chat-role {
  font-size: 10.5px;
  color: var(--text-dim);
}

.chat-bubble {
  padding: 8px 11px;
  border: 1px solid var(--border-soft);
  border-radius: 10px;
  border-top-left-radius: 3px;
  background: var(--bar-bg);
  color: var(--text-body);
  font-size: 12.5px;
  line-height: 1.65;
  white-space: pre-wrap;
  word-break: break-word;
  user-select: text;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.08);
}

.chat-item.user .chat-bubble {
  border-color: var(--accent-border);
  border-top-left-radius: 10px;
  border-top-right-radius: 3px;
  background: var(--accent-border);
  color: var(--text-strong);
}

.chat-bubble.thinking {
  display: flex;
  align-items: center;
  gap: 4px;
  color: var(--text-muted);
}

.thinking-text {
  margin-left: 4px;
  font-size: 11.5px;
  font-style: italic;
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
    transform: translateY(0);
  }
  40% {
    opacity: 1;
    transform: translateY(-3px);
  }
}

.chat-actions {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 0 2px;
}

.composer {
  flex: none;
  display: flex;
  flex-direction: column;
  gap: 6px;
  padding: 8px;
  border: 1px solid var(--border);
  border-radius: 10px;
  background: var(--bar-bg);
}

.composer :deep(.n-input) {
  --n-color: transparent !important;
  --n-color-focus: transparent !important;
  --n-border: 1px solid transparent !important;
  --n-border-hover: 1px solid transparent !important;
  --n-border-focus: 1px solid transparent !important;
  --n-box-shadow-focus: none !important;
}

.composer-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
}

.hint {
  min-width: 0;
  font-size: 11px;
  color: var(--text-dim);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.ai-footer {
  display: flex;
  align-items: center;
  gap: 6px;
  min-width: 0;
}

.footer-spacer {
  flex: 1;
}

.footer-hint {
  max-width: 55%;
}

.resize-grip {
  flex: none;
  width: 14px;
  height: 14px;
  cursor: nwse-resize;
  touch-action: none;
  opacity: 0.65;
  background-image: repeating-linear-gradient(
    135deg,
    var(--text-dim) 0,
    var(--text-dim) 1px,
    transparent 1px,
    transparent 4px
  );
  border-bottom-right-radius: 2px;
}

.resize-grip:hover {
  opacity: 1;
}
</style>
