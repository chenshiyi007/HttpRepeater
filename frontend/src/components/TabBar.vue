<script lang="ts" setup>
import { nextTick, ref } from 'vue'
import { useTabsStore } from '../stores/tabs'
import { t } from '../i18n'
import { modKey } from '../utils/platform'

const store = useTabsStore()

const editingId = ref<string | null>(null)
const editingTitle = ref('')
const renameInputRef = ref<HTMLInputElement | null>(null)

function startRename(tabId: string, title: string) {
  editingId.value = tabId
  editingTitle.value = title
  nextTick(() => {
    renameInputRef.value?.focus()
    renameInputRef.value?.select()
  })
}

function commitRename() {
  if (editingId.value !== null) {
    store.renameTab(editingId.value, editingTitle.value)
    editingId.value = null
  }
}

function cancelRename() {
  editingId.value = null
}
</script>

<template>
  <div class="tab-bar">
    <div class="tab-strip">
      <div
        v-for="tab in store.tabs"
        :key="tab.id"
        class="tab"
        :class="{ active: tab.id === store.activeTabId }"
        :title="tab.title"
        @click="store.setActive(tab.id)"
        @dblclick="startRename(tab.id, tab.title)"
      >
        <template v-if="editingId === tab.id">
          <input
            ref="renameInputRef"
            v-model="editingTitle"
            class="tab-rename"
            @keyup.enter="commitRename"
            @keyup.esc="cancelRename"
            @blur="commitRename"
            @click.stop
            @dblclick.stop
          />
        </template>
        <template v-else>
          <span class="tab-title">{{ tab.title }}</span>
          <span
            class="tab-close"
            :title="t('tab.close', { mod: modKey })"
            @click.stop="store.closeTab(tab.id)"
          >
            ×
          </span>
        </template>
      </div>

      <button class="tab-add" :title="t('tab.new', { mod: modKey })" @click="store.addTab()">+</button>
    </div>
  </div>
</template>

<style scoped>
.tab-bar {
  flex: none;
  background: var(--tabbar-bg);
  border-bottom: 1px solid var(--border);
}

.tab-strip {
  display: flex;
  align-items: stretch;
  height: 34px;
  padding: 0 6px;
  gap: 2px;
}

.tab {
  display: flex;
  align-items: center;
  gap: 6px;
  max-width: 180px;
  padding: 0 6px 0 12px;
  border: 1px solid var(--border);
  border-bottom: none;
  border-radius: 6px 6px 0 0;
  background: var(--tab-bg);
  cursor: pointer;
  color: var(--text-muted);
  position: relative;
  top: 2px;
  user-select: none;
}

.tab:hover {
  color: var(--text-strong);
}

.tab.active {
  background: var(--app-bg);
  color: var(--accent);
  border-color: var(--accent-border);
  border-bottom: 2px solid var(--app-bg);
  z-index: 1;
}

.tab-title {
  font-size: 12.5px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.tab-close {
  flex: none;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 16px;
  height: 16px;
  border-radius: 3px;
  font-size: 13px;
  line-height: 1;
  color: var(--text-dim);
}

.tab-close:hover {
  background: var(--hover-bg);
  color: var(--text-strong);
}

.tab-rename {
  width: 110px;
  height: 22px;
  padding: 0 4px;
  font-size: 12.5px;
  color: var(--text-strong);
  background: var(--input-bg);
  border: 1px solid var(--accent-border);
  border-radius: 3px;
  outline: none;
  user-select: text;
}

.tab-add {
  flex: none;
  width: 26px;
  margin-top: 3px;
  border: none;
  border-radius: 4px;
  background: transparent;
  color: var(--text-muted);
  font-size: 16px;
  cursor: pointer;
}

.tab-add:hover {
  background: var(--hover-bg);
  color: var(--text-strong);
}
</style>
