package repeater

import (
	"net"
	"runtime"
	"strconv"
	"strings"
)

// RequestToolInput 右键菜单工具的统一入参（当前报文 + 目标）
type RequestToolInput struct {
	RequestText string `json:"requestText"`
	Scheme      string `json:"scheme"`
	Host        string `json:"host"`
	Port        int    `json:"port"`
	Lang        string `json:"lang"`
}

// RequestToolResult 工具输出（URL / curl 命令 / 转换后的报文）
type RequestToolResult struct {
	Text string `json:"text"`
}

// BuildRequestURL 拼接完整 URL：协议 + Host[:Port] + 路径（含 GET 查询参数）
func BuildRequestURL(in RequestToolInput) (*RequestToolResult, error) {
	l := normalizeLang(in.Lang)
	p, err := parseRequestLang(in.RequestText, l)
	if err != nil {
		return nil, err
	}
	scheme, host, port := resolveRequestTarget(in, p)
	if host == "" {
		return nil, newLangError(l, "目标 Host 为空", "target Host is empty")
	}
	return &RequestToolResult{Text: composeURL(scheme, host, port, p.Path)}, nil
}

// BuildRequestCurl 生成可复制的 curl 命令（Windows 用 cmd 转义，其他平台用 POSIX）
func BuildRequestCurl(in RequestToolInput) (*RequestToolResult, error) {
	platform := "unix"
	if runtime.GOOS == "windows" {
		platform = "windows"
	}
	text, err := buildRequestCurl(in, platform)
	if err != nil {
		return nil, err
	}
	return &RequestToolResult{Text: text}, nil
}

func buildRequestCurl(in RequestToolInput, platform string) (string, error) {
	l := normalizeLang(in.Lang)
	p, err := parseRequestLang(in.RequestText, l)
	if err != nil {
		return "", err
	}
	scheme, host, port := resolveRequestTarget(in, p)
	if host == "" {
		return "", newLangError(l, "目标 Host 为空", "target Host is empty")
	}

	requestURL := composeURL(scheme, host, port, p.Path)
	authority := host
	if strings.Contains(authority, ":") && !strings.HasPrefix(authority, "[") {
		authority = "[" + authority + "]"
	}
	if port > 0 {
		authority = hostHeaderValue(host, scheme, port)
	}

	args := []string{"--url " + curlQuote(requestURL, platform)}
	args = append(args, "-X "+curlQuote(p.Method, platform))

	compressed := false
	for _, line := range p.headerLines {
		name, value, ok := splitHeaderLine(line)
		if !ok {
			args = append(args, "-H "+curlQuote(line, platform))
			continue
		}
		switch strings.ToLower(name) {
		case "content-length", "transfer-encoding":
			continue
		case "accept-encoding":
			if !compressed && value != "" && !strings.EqualFold(strings.TrimSpace(value), "identity") {
				args = append(args, "--compressed")
				compressed = true
			}
			continue
		case "host":
			if strings.EqualFold(strings.TrimSpace(value), authority) {
				continue
			}
		}
		header := name + ": " + value
		if value == "" {
			header = name + ";"
		}
		args = append(args, "-H "+curlQuote(header, platform))
	}

	if p.Body != "" {
		args = append(args, "--data-raw "+curlQuote(p.Body, platform))
	}

	sep := " \\\n  "
	if platform == "windows" {
		sep = " ^\n  "
	}
	if len(args) >= 3 {
		return "curl " + strings.Join(args, sep), nil
	}
	return "curl " + strings.Join(args, " "), nil
}

// ConvertRequestMethod 仿 Burp：GET↔POST，并在 URL 查询串与表单 body 之间搬参数
func ConvertRequestMethod(in RequestToolInput) (*RequestToolResult, error) {
	l := normalizeLang(in.Lang)
	p, err := parseRequestLang(in.RequestText, l)
	if err != nil {
		return nil, err
	}
	method := strings.ToUpper(p.Method)
	if method != "GET" && method != "POST" {
		return nil, newLangErrorf(l,
			"仅支持 GET 与 POST 互转（当前为 %s）",
			"only GET/POST conversion is supported (current method: %s)",
			method)
	}

	path := p.Path
	body := p.Body
	headerLines := append([]string(nil), p.headerLines...)

	if method == "GET" {
		movedQuery := false
		if idx := strings.IndexByte(path, '?'); idx >= 0 {
			query := path[idx+1:]
			path = path[:idx]
			if query != "" {
				switch {
				case body == "":
					body = query
					movedQuery = true
				case isFormContentType(headerValue(headerLines, "Content-Type")):
					body = query + "&" + body
					movedQuery = true
				default:
					// body 不是表单（如 JSON），不把查询串塞进 body，避免破坏内容
					path = path + "?" + query
				}
			}
		}
		if body != "" {
			if movedQuery || !hasHeader(headerLines, "Content-Type") {
				headerLines = setHeader(headerLines, "Content-Type", "application/x-www-form-urlencoded")
			}
			headerLines = setHeader(headerLines, "Content-Length", strconv.Itoa(len(body)))
		}
		method = "POST"
	} else {
		if body != "" && isFormContentType(headerValue(headerLines, "Content-Type")) {
			if strings.Contains(path, "?") {
				path += "&" + body
			} else {
				path += "?" + body
			}
			body = ""
			headerLines = removeHeader(headerLines, "Content-Type")
			headerLines = removeHeader(headerLines, "Content-Length")
		}
		method = "GET"
	}

	var b strings.Builder
	b.WriteString(method + " " + path + " " + p.Version + "\n")
	for _, line := range headerLines {
		b.WriteString(line + "\n")
	}
	b.WriteString("\n")
	b.WriteString(body)
	return &RequestToolResult{Text: b.String()}, nil
}

func resolveRequestTarget(in RequestToolInput, p *parsedRequest) (scheme, host string, port int) {
	scheme = strings.ToLower(strings.TrimSpace(in.Scheme))
	if scheme != "http" && scheme != "https" {
		scheme = p.Target.Scheme
	}
	if scheme != "http" && scheme != "https" {
		scheme = "http"
	}
	host = strings.TrimSpace(in.Host)
	if host == "" {
		host = p.Target.Host
	}
	port = in.Port
	if port <= 0 {
		port = p.Target.Port
	}
	return scheme, host, port
}

func composeURL(scheme, host string, port int, path string) string {
	scheme = strings.ToLower(strings.TrimSpace(scheme))
	if scheme != "http" && scheme != "https" {
		scheme = "http"
	}
	host = strings.TrimSpace(host)
	if strings.HasPrefix(host, "[") && strings.HasSuffix(host, "]") {
		host = strings.TrimSuffix(strings.TrimPrefix(host, "["), "]")
	}

	authority := host
	if strings.Contains(host, ":") {
		authority = "[" + host + "]"
	}
	defaultPort := 80
	if scheme == "https" {
		defaultPort = 443
	}
	if port > 0 && port != defaultPort {
		authority = net.JoinHostPort(host, strconv.Itoa(port))
	}
	if path == "" {
		path = "/"
	}
	if !strings.HasPrefix(path, "/") && path != "*" {
		path = "/" + path
	}
	return scheme + "://" + authority + path
}

func isFormContentType(contentType string) bool {
	contentType = strings.ToLower(strings.TrimSpace(contentType))
	return contentType == "" || strings.HasPrefix(contentType, "application/x-www-form-urlencoded")
}

func removeHeader(lines []string, name string) []string {
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		if strings.EqualFold(headerName(line), name) {
			continue
		}
		out = append(out, line)
	}
	return out
}

func curlQuote(value, platform string) string {
	if platform == "windows" {
		return quoteCurlWindows(value)
	}
	return "'" + strings.ReplaceAll(value, "'", `'\''`) + "'"
}

// quoteCurlWindows 生成 cmd 可用的 ^"..."^" 参数：
// 先按 Windows C 运行库规则转义双引号/反斜杠，再按 cmd 规则给特殊字符加 ^。
func quoteCurlWindows(value string) string {
	var crt strings.Builder
	crt.WriteByte('"')
	backslashes := 0
	flush := func() {
		for i := 0; i < backslashes; i++ {
			crt.WriteByte('\\')
		}
		backslashes = 0
	}
	for _, r := range value {
		switch r {
		case '\\':
			backslashes++
		case '"':
			for i := 0; i < backslashes*2+1; i++ {
				crt.WriteByte('\\')
			}
			crt.WriteByte('"')
			backslashes = 0
		default:
			flush()
			crt.WriteRune(r)
		}
	}
	for i := 0; i < backslashes*2; i++ {
		crt.WriteByte('\\')
	}
	crt.WriteByte('"')

	s := crt.String()
	var b strings.Builder
	b.Grow(len(s) + 8)
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c == '"':
			b.WriteString(`^"`)
		case c == '\r':
			if i+1 < len(s) && s[i+1] == '\n' {
				i++
			}
			b.WriteString("^\n\n")
		case c == '\n':
			b.WriteString("^\n\n")
		case c >= 0x80:
			b.WriteByte(c)
		case isCurlCmdSafe(c):
			b.WriteByte(c)
		case c == '\t' || c < 0x20 || c == 0x7f:
			b.WriteByte(' ')
		default:
			b.WriteByte('^')
			b.WriteByte(c)
		}
	}

	// cmd 的 %VAR% 展开：在 % 后插入 ^ 阻止变量名被识别
	escaped := b.String()
	var out strings.Builder
	out.Grow(len(escaped) + 4)
	for i := 0; i < len(escaped); i++ {
		out.WriteByte(escaped[i])
		if escaped[i] == '%' && i+1 < len(escaped) && isCmdEnvStart(escaped[i+1]) {
			out.WriteByte('^')
		}
	}
	return out.String()
}

func isCurlCmdSafe(c byte) bool {
	if c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' {
		return true
	}
	switch c {
	case ' ', '_', '-', ':', '=', '+', '~', '\'', '/', '.', ',', '?', ';', '*':
		return true
	}
	return false
}

func isCmdEnvStart(c byte) bool {
	return c == '_' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9'
}
