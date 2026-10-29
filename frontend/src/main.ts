import { createApp } from 'vue'
import { createPinia } from 'pinia'
import App from './App.vue'
import './style.css'
import { useSettingsStore } from './stores/settings'
import { initLang } from './i18n'

// 屏蔽 WebView 默认右键菜单，保持桌面应用的操作习惯
window.addEventListener('contextmenu', (event) => {
  event.preventDefault()
})

const pinia = createPinia()
const app = createApp(App).use(pinia)

// 挂载前应用主题与语言，避免首帧闪烁
useSettingsStore(pinia).initTheme()
initLang()

app.mount('#app')
