<script lang="ts" setup>
import { computed, onBeforeUnmount, onMounted } from 'vue'
import { NSplit, useMessage } from 'naive-ui'
import TabBar from '../components/TabBar.vue'
import TargetBar from '../components/TargetBar.vue'
import RequestPanel from '../components/RequestPanel.vue'
import ResponsePanel from '../components/ResponsePanel.vue'
import { useTabsStore } from '../stores/tabs'
import { t } from '../i18n'

const store = useTabsStore()
const message = useMessage()

async function doSend() {
  const { error } = await store.sendActiveTab()
  if (error) message.error(error)
}

// ⌘↵ / Ctrl+↵ 全局发送；⌘T / ⌘W 新建、关闭标签（编辑器内的 keymap 已 preventDefault）
function onKeydown(e: KeyboardEvent) {
  const mod = e.metaKey || e.ctrlKey
  if (!mod) return
  if (e.key === 'Enter') {
    if (e.defaultPrevented) return
    e.preventDefault()
    void doSend()
    return
  }
  if (e.shiftKey || e.altKey) return
  const key = e.key.toLowerCase()
  if (key === 't') {
    e.preventDefault()
    store.addTab()
  } else if (key === 'w') {
    e.preventDefault()
    store.closeTab(store.activeTabId)
  }
}

onMounted(() => window.addEventListener('keydown', onKeydown))
onBeforeUnmount(() => window.removeEventListener('keydown', onKeydown))

// ---- 状态栏派生数据：目标 / 代理指示灯 / 最近响应摘要 ----
const targetText = computed(() => {
  const target = store.activeTab?.target
  if (!target?.host) return t('status.noTarget')
  const port = target.port > 0 ? `:${target.port}` : ''
  return `${target.scheme}://${target.host}${port}`
})

const proxyText = computed(() => {
  const target = store.activeTab?.target
  if (!target?.useProxy) return t('status.proxyOff')
  return t('status.proxyOn', { url: target.proxyURL })
})

const lastSummary = computed(() => {
  const r = store.activeHistory?.response
  if (!r) return ''
  if (r.error) return r.error
  const kb = r.bytes < 1024 ? `${r.bytes} B` : `${(r.bytes / 1024).toFixed(1)} KB`
  return `${r.statusLine} · ${r.durationMs} ms · ${kb}`
})
</script>

<template>
  <div class="repeater-view">
    <TabBar />
    <TargetBar @send="doSend" />
    <div class="repeater-body">
      <n-split direction="horizontal" :default-size="0.5" :min="0.2" :max="0.8">
        <template #1>
          <RequestPanel @send="doSend" />
        </template>
        <template #2>
          <ResponsePanel />
        </template>
      </n-split>
    </div>

    <footer class="status-bar mono">
      <span class="sb-item" :title="targetText">{{ targetText }}</span>
      <span class="sb-item sb-proxy" :class="{ on: store.activeTab?.target.useProxy }">
        <span class="sb-dot"></span>{{ proxyText }}
      </span>
      <span class="sb-spacer"></span>
      <span
        v-if="lastSummary"
        class="sb-item sb-last"
        :class="{ err: !!store.activeHistory?.response?.error }"
        :title="lastSummary"
      >{{ lastSummary }}</span>
      <span class="sb-item">{{ t('status.history', { n: store.activeTab?.histories.length ?? 0 }) }}</span>
    </footer>
  </div>
</template>

<style scoped>
.repeater-view {
  height: 100%;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.repeater-body {
  flex: 1;
  min-height: 0;
}

.repeater-body :deep(.n-split) {
  height: 100%;
}

.repeater-body :deep(.n-split-pane-1),
.repeater-body :deep(.n-split-pane-2) {
  min-width: 0;
  overflow: hidden;
}

.status-bar {
  flex: none;
  height: 24px;
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 0 12px;
  border-top: 1px solid var(--border-soft);
  background: var(--bar-bg);
  font-size: 11px;
  color: var(--text-muted);
  user-select: none;
  overflow: hidden;
}

.sb-item {
  min-width: 0;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.sb-proxy {
  flex: none;
  display: inline-flex;
  align-items: center;
  gap: 5px;
}

.sb-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--text-dim);
}

.sb-proxy.on {
  color: var(--accent);
}

.sb-proxy.on .sb-dot {
  background: var(--accent);
  box-shadow: 0 0 4px var(--accent-soft);
}

.sb-spacer {
  flex: 1;
}

.sb-last {
  max-width: 46%;
}

.sb-last.err {
  color: var(--error);
}
</style>
