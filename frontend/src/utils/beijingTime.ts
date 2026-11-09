const TZ = 'Asia/Shanghai'

export interface BeijingClock {
  time: string
  dateLine: string
  weekday: string
}

export function formatBeijingNow(date = new Date()): BeijingClock {
  const time = new Intl.DateTimeFormat('zh-CN', {
    timeZone: TZ,
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
    hour12: false,
  }).format(date)

  const parts = Object.fromEntries(
    new Intl.DateTimeFormat('zh-CN', {
      timeZone: TZ,
      year: 'numeric',
      month: 'numeric',
      day: 'numeric',
      weekday: 'long',
    })
      .formatToParts(date)
      .map((p) => [p.type, p.value]),
  )

  const y = parts.year ?? ''
  const m = parts.month ?? ''
  const d = parts.day ?? ''
  const weekday = parts.weekday ?? ''
  const dateLine = `${y}年${m}月${d}日 ${weekday}`

  return { time, dateLine, weekday }
}
