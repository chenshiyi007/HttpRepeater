package repeater

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"io"
	"net"
	"strconv"
	"strings"
	"time"
)

const (
	maxResponseBytes = 64 << 20 // 响应体 64MB 上限
	maxHeaderBytes   = 1 << 20  // 响应头 1MB 上限
)

// SendRequest 发送请求参数（表单目标 + 报文 + 开关）
type SendRequest struct {
	Raw               string `json:"requestText"`
	Scheme            string `json:"scheme"`
	Host              string `json:"host"`
	Port              int    `json:"port"`
	UseProxy          bool   `json:"useProxy"`
	ProxyURL          string `json:"proxyURL"`
	FollowHostHeader  bool   `json:"followHostHeader"`
	AutoContentLength bool   `json:"autoContentLength"`
	VerifyTLS         bool   `json:"verifyTLS"`
	TimeoutMs         int    `json:"timeoutMs"`
	Lang              string `json:"lang"`
}

// RawResponse 响应结果（BodyRaw 为 base64，二进制安全）
type RawResponse struct {
	StatusLine  string `json:"statusLine"`
	StatusCode  int    `json:"statusCode"`
	HeadersRaw  string `json:"headersRaw"`
	BodyRaw     string `json:"bodyRaw"`
	BodyDecoded string `json:"bodyDecoded"`
	Bytes       int    `json:"bytes"`
	DurationMs  int64  `json:"durationMs"`
	Error       string `json:"error"`
}

// SendResult 实际发出的报文 + 响应
type SendResult struct {
	RequestSent string       `json:"requestSent"`
	Response    *RawResponse `json:"response"`
}

// Send 手动 socket 逐字节发送；网络层错误写入 Response.Error，不返回 error
// （失败的发送同样要生成历史记录）
func Send(req SendRequest) *SendResult {
	l := normalizeLang(req.Lang)
	timeout := time.Duration(req.TimeoutMs) * time.Millisecond
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	start := time.Now()
	res := &SendResult{Response: &RawResponse{}}
	fail := func(msg string) *SendResult {
		res.Response.Error = msg
		res.Response.DurationMs = time.Since(start).Milliseconds()
		return res
	}

	parsed, err := parseRequestLang(req.Raw, l)
	if err != nil {
		return fail(tr(l, "请求解析失败: ", "failed to parse the request: ") + err.Error())
	}

	scheme := strings.ToLower(strings.TrimSpace(req.Scheme))
	if scheme != "http" && scheme != "https" {
		scheme = parsed.Target.Scheme
		if scheme != "https" {
			scheme = "http"
		}
	}
	host := strings.TrimSpace(req.Host)
	if host == "" {
		host = parsed.Target.Host
	}
	if host == "" {
		return fail(tr(l, "目标 Host 为空", "target Host is empty"))
	}
	port := req.Port
	if port <= 0 {
		port = parsed.Target.Port
	}
	if port <= 0 {
		if scheme == "https" {
			port = 443
		} else {
			port = 80
		}
	}

	// 下游代理（发包出口）
	var proxy *proxyConfig
	if req.UseProxy {
		proxy, err = parseProxyURLLang(req.ProxyURL, l)
		if err != nil {
			return fail(err.Error())
		}
	}

	// Host 头跟随目标重写（默认开）
	headerLines := append([]string(nil), parsed.headerLines...)
	if req.FollowHostHeader {
		headerLines = setHeader(headerLines, "Host", hostHeaderValue(host, scheme, port))
	}

	// Content-Length 自动重算（有 body 时）
	body := normalizeCRLF(parsed.Body)
	if req.AutoContentLength && len(body) > 0 {
		headerLines = setHeader(headerLines, "Content-Length", strconv.Itoa(len(body)))
	}

	// HTTP 经代理走经典绝对 URL 请求行；代理认证头只发给代理
	requestTarget := parsed.Path
	if requestTarget == "" {
		requestTarget = "/"
	}
	if proxy != nil {
		if scheme == "http" {
			requestTarget = "http://" + hostHeaderValue(host, scheme, port) + requestTarget
		}
		if proxy.authLine != "" && !hasHeader(headerLines, "Proxy-Authorization") {
			headerLines = append(headerLines, proxy.authLine)
		}
	}

	var msg bytes.Buffer
	msg.WriteString(parsed.Method)
	msg.WriteByte(' ')
	msg.WriteString(requestTarget)
	msg.WriteByte(' ')
	msg.WriteString(parsed.Version)
	msg.WriteString("\r\n")
	for _, l := range headerLines {
		msg.WriteString(l)
		msg.WriteString("\r\n")
	}
	msg.WriteString("\r\n")
	msg.WriteString(body)
	requestBytes := msg.Bytes()
	res.RequestSent = string(requestBytes)

	address := net.JoinHostPort(host, strconv.Itoa(port))
	dialer := &net.Dialer{Timeout: timeout}
	var conn net.Conn
	if proxy != nil {
		conn, err = dialViaProxy(proxy, scheme, host, port, timeout, l)
		if err != nil {
			return fail(err.Error())
		}
	} else {
		conn, err = dialer.Dial("tcp", address)
		if err != nil {
			return fail(tr(l, "连接失败: ", "connection failed: ") + err.Error())
		}
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(timeout))

	var rw io.ReadWriter = conn
	if scheme == "https" {
		tlsConn := tls.Client(conn, &tls.Config{
			InsecureSkipVerify: !req.VerifyTLS, // 默认忽略证书校验
			ServerName:         host,
		})
		if err := tlsConn.Handshake(); err != nil {
			return fail(tr(l, "TLS 握手失败: ", "TLS handshake failed: ") + err.Error())
		}
		rw = tlsConn
	}

	if _, err := rw.Write(requestBytes); err != nil {
		return fail(tr(l, "发送请求失败: ", "failed to send the request: ") + err.Error())
	}

	rawResp, readErr := readResponseLang(rw, l)
	res.Response.StatusLine = rawResp.statusLine
	res.Response.StatusCode = parseStatusCode(rawResp.statusLine)
	res.Response.HeadersRaw = strings.Join(rawResp.headerLines, "\r\n")
	res.Response.Bytes = wireSize(rawResp)
	res.Response.BodyRaw = base64.StdEncoding.EncodeToString(rawResp.body)
	res.Response.BodyDecoded = string(decodeBody(rawResp.headerLines, rawResp.body))
	res.Response.DurationMs = time.Since(start).Milliseconds()
	if readErr != nil {
		res.Response.Error = tr(l, "读取响应失败: ", "failed to read the response: ") + readErr.Error()
	}
	return res
}

func setHeader(lines []string, name, value string) []string {
	out := make([]string, 0, len(lines)+1)
	replaced := false
	for _, l := range lines {
		if strings.EqualFold(headerName(l), name) {
			if !replaced {
				out = append(out, name+": "+value)
				replaced = true
			}
			continue
		}
		out = append(out, l)
	}
	if !replaced {
		out = append(out, name+": "+value)
	}
	return out
}

func hasHeader(lines []string, name string) bool {
	for _, l := range lines {
		if strings.EqualFold(headerName(l), name) {
			return true
		}
	}
	return false
}

func headerName(line string) string {
	if i := strings.Index(line, ":"); i > 0 {
		return strings.TrimSpace(line[:i])
	}
	return strings.TrimSpace(line)
}

func headerValue(headerLines []string, name string) string {
	for _, l := range headerLines {
		if i := strings.Index(l, ":"); i > 0 && strings.EqualFold(strings.TrimSpace(l[:i]), name) {
			return strings.TrimSpace(l[i+1:])
		}
	}
	return ""
}

func parseContentLength(headerLines []string) (int, bool) {
	v := headerValue(headerLines, "Content-Length")
	if v == "" {
		return 0, false
	}
	n, err := strconv.Atoi(strings.TrimSpace(v))
	if err != nil || n < 0 {
		return 0, false
	}
	return n, true
}

func parseStatusCode(statusLine string) int {
	fields := strings.Fields(statusLine)
	if len(fields) < 2 {
		return 0
	}
	code, err := strconv.Atoi(fields[1])
	if err != nil {
		return 0
	}
	return code
}

// hostHeaderValue Host 头重写值：非默认端口才带端口，IPv6 加方括号
func hostHeaderValue(host, scheme string, port int) string {
	defaultPort := 80
	if scheme == "https" {
		defaultPort = 443
	}
	h := host
	if strings.Contains(h, ":") && !strings.HasPrefix(h, "[") {
		h = "[" + h + "]"
	}
	if port != defaultPort {
		return net.JoinHostPort(host, strconv.Itoa(port))
	}
	return h
}

func normalizeCRLF(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	return strings.ReplaceAll(s, "\n", "\r\n")
}
