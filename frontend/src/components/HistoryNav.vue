<script lang="ts" setup>
import { computed, h, ref } from 'vue'
import { NButton, NDropdown, NIcon, useDialog } from 'naive-ui'
import { ChevronBackOutline, ChevronForwardOutline } from '@vicons/ionicons5'
import { useTabsStore } from '../stores/tabs'
import type { HistoryItem } from '../types'
import { t } from '../i18n'

const store = useTabsStore()
const dialog = useDialog()
const showList = ref(false)

const total = computed(() => store.activeTab?.histories.length ?? 0)
const index = computed(() => store.activeTab?.activeHistoryIndex ?? -1)
const label = computed(() =>
  total.value > 0 && index.value >= 0 ? `${index.value + 1}/${total.value}` : `0/${total.value}`,
)

// 下拉列表：新的在前，便于找最近一次发送
const listOptions = computed(() => {
  const items = store.activeTab?.histories ?? []
  return items
    .map((item, i) => ({ key: i, item }))
    .reverse()
    .map(({ key, item }) => ({
      key,
      type: 'render' as const,
      render: () => renderHistoryRow(key, item),
    }))
})

// 下拉挂在 body 下，scoped 样式不可达，行内样式渲染
function renderHistoryRow(i: number, item: HistoryItem) {
  const failed = !!item.response?.error
  const code = item.response ? (failed ? t('history.failed') : String(item.response.statusCode)) : '—'
  const time = new Date(item.sentAt).toTimeString().slice(0, 8)
  const isCurrent = i === index.value
  return h(
    'div',
    {
      style: {
        display: 'flex',
        alignItems: 'center',
        gap: '10px',
        padding: '5px 12px',
        minWidth: '230px',
        fontSize: '12px',
        fontFamily: "'SF Mono', Menlo, Consolas, monospace",
        color: isCurrent ? 'var(--accent)' : 'var(--text-body)',
      },
    },
    [
      h('span', {
        style: {
          flex: 'none',
          width: '7px',
          height: '7px',
          borderRadius: '50%',
          background: failed ? 'var(--error)' : 'var(--accent-soft)',
        },
      }),
      h('span', { style: { flex: 'none', opacity: '0.65' } }, `#${item.seq}`),
      h('span', { style: { flex: 'none', opacity: '0.65' } }, time),
      h('span', { style: { color: failed ? 'var(--error)' : 'inherit' } }, code),
      h(
        'span',
        { style: { marginLeft: 'auto', opacity: '0.55' } },
        item.response ? `${item.response.durationMs} ms` : '',
      ),
    ],
  )
}

function jumpTo(target: number) {
  const tab = store.activeTab
  if (!tab || target === tab.activeHistoryIndex) return
  if (target < 0 || target >= tab.histories.length) return
  if (store.historyDirty) {
    dialog.warning({
      title: t('history.discardTitle'),
      content: t('history.discardContent'),
      positiveText: t('history.discardConfirm'),
      negativeText: t('history.cancel'),
      onPositiveClick: () => store.loadHistory(target),
    })
  } else {
    store.loadHistory(target)
  }
}

function step(delta: number) {
  jumpTo(index.value + delta)
}
</script>

<template>
  <div v-if="total > 0" class="history-nav">
    <span
      v-if="store.historyDirty"
      class="dirty-dot"
      :title="t('history.dirty')"
    ></span>

    <n-button
      size="tiny"
      quaternary
      class="nav-btn"
      :disabled="index <= 0"
      :title="t('history.prev')"
      @click="step(-1)"
    >
      <n-icon :component="ChevronBackOutline" :size="13" />
    </n-button>

    <n-dropdown
      trigger="click"
      :show="showList"
      :options="listOptions"
      placement="bottom-start"
      @update:show="(v: boolean) => (showList = v)"
      @select="(key: string | number) => jumpTo(Number(key))"
    >
      <button class="nav-label mono" :title="t('history.viewList')">
        {{ t('history.label', { label }) }}
      </button>
    </n-dropdown>

    <n-button
      size="tiny"
      quaternary
      class="nav-btn"
      :disabled="index < 0 || index >= total - 1"
      :title="t('history.next')"
      @click="step(1)"
    >
      <n-icon :component="ChevronForwardOutline" :size="13" />
    </n-button>
  </div>
</template>

<style scoped>
.history-nav {
  flex: none;
  display: flex;
  align-items: center;
  gap: 2px;
}

.dirty-dot {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: var(--dirty);
  margin: 0 5px 0 2px;
  flex: none;
}

.nav-btn {
  padding: 0 2px !important;
}

.nav-label {
  border: none;
  background: transparent;
  padding: 2px 6px;
  font-size: 12px;
  color: var(--text);
  cursor: pointer;
  border-radius: 4px;
  white-space: nowrap;
}

.nav-label:hover {
  background: var(--hover-bg-soft);
  color: var(--text-strong);
}
</style>
