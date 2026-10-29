package repeater

import (
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
)

const maxRequestBytes = 10 << 20 // 单请求 10MB 上限

// ParsedTarget 目标三元组（Port=0 表示协议默认端口）
type ParsedTarget struct {
	Scheme string `json:"scheme"`
	Host   string `json:"host"`
	Port   int    `json:"port"`
}

// ParsedRequest 原始报文的解析结果（前端表单联动用）
type ParsedRequest struct {
	Method    string       `json:"method"`
	Path      string       `json:"path"`
	Version   string       `json:"version"`
	Headers   [][2]string  `json:"headers"`
	Body      string       `json:"body"`
	Target    ParsedTarget `json:"target"`
	IsBareURL bool         `json:"isBareUrl"`
	Skeleton  string       `json:"skeleton"`
}

// parsedRequest 内部完整解析结果：额外保留原始 header 行，发送时按原顺序/大小写/重复头保真写出
type parsedRequest struct {
	ParsedRequest
	headerLines []string
}

// ParseRaw 解析原始请求报文（不发起网络请求，仅用于目标推断），默认中文错误
func ParseRaw(raw string) (*ParsedRequest, error) {
	return ParseRawLang(raw, "")
}

// ParseRawLang 按指定语言返回解析错误
func ParseRawLang(raw, language string) (*ParsedRequest, error) {
	p, err := parseRequestLang(raw, normalizeLang(language))
	if err != nil {
		return nil, err
	}
	return &p.ParsedRequest, nil
}

func parseRequest(raw string) (*parsedRequest, error) {
	return parseRequestLang(raw, langZH)
}

func parseRequestLang(raw string, l lang) (*parsedRequest, error) {
	if len(raw) > maxRequestBytes {
		return nil, newLangErrorf(l, "请求超过大小上限（%d MB）", "request exceeds size limit (%d MB)", maxRequestBytes>>20)
	}
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil, newLangError(l, "请求内容为空", "request is empty")
	}

	// 直接粘贴一个 URL：生成 GET 请求骨架
	if isBareURL(trimmed) {
		return parseBareURL(trimmed, l)
	}

	head, body := splitHeadBody(raw)
	lines := splitLines(head)
	if len(lines) == 0 || strings.TrimSpace(lines[0]) == "" {
		return nil, newLangError(l, "缺少请求行", "missing request line")
	}

	fields := strings.Fields(lines[0])
	if len(fields) < 2 {
		return nil, newLangErrorf(l, "请求行格式不正确: %q", "malformed request line: %q", lines[0])
	}
	method := strings.ToUpper(fields[0])
	requestTarget := fields[1]
	version := "HTTP/1.1"
	if len(fields) >= 3 {
		version = fields[2]
		if !strings.HasPrefix(strings.ToUpper(version), "HTTP/") {
			return nil, newLangErrorf(l, "请求行版本格式不正确: %q", "malformed HTTP version: %q", version)
		}
	}

	headerLines := make([]string, 0, len(lines)-1)
	for _, l := range lines[1:] {
		if strings.TrimSpace(l) == "" {
			continue
		}
		headerLines = append(headerLines, l)
	}
	pairs := parseHeaderPairs(headerLines)

	// 请求行可能是 absolute-form（代理式写法 http(s)://host:port/path）
	scheme := ""
	path := requestTarget
	absHost := ""
	absPort := 0
	lowerTarget := strings.ToLower(requestTarget)
	if strings.HasPrefix(lowerTarget, "http://") || strings.HasPrefix(lowerTarget, "https://") {
		u, err := url.Parse(requestTarget)
		if err != nil || u.Hostname() == "" {
			return nil, newLangErrorf(l, "无法解析绝对 URL: %q", "cannot parse absolute URL: %q", requestTarget)
		}
		scheme = strings.ToLower(u.Scheme)
		absHost = u.Hostname()
		if ps := u.Port(); ps != "" {
			absPort, _ = strconv.Atoi(ps)
		}
		path = u.RequestURI()
		if path == "" {
			path = "/"
		}
	}

	hostHeader, _ := findHeader(pairs, "Host")
	host, hostPort := splitHostPortValue(hostHeader)
	if absHost != "" {
		host = absHost
	}

	port := 0
	switch {
	case absPort > 0:
		port = absPort
	case hostPort > 0:
		port = hostPort
	}
	if scheme == "" {
		// 推断 scheme：显式端口 443 → https，否则 http
		if port == 443 {
			scheme = "https"
		} else {
			scheme = "http"
		}
	}

	return &parsedRequest{
		ParsedRequest: ParsedRequest{
			Method:  method,
			Path:    path,
			Version: version,
			Headers: pairs,
			Body:    body,
			Target:  ParsedTarget{Scheme: scheme, Host: host, Port: port},
		},
		headerLines: headerLines,
	}, nil
}

func isBareURL(s string) bool {
	if strings.ContainsAny(s, "\r\n") {
		return false
	}
	lower := strings.ToLower(s)
	return strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://")
}

func parseBareURL(s string, l lang) (*parsedRequest, error) {
	u, err := url.Parse(s)
	if err != nil || u.Hostname() == "" {
		return nil, newLangError(l, "URL 格式不正确", "invalid URL")
	}
	scheme := strings.ToLower(u.Scheme)
	if scheme != "http" && scheme != "https" {
		return nil, newLangErrorf(l, "不支持的协议: %s", "unsupported scheme: %s", scheme)
	}
	port := 0
	if ps := u.Port(); ps != "" {
		if n, err := strconv.Atoi(ps); err == nil {
			port = n
		}
	}
	path := u.RequestURI()
	if path == "" {
		path = "/"
	}
	hostHeader := u.Host // 含显式端口（如有）
	headerLines := []string{"Host: " + hostHeader}
	skeleton := fmt.Sprintf("GET %s HTTP/1.1\nHost: %s\n\n", path, hostHeader)

	return &parsedRequest{
		ParsedRequest: ParsedRequest{
			Method:    "GET",
			Path:      path,
			Version:   "HTTP/1.1",
			Headers:   [][2]string{{"Host", hostHeader}},
			Target:    ParsedTarget{Scheme: scheme, Host: u.Hostname(), Port: port},
			IsBareURL: true,
			Skeleton:  skeleton,
		},
		headerLines: headerLines,
	}, nil
}

// splitHeadBody 以第一个空行为界切分 head / body
func splitHeadBody(raw string) (head, body string) {
	if i := strings.Index(raw, "\r\n\r\n"); i >= 0 {
		return raw[:i], raw[i+4:]
	}
	if i := strings.Index(raw, "\n\n"); i >= 0 {
		return raw[:i], raw[i+2:]
	}
	return raw, ""
}

func splitLines(s string) []string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	lines := strings.Split(s, "\n")
	for i, l := range lines {
		lines[i] = strings.TrimSuffix(l, "\r")
	}
	return lines
}

func parseHeaderPairs(lines []string) [][2]string {
	pairs := make([][2]string, 0, len(lines))
	for _, l := range lines {
		l = strings.TrimSuffix(l, "\r")
		if strings.TrimSpace(l) == "" {
			continue
		}
		if i := strings.Index(l, ":"); i > 0 {
			pairs = append(pairs, [2]string{strings.TrimSpace(l[:i]), strings.TrimSpace(l[i+1:])})
		} else {
			pairs = append(pairs, [2]string{strings.TrimSpace(l), ""})
		}
	}
	return pairs
}

func findHeader(pairs [][2]string, name string) (string, bool) {
	for _, p := range pairs {
		if strings.EqualFold(p[0], name) {
			return p[1], true
		}
	}
	return "", false
}

// splitHostPortValue 解析 Host 头值：host[:port] / [v6]:port / [v6]
func splitHostPortValue(v string) (string, int) {
	v = strings.TrimSpace(v)
	if v == "" {
		return "", 0
	}
	if host, portStr, err := net.SplitHostPort(v); err == nil {
		p, _ := strconv.Atoi(portStr)
		return host, p
	}
	return strings.Trim(v, "[]"), 0
}
