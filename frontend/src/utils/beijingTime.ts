const TZ = 'Asia/Shanghai'

export function formatBeijingNow(date = new Date()): {
  time: string
  date: string
  weekday: string
} {
  const time = new Intl.DateTimeFormat('zh-CN', {
    timeZone: TZ,
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
    hour12: false,
  }).format(date)

  const dateStr = new Intl.DateTimeFormat('zh-CN', {
    timeZone: TZ,
    year: 'numeric',
    month: 'long',
    day: 'numeric',
  }).format(date)

  const weekday = new Intl.DateTimeFormat('zh-CN', {
    timeZone: TZ,
    weekday: 'long',
  }).format(date)

  return { time, date: dateStr, weekday }
}
