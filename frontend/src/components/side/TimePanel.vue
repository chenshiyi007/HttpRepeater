<script lang="ts" setup>
import { onBeforeUnmount, onMounted, ref } from 'vue'
import { NAlert, NButton } from 'naive-ui'
import { BrowserOpenURL } from '../../../wailsjs/runtime/runtime'
import { t } from '../../i18n'
import { formatBeijingNow } from '../../utils/beijingTime'

/** 嵌入 beijing-time.org；失败时回退本地时钟卡片 */
const TIME_IFRAME_URL = 'https://www.beijing-time.org/'
const LOAD_TIMEOUT_MS = 12000

const useFallback = ref(false)
const clock = ref(formatBeijingNow())
let tickTimer: ReturnType<typeof setInterval> | undefined
let loadTimer: ReturnType<typeof setTimeout> | undefined

function startLocalClock() {
  clock.value = formatBeijingNow()
  tickTimer = setInterval(() => {
    clock.value = formatBeijingNow()
  }, 1000)
}

function failToFallback() {
  if (useFallback.value) return
  useFallback.value = true
  if (loadTimer) {
    clearTimeout(loadTimer)
    loadTimer = undefined
  }
  startLocalClock()
}

function onIframeLoad() {
  if (loadTimer) {
    clearTimeout(loadTimer)
    loadTimer = undefined
  }
}

function onIframeError() {
  failToFallback()
}

function retryIframe() {
  if (tickTimer) {
    clearInterval(tickTimer)
    tickTimer = undefined
  }
  useFallback.value = false
  loadTimer = setTimeout(() => failToFallback(), LOAD_TIMEOUT_MS)
}

function openSource() {
  try {
    BrowserOpenURL(TIME_IFRAME_URL)
  } catch {
    window.open(TIME_IFRAME_URL, '_blank', 'noopener,noreferrer')
  }
}

onMounted(() => {
  loadTimer = setTimeout(() => failToFallback(), LOAD_TIMEOUT_MS)
})

onBeforeUnmount(() => {
  if (tickTimer) clearInterval(tickTimer)
  if (loadTimer) clearTimeout(loadTimer)
})
</script>

<template>
  <div class="time-panel">
    <div class="time-head">
      <span class="time-title">{{ t('time.title') }}</span>
      <span class="time-sub">{{ t('time.sourceHint') }}</span>
    </div>

    <n-alert v-if="useFallback" type="warning" :bordered="false" class="time-alert">
      <div class="alert-inner">
        <span>{{ t('time.loadFailed') }}</span>
        <n-button size="tiny" tertiary @click="retryIframe">{{ t('time.retry') }}</n-button>
      </div>
    </n-alert>

    <iframe
      v-if="!useFallback"
      class="time-iframe"
      :src="TIME_IFRAME_URL"
      title="beijing-time.org"
      referrerpolicy="no-referrer"
      @load="onIframeLoad"
      @error="onIframeError"
    />

    <div v-else class="time-local">
      <p class="card-caption">{{ t('time.nowCaption') }}</p>
      <div class="card">
        <p class="card-date">{{ clock.dateLine }}</p>
        <p class="card-time mono">{{ clock.time }}</p>
        <p class="card-zone">{{ t('time.zoneDetail') }}</p>
      </div>
      <n-button text size="tiny" class="open-link" @click="openSource">{{ t('time.openWeb') }}</n-button>
    </div>
  </div>
</template>

<style scoped>
.time-panel {
  height: 100%;
  display: flex;
  flex-direction: column;
  min-height: 0;
  gap: 8px;
}

.time-head {
  flex: none;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.time-title {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-strong);
}

.time-sub {
  font-size: 10.5px;
  color: var(--text-dim);
  line-height: 1.4;
}

.time-alert {
  flex: none;
}

.alert-inner {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  flex-wrap: wrap;
}

.time-iframe {
  flex: 1;
  min-height: 220px;
  width: 100%;
  border: 1px solid var(--border-soft);
  border-radius: 8px;
  background: #f2f2f2;
}

.time-local {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: stretch;
  gap: 10px;
  min-height: 0;
  overflow-y: auto;
}

.card-caption {
  margin: 0;
  text-align: center;
  font-size: 14px;
  font-weight: 700;
  color: #2f86d8;
}

.card {
  padding: 14px 12px 16px;
  border-radius: 6px;
  background: #fff;
  border: 1px solid #d9d9d9;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.06);
  text-align: center;
}

.card-date {
  margin: 0 0 10px;
  font-size: 13px;
  color: #333;
}

.card-time {
  margin: 0;
  font-size: 42px;
  font-weight: 700;
  letter-spacing: 0.02em;
  color: #f3a22a;
  line-height: 1.1;
}

.card-zone {
  margin: 12px 0 0;
  font-size: 11px;
  color: #666;
  line-height: 1.5;
}

.open-link {
  align-self: center;
  color: var(--accent-soft);
}
</style>
