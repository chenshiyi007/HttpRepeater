<script lang="ts" setup>
import { onBeforeUnmount, onMounted, ref } from 'vue'
import { NAlert } from 'naive-ui'
import { t } from '../../i18n'
import { formatBeijingNow } from '../../utils/beijingTime'

/** 优先尝试嵌入的北京时间页面；失败则回退本地时钟 */
const TIME_IFRAME_URL = 'https://time.is/zh/Beijing'
const LOAD_TIMEOUT_MS = 8000

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

onMounted(() => {
  loadTimer = setTimeout(() => {
    failToFallback()
  }, LOAD_TIMEOUT_MS)
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
      <span class="time-sub mono">{{ t('time.zone') }}</span>
    </div>

    <n-alert
      v-if="useFallback"
      type="warning"
      :bordered="false"
      class="time-alert"
    >
      {{ t('time.loadFailed') }}
    </n-alert>

    <iframe
      v-if="!useFallback"
      class="time-iframe"
      :src="TIME_IFRAME_URL"
      title="Beijing time"
      referrerpolicy="no-referrer"
      @load="onIframeLoad"
      @error="onIframeError"
    />

    <div v-else class="time-local">
      <div class="time-display mono">{{ clock.time }}</div>
      <div class="time-date">{{ clock.date }}</div>
      <div class="time-weekday">{{ clock.weekday }}</div>
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
  font-size: 11px;
  color: var(--text-dim);
}

.time-alert {
  flex: none;
}

.time-iframe {
  flex: 1;
  min-height: 200px;
  width: 100%;
  border: 1px solid var(--border-soft);
  border-radius: 8px;
  background: var(--input-bg);
}

.time-local {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 24px 12px;
  border: 1px solid var(--border-soft);
  border-radius: 8px;
  background: var(--input-bg);
  text-align: center;
}

.time-display {
  font-size: 36px;
  font-weight: 600;
  letter-spacing: 0.04em;
  color: var(--accent-soft);
}

.time-date {
  font-size: 14px;
  color: var(--text-body);
}

.time-weekday {
  font-size: 12px;
  color: var(--text-muted);
}
</style>
