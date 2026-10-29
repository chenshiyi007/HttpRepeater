package repeater

import (
	"bufio"
	"bytes"
	"compress/flate"
	"compress/gzip"
	"compress/zlib"
	"io"
	"strconv"
	"strings"

	"github.com/andybalholm/brotli"
)

// rawHTTPResponse 手动 socket 读到的原始响应（头保序、body 为线上字节）
type rawHTTPResponse struct {
	statusLine  string
	headerLines []string
	body        []byte
}

// readResponse 读取状态行、原始头、body（默认中文错误）
func readResponse(r io.Reader) (*rawHTTPResponse, error) {
	return readResponseLang(r, langZH)
}

// readResponseLang 读取状态行、原始头、body（Content-Length / chunked / EOF 三种模式）
func readResponseLang(r io.Reader, l lang) (*rawHTTPResponse, error) {
	br := bufio.NewReader(r)
	resp := &rawHTTPResponse{}

	statusLine, err := br.ReadString('\n')
	if err != nil {
		if len(statusLine) == 0 {
			return resp, newLangErrorf(l, "读取状态行失败: %w", "failed to read the status line: %w", err)
		}
		return resp, nil
	}
	resp.statusLine = strings.TrimRight(statusLine, "\r\n")

	headerBytes := 0
	for {
		line, err := br.ReadString('\n')
		if err != nil {
			return resp, newLangErrorf(l, "读取响应头失败: %w", "failed to read response headers: %w", err)
		}
		headerBytes += len(line)
		if headerBytes > maxHeaderBytes {
			return resp, newLangError(l, "响应头超过大小上限", "response headers exceed the size limit")
		}
		trimmed := strings.TrimRight(line, "\r\n")
		if trimmed == "" {
			break
		}
		resp.headerLines = append(resp.headerLines, trimmed)
	}

	code := parseStatusCode(resp.statusLine)
	if (code >= 100 && code < 200) || code == 204 || code == 304 {
		return resp, nil
	}

	te := headerValue(resp.headerLines, "Transfer-Encoding")
	if strings.Contains(strings.ToLower(te), "chunked") {
		body, err := readChunked(br, l)
		resp.body = body
		if err != nil {
			return resp, newLangErrorf(l, "解析 chunked 响应失败: %w", "failed to parse the chunked response: %w", err)
		}
		return resp, nil
	}

	if cl, ok := parseContentLength(resp.headerLines); ok {
		if cl > maxResponseBytes {
			return resp, newLangErrorf(l, "响应体超过大小上限（%d MB）", "response body exceeds the size limit (%d MB)", maxResponseBytes>>20)
		}
		body := make([]byte, cl)
		n, err := io.ReadFull(br, body)
		resp.body = body[:n]
		if err != nil {
			return resp, newLangErrorf(l, "读取响应体失败: %w", "failed to read the response body: %w", err)
		}
		return resp, nil
	}

	// 无 Content-Length / chunked：读到连接关闭（EOF）
	body, err := io.ReadAll(io.LimitReader(br, maxResponseBytes+1))
	if len(body) > maxResponseBytes {
		return resp, newLangErrorf(l, "响应体超过大小上限（%d MB）", "response body exceeds the size limit (%d MB)", maxResponseBytes>>20)
	}
	resp.body = body
	if err != nil {
		return resp, newLangErrorf(l, "读取响应体失败: %w", "failed to read the response body: %w", err)
	}
	return resp, nil
}

func readChunked(br *bufio.Reader, l lang) ([]byte, error) {
	var out bytes.Buffer
	for {
		line, err := br.ReadString('\n')
		if err != nil {
			return out.Bytes(), err
		}
		sizeLine := strings.TrimSpace(line)
		if sizeLine == "" {
			continue
		}
		if i := strings.Index(sizeLine, ";"); i >= 0 { // chunk extension
			sizeLine = sizeLine[:i]
		}
		size, err := strconv.ParseInt(sizeLine, 16, 64)
		if err != nil {
			return out.Bytes(), newLangErrorf(l, "chunk 长度无效: %q", "invalid chunk size: %q", sizeLine)
		}
		if size == 0 {
			// 读取 trailer 直到空行
			for {
				trailer, err := br.ReadString('\n')
				if err != nil {
					return out.Bytes(), err
				}
				if strings.TrimRight(trailer, "\r\n") == "" {
					return out.Bytes(), nil
				}
			}
		}
		if out.Len()+int(size) > maxResponseBytes {
			return out.Bytes(), newLangErrorf(l, "响应体超过大小上限（%d MB）", "response body exceeds the size limit (%d MB)", maxResponseBytes>>20)
		}
		if _, err := io.CopyN(&out, br, size); err != nil {
			return out.Bytes(), err
		}
		if _, err := br.ReadString('\n'); err != nil { // chunk 结尾 CRLF
			return out.Bytes(), err
		}
	}
}

// decodeBody 按 Content-Encoding 解码（gzip / deflate / br，多个编码按逆序处理）
func decodeBody(headerLines []string, body []byte) []byte {
	encoding := strings.ToLower(strings.TrimSpace(headerValue(headerLines, "Content-Encoding")))
	if encoding == "" || encoding == "identity" {
		return body
	}
	encodings := strings.Split(encoding, ",")
	for i := len(encodings) - 1; i >= 0; i-- {
		enc := strings.TrimSpace(encodings[i])
		var (
			r   io.Reader
			err error
			out []byte
		)
		switch {
		case enc == "gzip" || enc == "x-gzip":
			zr, e := gzip.NewReader(bytes.NewReader(body))
			if e != nil {
				return body
			}
			out, err = io.ReadAll(zr)
			zr.Close()
		case enc == "deflate":
			if zr, e := zlib.NewReader(bytes.NewReader(body)); e == nil {
				out, err = io.ReadAll(zr)
				zr.Close()
			} else {
				fr := flate.NewReader(bytes.NewReader(body))
				out, err = io.ReadAll(fr)
				fr.Close()
			}
		case enc == "br":
			r = brotli.NewReader(bytes.NewReader(body))
			out, err = io.ReadAll(r)
		default:
			continue
		}
		if err != nil || out == nil {
			return body
		}
		body = out
	}
	return body
}

func wireSize(resp *rawHTTPResponse) int {
	size := len(resp.statusLine) + 2 + 2
	for _, h := range resp.headerLines {
		size += len(h) + 2
	}
	return size + len(resp.body)
}
