// 平台相关的快捷键显示
export const isMac = /Mac|iPhone|iPad/.test(navigator.platform || navigator.userAgent)
export const modKey = isMac ? '⌘' : 'Ctrl'
export const sendShortcut = isMac ? '⌘↵' : 'Ctrl+↵'
