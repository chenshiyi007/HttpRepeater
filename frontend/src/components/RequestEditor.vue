<script lang="ts" setup>
import { onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { Compartment, EditorState } from '@codemirror/state'
import { EditorView, keymap } from '@codemirror/view'
import { basicSetup } from 'codemirror'
import { oneDark } from '@codemirror/theme-one-dark'
import { httpLanguage } from '../utils/httpHighlight'
import { useTabsStore } from '../stores/tabs'
import { useSettingsStore, type ThemeMode } from '../stores/settings'

const emit = defineEmits<{
  (e: 'send'): void
  (e: 'paste', text: string): void
}>()

const store = useTabsStore()
const settings = useSettingsStore()
const container = ref<HTMLDivElement | null>(null)
let view: EditorView | null = null
let syncingFromStore = false

// 主题用 Compartment 承载，切换时只重配主题、不重建文档
const themeCompartment = new Compartment()

function buildTheme() {
  return EditorView.theme({
    '&': { height: '100%', fontSize: '12px', backgroundColor: 'var(--app-bg)' },
    '.cm-scroller': {
      fontFamily: "'SF Mono', Menlo, Monaco, Consolas, 'Courier New', monospace",
      lineHeight: '1.6',
      backgroundColor: 'var(--app-bg)',
    },
    '.cm-content': {
      padding: '8px 0',
      color: 'var(--text-body)',
      caretColor: 'var(--text-strong)',
    },
    '&.cm-focused': { outline: 'none' },
    '.cm-cursor, .cm-dropCursor': { borderLeftColor: 'var(--text-strong)' },
    '.cm-gutters': {
      backgroundColor: 'var(--app-bg)',
      color: 'var(--text-dim)',
      border: 'none',
    },
    '.cm-activeLine': { backgroundColor: 'var(--hover-bg-soft)' },
    '.cm-activeLineGutter': { backgroundColor: 'var(--hover-bg-soft)' },
  })
}

function themeExtension(mode: ThemeMode) {
  // oneDark 在深色时先加载，随后 buildTheme 覆盖背景/行号等细节
  return mode === 'dark' ? [oneDark, buildTheme()] : [buildTheme()]
}

function createState(doc: string): EditorState {
  return EditorState.create({
    doc,
    extensions: [
      basicSetup, // 行号、历史、搜索（⌘F）等
      themeCompartment.of(themeExtension(settings.theme)),
      httpLanguage,
      keymap.of([
        {
          key: 'Mod-Enter',
          run: () => {
            emit('send')
            return true
          },
        },
      ]),
      EditorView.updateListener.of((update) => {
        if (update.docChanged && !syncingFromStore) {
          store.setRequestText(update.state.doc.toString())
        }
      }),
      EditorView.domEventHandlers({
        paste: (event) => {
          // 优先用剪贴板原文（curl/fetch 识别更准）；拿不到再等 CodeMirror 应用后取全文
          const pasted = event.clipboardData?.getData('text') ?? ''
          setTimeout(() => {
            if (pasted) {
              emit('paste', pasted)
            } else if (view) {
              emit('paste', view.state.doc.toString())
            }
          }, 0)
          return false
        },
      }),
    ],
  })
}

onMounted(() => {
  if (!container.value) return
  view = new EditorView({
    state: createState(store.activeTab?.requestText ?? ''),
    parent: container.value,
  })
})

onBeforeUnmount(() => {
  view?.destroy()
  view = null
})

// 切换标签：整体替换编辑器内容（沿用当前主题）
watch(
  () => store.activeTabId,
  () => {
    if (!view) return
    syncingFromStore = true
    view.setState(createState(store.activeTab?.requestText ?? ''))
    syncingFromStore = false
  },
)

// 外部修改（模板插入、解析生成的骨架、历史切换）：同步到编辑器
watch(
  () => store.activeTab?.requestText ?? '',
  (text) => {
    if (!view || syncingFromStore) return
    if (text === view.state.doc.toString()) return
    syncingFromStore = true
    view.dispatch({
      changes: { from: 0, to: view.state.doc.length, insert: text },
    })
    syncingFromStore = false
  },
)

// 深浅色切换：只重配主题扩展，保留撤销栈与光标
watch(
  () => settings.theme,
  (mode) => {
    if (!view) return
    view.dispatch({ effects: themeCompartment.reconfigure(themeExtension(mode)) })
  },
)
</script>

<template>
  <div ref="container" class="request-editor"></div>
</template>

<style scoped>
.request-editor {
  height: 100%;
  overflow: hidden;
}

.request-editor :deep(.cm-editor) {
  height: 100%;
}
</style>
