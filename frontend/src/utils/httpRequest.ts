// 从 AI 回复中提取可直接应用的原始 HTTP 请求
const REQUEST_LINE = /^[A-Za-z][A-Za-z0-9-]*[ \t]+\S+[ \t]+HTTP\/\d(?:\.\d)?[ \t]*$/m
const REQUEST_START = /^[A-Za-z][A-Za-z0-9-]*[ \t]+\S+[ \t]+HTTP\/\d(?:\.\d)?[ \t]*\r?\n/

function looksLikeRequest(text: string): boolean {
  return REQUEST_START.test(text)
}

// 补上 head/body 之间的空行，避免模型输出没留空行导致 body 被忽略
function normalizeRequest(text: string): string {
  const normalized = text.replace(/\r\n/g, '\n').replace(/\r/g, '\n').trim()
  if (!normalized.includes('\n\n')) {
    return `${normalized}\n\n`
  }
  return normalized
}

export function extractHttpRequest(content: string): string | null {
  const fenced = /```([^\n`]*)\r?\n([\s\S]*?)```/g
  let fallback: string | null = null
  for (const match of content.matchAll(fenced)) {
    const lang = (match[1] || '').trim().toLowerCase()
    const body = (match[2] || '').trim()
    if (!body || !looksLikeRequest(body)) continue
    const preferred = lang === 'http' || lang === 'http-request' || lang === 'request'
    if (preferred) return normalizeRequest(body)
    if (!fallback) fallback = normalizeRequest(body)
  }
  if (fallback) return fallback

  const trimmed = content.trim()
  if (looksLikeRequest(trimmed) && REQUEST_LINE.test(trimmed)) {
    return normalizeRequest(trimmed)
  }
  return null
}
