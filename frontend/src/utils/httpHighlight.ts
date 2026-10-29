import { StreamLanguage, type StreamParser } from '@codemirror/language'
import { tags as t } from '@lezer/highlight'

// 极简 HTTP 报文高亮：
// 行 0 = 请求行（方法 keyword / target string / 版本 number）
// 之后 = header 行（名字 propertyName / 值 string），遇到空行进入 body
interface HttpState {
  phase: 0 | 1 | 2 // 0=请求行 1=头 2=body
}

const httpParser: StreamParser<HttpState> = {
  startState: () => ({ phase: 0 }),

  tokenTable: {
    keyword: t.keyword,
    string: t.string,
    number: t.number,
    propertyName: t.propertyName,
  },

  token(stream, state) {
    if (state.phase === 2) {
      stream.skipToEnd()
      return null
    }

    if (state.phase === 0) {
      if (stream.match(/^[A-Za-z]+/)) return 'keyword'
      if (stream.match(/^\s+/)) return null
      if (stream.match(/^HTTP\/[0-9.]+/i)) return 'number'
      if (stream.match(/^\S+/)) return 'string'
      stream.skipToEnd()
      state.phase = 1
      return null
    }

    // headers
    if (stream.sol()) {
      if (stream.eol()) {
        state.phase = 2
        return null
      }
      const name = stream.match(/^[^\s:]+/)
      if (name) {
        return stream.peek() === ':' ? 'propertyName' : 'string'
      }
    }
    if (stream.match(/^:/)) return null
    if (stream.match(/^.*$/)) return 'string'
    stream.skipToEnd()
    return null
  },
}

export const httpLanguage = StreamLanguage.define(httpParser)
