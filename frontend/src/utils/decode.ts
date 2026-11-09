export type DecodeKind = 'url' | 'base64' | 'hex' | 'html'

export function decodeText(kind: DecodeKind, input: string): { ok: true; text: string } | { ok: false; error: string } {
  const raw = input
  if (!raw) return { ok: false, error: 'empty' }
  try {
    switch (kind) {
      case 'url':
        return { ok: true, text: decodeURIComponent(raw.replace(/\+/g, ' ')) }
      case 'base64': {
        const normalized = raw.replace(/\s/g, '')
        const binary = atob(normalized)
        const bytes = Uint8Array.from(binary, (c) => c.charCodeAt(0))
        return { ok: true, text: new TextDecoder().decode(bytes) }
      }
      case 'hex': {
        const hex = raw.replace(/^0x/i, '').replace(/[\s:]/g, '')
        if (!/^[0-9a-fA-F]*$/.test(hex) || hex.length % 2 !== 0) {
          return { ok: false, error: 'invalid hex' }
        }
        const bytes = new Uint8Array(hex.length / 2)
        for (let i = 0; i < bytes.length; i++) {
          bytes[i] = parseInt(hex.slice(i * 2, i * 2 + 2), 16)
        }
        return { ok: true, text: new TextDecoder().decode(bytes) }
      }
      case 'html': {
        const el = document.createElement('textarea')
        el.innerHTML = raw
        return { ok: true, text: el.value }
      }
      default:
        return { ok: false, error: 'unknown' }
    }
  } catch {
    return { ok: false, error: 'decode failed' }
  }
}
