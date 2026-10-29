<script lang="ts" setup>
import { computed, onBeforeUnmount, ref, watch } from 'vue'
import { NButton, NDropdown, NIcon, useMessage } from 'naive-ui'
import type { DropdownDividerOption, DropdownOption } from 'naive-ui'
import {
  ChatbubbleEllipsesOutline,
  ChevronDownOutline,
  DocumentTextOutline,
  DownloadOutline,
} from '@vicons/ionicons5'
import { ClipboardGetText } from '../../wailsjs/runtime/runtime'
import { useTabsStore } from '../stores/tabs'
import { useUiStore } from '../stores/ui'
import { builtinTemplates } from '../utils/templates'
import {
  buildRequestCurl,
  buildRequestURL,
  convertRequestMethod,
  importRequest,
  parseRaw,
} from '../api/repeater'
import { copyText } from '../utils/clipboard'
import { lang, t } from '../i18n'
import type { RequestToolInput } from '../types'
import RequestEditor from './RequestEditor.vue'
import HistoryNav from './HistoryNav.vue'

const emit = defineEmits<{ (e: 'send'): void }>()

const store = useTabsStore()
const ui = useUiStore()
const message = useMessage()
const showDropdown = ref(false)

const dropdownOptions = computed(() =>
  builtinTemplates.map((tpl) => ({
    key: tpl.key,
    label: t(tpl.labelKey),
  })),
)

function onTemplateSelect(key: string | number) {
  const tpl = builtinTemplates.find((t) => t.key === key)
  if (tpl) {
    store.setRequestText(tpl.build()) // 模板插入 = 整体替换
    // Host 留空模板联动关闭「Host 跟随」，否则发送时会自动补 Host
    if (tpl.disableFollowHost) store.setFollowHostHeader(false)
  }
  showDropdown.value = false
}

// ---- curl / fetch 导入 ----

function looksLikeImport(text: string): boolean {
  return /^\s*(?:[$>»]\s*)?curl(?:\.exe)?\b/i.test(text) || /\bfetch\s*\(/.test(text)
}

async function applyImport(text: string): Promise<{ error?: string }> {
  try {
    const res = await importRequest(text, lang.value)
    store.setRequestText(res.requestText)
    store.updateTarget({
      scheme: res.target.scheme,
      host: res.target.host,
      port: res.target.port,
    })
    // 报文自带 Host 与 URL 不一致时，关闭“Host 跟随”，否则发送会被改写
    if (res.hostHeaderMismatch) {
      store.setFollowHostHeader(false)
      message.warning(t('request.importedHostMismatch'))
    }
    if (res.warnings?.length) {
      message.warning(res.warnings.join(lang.value === 'zh' ? '；' : '; '), { duration: 6000 })
    }
    message.success(t('request.imported', { source: res.source }))
    return {}
  } catch (e) {
    return {
      error: t('request.importFailed', {
        msg: e instanceof Error ? e.message : String(e),
      }),
    }
  }
}

// 模板旁的「导入」按钮：读剪贴板（Wails runtime 优先，浏览器调试回退 Web API）
async function importFromClipboard() {
  let text = ''
  try {
    text = await ClipboardGetText()
  } catch {
    // 无 Wails runtime 时忽略
  }
  if (!text.trim()) {
    try {
      text = await navigator.clipboard.readText()
    } catch {
      // 统一在下方提示
    }
  }
  if (!text.trim()) {
    message.warning(t('request.clipboardEmpty'))
    return
  }
  const { error } = await applyImport(text)
  if (error) message.error(error)
}

// 粘贴：curl / fetch 自动转换；否则按原逻辑解析目标（裸 URL 转 GET 骨架）
async function onPaste(fullText: string) {
  if (looksLikeImport(fullText)) {
    const { error } = await applyImport(fullText)
    if (error) message.error(error)
    return
  }
  try {
    const parsed = await parseRaw(fullText, lang.value)
    if (parsed.isBareUrl && parsed.skeleton) {
      store.setRequestText(parsed.skeleton)
    }
    if (parsed.target.host) {
      store.updateTarget({
        scheme: parsed.target.scheme,
        host: parsed.target.host,
        port: parsed.target.port,
      })
    }
  } catch {
    // 自动解析失败不阻塞输入（手动点"解析"时再提示）
  }
}

// ---- 编辑器右键菜单：复制 URL / 复制为 curl / GET↔POST ----

const ctxShow = ref(false)
const ctxX = ref(0)
const ctxY = ref(0)

const ctxMethod = computed(() => {
  const text = store.activeTab?.requestText.trim() ?? ''
  return text ? text.split(/\s+/)[0].toUpperCase() : ''
})

const ctxHasRequest = computed(() => !!store.activeTab?.requestText.trim())

const ctxOptions = computed<Array<DropdownOption | DropdownDividerOption>>(() => {
  const method = ctxMethod.value
  const convertible = method === 'GET' || method === 'POST'
  return [
    { label: t('ctx.copyUrl'), key: 'copy-url', disabled: !ctxHasRequest.value },
    { label: t('ctx.copyCurl'), key: 'copy-curl', disabled: !ctxHasRequest.value },
    { type: 'divider', key: 'ctx-divider' },
    {
      label:
        method === 'GET'
          ? t('ctx.convertToPost')
          : method === 'POST'
            ? t('ctx.convertToGet')
            : t('ctx.convertUnsupported'),
      key: 'convert-method',
      disabled: !ctxHasRequest.value || !convertible,
    },
  ]
})

function onContextMenu(pos: { x: number; y: number }) {
  ctxX.value = pos.x
  ctxY.value = pos.y
  ctxShow.value = true
}

// manual dropdown 不触发 clickoutside，这里用 document 监听保证点外部/Esc 能关闭
function onDocumentPointerDown(event: PointerEvent) {
  for (const node of event.composedPath()) {
    if (
      node instanceof HTMLElement &&
      (node.classList.contains('ctx-menu') || node.classList.contains('n-dropdown-menu'))
    ) {
      return
    }
  }
  ctxShow.value = false
}

function onDocumentKeydown(event: KeyboardEvent) {
  if (event.key === 'Escape') ctxShow.value = false
}

watch(ctxShow, (show) => {
  if (show) {
    document.addEventListener('pointerdown', onDocumentPointerDown, true)
    document.addEventListener('keydown', onDocumentKeydown, true)
  } else {
    document.removeEventListener('pointerdown', onDocumentPointerDown, true)
    document.removeEventListener('keydown', onDocumentKeydown, true)
  }
})

onBeforeUnmount(() => {
  document.removeEventListener('pointerdown', onDocumentPointerDown, true)
  document.removeEventListener('keydown', onDocumentKeydown, true)
})

function currentToolInput(): RequestToolInput | null {
  const tab = store.activeTab
  if (!tab) return null
  return {
    requestText: tab.requestText,
    scheme: tab.target.scheme,
    host: tab.target.host,
    port: tab.target.port,
    lang: lang.value,
  }
}

async function onContextSelect(key: string | number) {
  ctxShow.value = false
  const input = currentToolInput()
  if (!input) return
  try {
    if (key === 'copy-url') {
      const res = await buildRequestURL(input)
      await copyText(res.text)
      message.success(t('ctx.copyUrlDone'))
    } else if (key === 'copy-curl') {
      const res = await buildRequestCurl(input)
      await copyText(res.text)
      message.success(t('ctx.copyCurlDone'))
    } else if (key === 'convert-method') {
      const from = ctxMethod.value
      const res = await convertRequestMethod(input)
      store.setRequestText(res.text)
      message.success(t('ctx.converted', { method: from === 'GET' ? 'POST' : 'GET' }))
    }
  } catch (e) {
    const msg = e instanceof Error ? e.message : String(e)
    if (key === 'convert-method') {
      message.error(t('ctx.convertFailed', { msg }))
    } else {
      message.error(t('ctx.copyFailed', { msg }))
    }
  }
}
</script>

<template>
  <div class="panel">
    <div class="panel-header">
      <span class="panel-title">{{ t('request.title') }}</span>

      <n-dropdown
        trigger="click"
        :show="showDropdown"
        :options="dropdownOptions"
        placement="bottom-start"
        @update:show="(v: boolean) => (showDropdown = v)"
        @select="onTemplateSelect"
      >
        <n-button size="tiny" quaternary class="tpl-btn">
          <template #icon>
            <n-icon :component="DocumentTextOutline" />
          </template>
          {{ t('request.templates') }}
          <n-icon :component="ChevronDownOutline" :size="12" />
        </n-button>
      </n-dropdown>

      <n-button
        size="tiny"
        quaternary
        class="tpl-btn"
        :title="t('request.importTitle')"
        @click="importFromClipboard"
      >
        <template #icon>
          <n-icon :component="DownloadOutline" />
        </template>
        {{ t('request.import') }}
      </n-button>

      <n-button
        size="tiny"
        quaternary
        class="tpl-btn"
        :title="t('request.aiTitle')"
        @click="ui.openAi()"
      >
        <template #icon>
          <n-icon :component="ChatbubbleEllipsesOutline" />
        </template>
        {{ t('request.ai') }}
      </n-button>

      <div class="header-spacer"></div>

      <HistoryNav />
    </div>

    <div class="panel-body">
      <RequestEditor
        @send="emit('send')"
        @paste="onPaste"
        @contextmenu="onContextMenu"
      />
    </div>

    <n-dropdown
      trigger="manual"
      placement="bottom-start"
      class="ctx-menu"
      :show="ctxShow"
      :x="ctxX"
      :y="ctxY"
      :options="ctxOptions"
      @select="onContextSelect"
    />
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
  gap: 8px;
  padding: 0 12px;
  border-bottom: 1px solid var(--border-soft);
  background: var(--bar-bg);
}

.panel-title {
  font-size: 12.5px;
  font-weight: 600;
  color: var(--text);
  letter-spacing: 1px;
}

.tpl-btn {
  font-size: 12px;
}

.header-spacer {
  flex: 1;
}

.panel-body {
  flex: 1;
  min-height: 0;
  overflow: hidden;
}
</style>
