import { ClipboardSetText } from '../../wailsjs/runtime/runtime'

// Wails 运行时优先，浏览器调试环境回退 Web Clipboard API
export async function copyText(text: string): Promise<void> {
  try {
    const ok = await ClipboardSetText(text)
    if (ok) return
  } catch {
    // 无 Wails runtime 时忽略
  }
  await navigator.clipboard.writeText(text)
}
