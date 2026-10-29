<script lang="ts" setup>
import { computed } from 'vue'
import { NButton, NIcon, NInput, NPopover, NSelect, NSwitch, useMessage } from 'naive-ui'
import { CloseCircleOutline, LinkOutline, PaperPlaneOutline, RefreshOutline, SettingsOutline } from '@vicons/ionicons5'
import { useTabsStore } from '../stores/tabs'
import { DEFAULT_PROXY_URL, useSettingsStore } from '../stores/settings'
import { parseRaw } from '../api/repeater'
import { lang, t } from '../i18n'
import { sendShortcut } from '../utils/platform'

const emit = defineEmits<{ (e: 'send'): void }>()

const store = useTabsStore()
const settings = useSettingsStore()
const message = useMessage()

const defaultProxyURL = computed(() => settings.defaultProxyUrl || DEFAULT_PROXY_URL)

const tab = computed(() => store.activeTab)
const schemeOptions = [
  { label: 'HTTP', value: 'http' },
  { label: 'HTTPS', value: 'https' },
]

const portText = computed(() => {
  const port = tab.value?.target.port ?? 0
  return port > 0 ? String(port) : ''
})

function onSchemeChange(value: string) {
  store.updateTarget({ scheme: value === 'https' ? 'https' : 'http' })
}

function onHostInput(value: string) {
  store.updateTarget({ host: value.trim() })
}

function onPortInput(value: string) {
  const n = parseInt(value, 10)
  store.updateTarget({ port: Number.isFinite(n) && n > 0 ? n : 0 })
}

async function parseTarget() {
  const current = tab.value
  if (!current) return
  if (!current.requestText.trim()) {
    message.warning(t('error.requestEmpty'))
    return
  }
  try {
    const parsed = await parseRaw(current.requestText, lang.value)
    if (parsed.isBareUrl && parsed.skeleton) {
      store.setRequestText(parsed.skeleton)
    }
    if (!parsed.target.host) {
      message.warning(t('error.parseHostMissing'))
      return
    }
    store.updateTarget({
      scheme: parsed.target.scheme,
      host: parsed.target.host,
      port: parsed.target.port,
    })
    const suffix = parsed.target.port > 0 ? `:${parsed.target.port}` : ''
    message.success(
      t('error.parseSuccess', {
        url: `${parsed.target.scheme}://${parsed.target.host}${suffix}`,
      }),
    )
  } catch (e) {
    const msg = e instanceof Error ? e.message : String(e)
    message.error(t('error.parseFailed', { msg }))
  }
}

// ---- 下游代理（发包出口；每标签独立记忆，下一次发送生效）----
const useProxy = computed(() => tab.value?.target.useProxy ?? false)

function toggleProxy() {
  const current = tab.value
  if (current) store.updateTarget({ useProxy: !current.target.useProxy })
}

function onProxyInput(value: string) {
  store.updateTarget({ proxyURL: value.trim() })
}

function resetProxy() {
  store.updateTarget({ proxyURL: defaultProxyURL.value })
}
</script>

<template>
  <div class="target-bar">
    <span class="target-label">{{ t('target.label') }}</span>

    <n-select
      size="small"
      class="scheme-select"
      :value="tab?.target.scheme ?? 'http'"
      :options="schemeOptions"
      @update:value="onSchemeChange"
    />

    <n-input
      size="small"
      class="host-input"
      :placeholder="t('target.hostPlaceholder')"
      :value="tab?.target.host ?? ''"
      @update:value="onHostInput"
    />

    <n-input
      size="small"
      class="port-input"
      :placeholder="t('target.portPlaceholder')"
      :value="portText"
      @update:value="onPortInput"
    />

    <n-button size="small" quaternary class="parse-btn" :title="t('target.parseTitle')" @click="parseTarget">
      <template #icon>
        <n-icon :component="RefreshOutline" />
      </template>
      {{ t('target.parse') }}
    </n-button>

    <div class="bar-spacer"></div>

    <div class="proxy-wrap">
      <n-button
        size="small"
        class="proxy-btn"
        :type="useProxy ? 'primary' : 'default'"
        :secondary="useProxy"
        :title="t('target.proxyToggleTitle')"
        @click="toggleProxy"
      >
        <template #icon>
          <n-icon :component="LinkOutline" />
        </template>
        {{
          t('target.proxyLabel', {
            state: t(useProxy ? 'target.proxyOn' : 'target.proxyOff'),
          })
        }}
      </n-button>

      <n-popover trigger="click" placement="bottom-end">
        <template #trigger>
          <n-button size="small" quaternary class="proxy-setting-btn" :title="t('target.proxySettingsTitle')">
            <n-icon :component="SettingsOutline" :size="14" />
          </n-button>
        </template>
        <div class="proxy-pop">
          <div class="proxy-pop-title">{{ t('target.proxyPopTitle') }}</div>
          <n-input
            size="small"
            :value="tab?.target.proxyURL ?? ''"
            :placeholder="defaultProxyURL"
            @update:value="onProxyInput"
          />
          <n-button size="tiny" quaternary @click="resetProxy">
            {{ t('target.proxyReset', { url: defaultProxyURL }) }}
          </n-button>
          <div class="proxy-pop-hint">{{ t('target.proxyHint') }}</div>
        </div>
      </n-popover>
    </div>

    <label class="toggle">
      <n-switch
        size="small"
        :value="tab?.followHostHeader ?? true"
        @update:value="(v: boolean) => store.setFollowHostHeader(v)"
      />
      <span class="toggle-label">{{ t('target.followHost') }}</span>
    </label>

    <label class="toggle">
      <n-switch
        size="small"
        :value="tab?.autoContentLength ?? true"
        @update:value="(v: boolean) => store.setAutoContentLength(v)"
      />
      <span class="toggle-label">{{ t('target.autoCL') }}</span>
    </label>

    <n-button
      v-if="store.sending"
      type="error"
      secondary
      size="small"
      class="send-btn"
      @click="store.cancelSending()"
    >
      <template #icon>
        <n-icon :component="CloseCircleOutline" />
      </template>
      {{ t('target.cancel') }}
    </n-button>
    <n-button
      v-else
      type="primary"
      size="small"
      class="send-btn"
      @click="emit('send')"
    >
      <template #icon>
        <n-icon :component="PaperPlaneOutline" />
      </template>
      {{ t('target.send') }}
      <span class="shortcut">{{ sendShortcut }}</span>
    </n-button>
  </div>
</template>

<style scoped>
.target-bar {
  flex: none;
  display: flex;
  align-items: center;
  gap: 8px;
  height: 42px;
  padding: 0 12px;
  border-bottom: 1px solid var(--border);
  background: var(--bar-bg);
}

.target-label {
  font-size: 12px;
  color: var(--text-muted);
  white-space: nowrap;
}

.scheme-select {
  width: 92px;
  flex: none;
}

.host-input {
  flex: 1;
  min-width: 120px;
  max-width: 340px;
}

.port-input {
  width: 78px;
  flex: none;
}

.parse-btn {
  flex: none;
  font-size: 12px;
}

.bar-spacer {
  flex: 1;
}

.proxy-wrap {
  flex: none;
  display: flex;
  align-items: center;
  gap: 2px;
}

.proxy-btn {
  font-size: 12px;
}

.proxy-setting-btn {
  padding: 0 4px !important;
}

.proxy-pop {
  display: flex;
  flex-direction: column;
  gap: 8px;
  width: 260px;
}

.proxy-pop-title {
  font-size: 12px;
  color: var(--text);
}

.proxy-pop-hint {
  font-size: 11px;
  line-height: 1.5;
  color: var(--text-muted);
}

.toggle {
  display: flex;
  align-items: center;
  gap: 6px;
  flex: none;
}

.toggle-label {
  font-size: 12px;
  color: var(--text);
  white-space: nowrap;
}

.send-btn {
  flex: none;
}

.shortcut {
  font-size: 11px;
  opacity: 0.65;
  margin-left: 2px;
}
</style>
