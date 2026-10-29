// 内置请求模板（从请求面板头部「模板」下拉选择，整体替换编辑器内容）
import type { MessageKey } from '../i18n'

const BOUNDARY = '----HttpRepeaterBoundary'

// 统一拼装：头 + 空行 + body；有 body 时按 LF 计长（发送管道开启自动重算会按实际字节重算）
function buildRequest(method: string, path: string, headers: string[], body = ''): string {
  const lines = [`${method} ${path} HTTP/1.1`, ...headers]
  if (body) lines.push(`Content-Length: ${body.length}`)
  return `${lines.join('\n')}\n\n${body}`
}

export function getTemplate(): string {
  return buildRequest('GET', '/', ['Host: example.com'])
}

// POST 表单：application/x-www-form-urlencoded
export function postFormTemplate(): string {
  return buildRequest(
    'POST',
    '/login',
    [
      'Host: example.com',
      'Content-Type: application/x-www-form-urlencoded',
    ],
    'username=admin&password=admin',
  )
}

// POST JSON
export function postJSONTemplate(): string {
  return buildRequest(
    'POST',
    '/api',
    [
      'Host: example.com',
      'Content-Type: application/json',
    ],
    '{"username":"admin","password":"admin"}',
  )
}

// 文件上传：multipart/form-data + boundary + 文件字段占位
export function fileUploadTemplate(): string {
  const body = [
    `--${BOUNDARY}`,
    'Content-Disposition: form-data; name="file"; filename="example.txt"',
    'Content-Type: text/plain',
    '',
    'Hello from HttpRepeater',
    `--${BOUNDARY}--`,
  ].join('\n')

  return buildRequest(
    'POST',
    '/upload',
    [
      'Host: example.com',
      `Content-Type: multipart/form-data; boundary=${BOUNDARY}`,
    ],
    body,
  )
}

// Host 留空：不发 Host 头（配合关闭「Host 跟随」用于 Host 校验类测试）
export function emptyHostTemplate(): string {
  return [
    'GET / HTTP/1.1',
    'User-Agent: HttpRepeater/0.1.0',
    'Accept: */*',
    '',
    '',
  ].join('\n')
}

export interface RequestTemplate {
  key: string
  labelKey: MessageKey
  build: () => string
  // 应用模板时联动关闭「Host 跟随」（Host 留空模板）
  disableFollowHost?: boolean
}

export const builtinTemplates: RequestTemplate[] = [
  { key: 'get', labelKey: 'tpl.get', build: getTemplate },
  { key: 'post-form', labelKey: 'tpl.postForm', build: postFormTemplate },
  { key: 'post-json', labelKey: 'tpl.postJson', build: postJSONTemplate },
  { key: 'multipart-upload', labelKey: 'tpl.upload', build: fileUploadTemplate },
  { key: 'empty-host', labelKey: 'tpl.emptyHost', build: emptyHostTemplate, disableFollowHost: true },
  { key: 'clear', labelKey: 'tpl.clear', build: () => '' },
]
