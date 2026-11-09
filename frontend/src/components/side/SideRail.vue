<script lang="ts" setup>
import { computed, onBeforeUnmount } from 'vue'
import { NButton, NIcon } from 'naive-ui'
import {
  ChevronBackOutline,
  CodeSlashOutline,
  SparklesOutline,
  TimeOutline,
} from '@vicons/ionicons5'
import { useUiStore, type SideRailTab } from '../../stores/ui'
import { t, type MessageKey } from '../../i18n'
import AiChatPanel from './AiChatPanel.vue'
import DecoderPanel from './DecoderPanel.vue'
import TimePanel from './TimePanel.vue'

const ui = useUiStore()

const RAIL_STRIP = 44

const railWidth = computed(() => (ui.sideRailOpen ? ui.sideRailWidth : RAIL_STRIP))

const tabs: { key: SideRailTab; icon: typeof SparklesOutline; titleKey: MessageKey }[] = [
  { key: 'ai', icon: SparklesOutline, titleKey: 'side.ai' },
  { key: 'decode', icon: CodeSlashOutline, titleKey: 'side.decode' },
  { key: 'time', icon: TimeOutline, titleKey: 'side.time' },
]

function onTabClick(tab: SideRailTab) {
  if (ui.sideRailOpen && ui.sideRailTab === tab) {
    ui.closeSideRail()
    return
  }
  ui.setSideRailTab(tab)
}

let resizing = false
let startX = 0
let startW = 0

function onResizeMove(event: PointerEvent) {
  if (!resizing) return
  const delta = startX - event.clientX
  ui.setSideRailWidth(startW + delta)
}

function onResizeEnd() {
  resizing = false
  window.removeEventListener('pointermove', onResizeMove)
  window.removeEventListener('pointerup', onResizeEnd)
  document.body.style.cursor = ''
}

function onResizeStart(event: PointerEvent) {
  if (!ui.sideRailOpen) return
  event.preventDefault()
  resizing = true
  startX = event.clientX
  startW = ui.sideRailWidth
  document.body.style.cursor = 'col-resize'
  window.addEventListener('pointermove', onResizeMove)
  window.addEventListener('pointerup', onResizeEnd)
}

onBeforeUnmount(() => {
  onResizeEnd()
})

const showContent = computed(() => ui.sideRailOpen)
</script>

<template>
  <aside class="side-rail" :style="{ width: `${railWidth}px` }">
    <div
      v-if="showContent"
      class="resize-handle"
      :title="t('side.resize')"
      @pointerdown="onResizeStart"
    />

    <div class="rail-strip">
      <n-button
        v-for="item in tabs"
        :key="item.key"
        quaternary
        circle
        size="small"
        class="rail-tab"
        :class="{ active: ui.sideRailTab === item.key && ui.sideRailOpen }"
        :title="`${t(item.titleKey)} · ${t('side.tabHint')}`"
        @click="onTabClick(item.key)"
      >
        <template #icon>
          <n-icon :component="item.icon" :size="17" />
        </template>
      </n-button>
      <n-button
        v-if="ui.sideRailOpen"
        quaternary
        circle
        size="tiny"
        class="rail-collapse"
        :title="t('side.collapse')"
        @click="ui.closeSideRail()"
      >
        <template #icon>
          <n-icon :component="ChevronBackOutline" :size="14" />
        </template>
      </n-button>
    </div>

    <div v-if="showContent" class="rail-content">
      <AiChatPanel v-show="ui.sideRailTab === 'ai'" />
      <DecoderPanel v-show="ui.sideRailTab === 'decode'" />
      <TimePanel v-show="ui.sideRailTab === 'time'" />
    </div>
  </aside>
</template>

<style scoped>
.side-rail {
  position: relative;
  flex: none;
  height: 100%;
  display: flex;
  flex-direction: row;
  border-left: 1px solid var(--border-soft);
  background: var(--bar-bg);
  overflow: hidden;
  min-width: 44px;
}

.resize-handle {
  position: absolute;
  left: 0;
  top: 0;
  bottom: 0;
  width: 5px;
  cursor: col-resize;
  z-index: 2;
  touch-action: none;
}

.resize-handle:hover {
  background: var(--accent-border);
}

.rail-strip {
  flex: none;
  width: 44px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
  padding: 8px 0;
  border-right: 1px solid var(--border-soft);
}

.rail-tab.active {
  color: var(--accent);
  background: var(--hover-bg-soft);
}

.rail-collapse {
  margin-top: auto;
}

.rail-content {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  padding: 10px;
  overflow: hidden;
}

.rail-content > * {
  flex: 1;
  min-height: 0;
}
</style>
