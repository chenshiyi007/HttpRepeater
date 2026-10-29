<script lang="ts" setup>
import { computed } from 'vue'
import {
  NButton,
  NConfigProvider,
  NDialogProvider,
  NIcon,
  NMessageProvider,
  darkTheme,
  lightTheme,
  zhCN,
  dateZhCN,
  enUS,
  dateEnUS,
} from 'naive-ui'
import { MoonOutline, SunnyOutline, LogoGithub } from '@vicons/ionicons5'
import RepeaterView from './views/RepeaterView.vue'
import { useSettingsStore } from './stores/settings'
import { lang, t, toggleLang } from './i18n'
import { BrowserOpenURL } from '../wailsjs/runtime/runtime'

const GITHUB_URL = 'https://github.com/chenshiyi007/HttpRepeater'

const rawVersion = String(import.meta.env.VITE_APP_VERSION ?? '').trim()
const rawVersionNumber = rawVersion.replace(/^v/, '')
const appVersion =
  !rawVersion || rawVersionNumber === '0.0.0' || rawVersion === '(devel)'
    ? ''
    : rawVersion.startsWith('v') || rawVersion.startsWith('dev+')
      ? rawVersion
      : `v${rawVersion}`
const appVersionDetail = [
  import.meta.env.VITE_APP_COMMIT ? `commit ${import.meta.env.VITE_APP_COMMIT}` : '',
  import.meta.env.VITE_APP_BUILD_DATE ? `built ${import.meta.env.VITE_APP_BUILD_DATE}` : '',
]
  .filter(Boolean)
  .join(' · ')

const settings = useSettingsStore()
const naiveTheme = computed(() => (settings.theme === 'dark' ? darkTheme : lightTheme))
const naiveLocale = computed(() => (lang.value === 'zh' ? zhCN : enUS))
const naiveDateLocale = computed(() => (lang.value === 'zh' ? dateZhCN : dateEnUS))

function openGitHub() {
  try {
    BrowserOpenURL(GITHUB_URL)
  } catch {
    window.open(GITHUB_URL, '_blank', 'noopener,noreferrer')
  }
}
</script>

<template>
  <n-config-provider
    class="app-root"
    :theme="naiveTheme"
    :locale="naiveLocale"
    :date-locale="naiveDateLocale"
  >
    <n-dialog-provider>
      <n-message-provider>
        <div class="app-shell">
          <header class="app-header">
            <span class="app-title">Http<b class="app-title-accent">Repeater</b></span>
            <span
              v-if="appVersion"
              class="app-version mono"
              :title="appVersionDetail || appVersion"
            >{{ appVersion }}</span>
            <span class="app-subtitle">{{ t('app.subtitle') }}</span>
            <div class="app-header-spacer"></div>
            <a
              class="github-link"
              :href="GITHUB_URL"
              :title="GITHUB_URL"
              @click.prevent="openGitHub"
            >
              <n-icon :component="LogoGithub" :size="15" />
              <span class="github-url mono">github.com/chenshiyi007/HttpRepeater</span>
              <span class="github-sep">·</span>
              <span class="github-author">{{ t('app.byAuthor', { name: 'chenshiyi' }) }}</span>
            </a>
            <n-button
              quaternary
              size="small"
              class="lang-btn"
              :title="t('app.switchLang')"
              @click="toggleLang()"
            >
              {{ lang === 'zh' ? 'EN' : '中' }}
            </n-button>
            <n-button
              quaternary
              circle
              size="small"
              :title="settings.theme === 'dark' ? t('app.switchToLight') : t('app.switchToDark')"
              @click="settings.toggleTheme()"
            >
              <template #icon>
                <n-icon :component="settings.theme === 'dark' ? SunnyOutline : MoonOutline" />
              </template>
            </n-button>
          </header>
          <div class="app-body">
            <RepeaterView />
          </div>
        </div>
      </n-message-provider>
    </n-dialog-provider>
  </n-config-provider>
</template>

<style scoped>
.app-root {
  height: 100%;
}

.app-shell {
  display: flex;
  flex-direction: column;
  height: 100%;
  overflow: hidden;
  background: var(--app-bg);
}

.app-header {
  height: 40px;
  flex: none;
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 0 10px 0 14px;
  border-bottom: 1px solid var(--border);
  background: var(--tabbar-bg);
}

.app-title {
  font-size: 15px;
  font-weight: 600;
  letter-spacing: 0.5px;
  color: var(--text-strong);
}

.app-title-accent {
  color: var(--accent);
}

.app-version {
  flex: none;
  padding: 1px 6px;
  border: 1px solid var(--border);
  border-radius: 9px;
  background: var(--hover-bg-soft);
  color: var(--text-muted);
  font-size: 10.5px;
  line-height: 15px;
  letter-spacing: 0.2px;
}

.app-subtitle {
  font-size: 12px;
  color: var(--text-dim);
}

.app-header-spacer {
  flex: 1;
}

.github-link {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  min-width: 0;
  max-width: 380px;
  padding: 3px 8px;
  border-radius: 5px;
  color: var(--text-muted);
  text-decoration: none;
  font-size: 11.5px;
  white-space: nowrap;
  user-select: none;
}

.github-link:hover {
  background: var(--hover-bg-soft);
  color: var(--text);
}

.github-url {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  font-size: 11px;
}

.github-sep {
  flex: none;
  color: var(--text-dim);
}

.github-author {
  flex: none;
  color: var(--text-dim);
}

.lang-btn {
  min-width: 30px;
  font-size: 12px;
  font-weight: 600;
}

.app-body {
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
}
</style>
