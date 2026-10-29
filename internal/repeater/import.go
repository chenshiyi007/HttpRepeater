package repeater

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/url"
	"strconv"
	"strings"
)

// ImportedRequest 由 curl 命令 / fetch 代码转换出的可编辑请求
type ImportedRequest struct {
	Source             string       `json:"source"`      // curl | fetch
	RequestText        string       `json:"requestText"` // LF 展示格式
	Target             ParsedTarget `json:"target"`
	HostHeaderMismatch bool         `json:"hostHeaderMismatch"` // 报文自带 Host 与 URL 不一致
	Warnings           []string     `json:"warnings"`
}

// ImportRequest 自动识别 curl / fetch 并转换为可编辑报文；默认中文错误
func ImportRequest(text string) (*ImportedRequest, error) {
	return ImportRequestLang(text, "")
}

// ImportRequestLang 按指定语言返回导入错误与警告
func ImportRequestLang(text, language string) (*ImportedRequest, error) {
	l := normalizeLang(language)
	trimmed := strings.TrimSpace(text)
	for {
		changed := false
		for _, prefix := range []string{"$ ", "> ", "» "} {
			if strings.HasPrefix(trimmed, prefix) {
				trimmed = strings.TrimSpace(strings.TrimPrefix(trimmed, prefix))
				changed = true
			}
		}
		if !changed {
			break
		}
	}
	if trimmed == "" {
		return nil, newLangError(l, "不是 curl 命令或 fetch 代码", "not a curl command or fetch code")
	}
	lower := strings.ToLower(trimmed)
	if isCurlPrefix(trimmed) {
		return parseCurl(trimmed, l)
	}
	if strings.Contains(lower, "fetch(") || strings.Contains(lower, "fetch (") {
		return parseFetch(trimmed, l)
	}
	return nil, newLangError(l, "不是 curl 命令或 fetch 代码", "not a curl command or fetch code")
}

func isCurlPrefix(s string) bool {
	if len(s) < 4 || !strings.EqualFold(s[:4], "curl") {
		return false
	}
	if len(s) == 4 {
		return true
	}
	c := s[4]
	return c == ' ' || c == '\t' || c == '\\' || c == '\r' || c == '\n'
}

// ---- curl ----

type curlData struct {
	value     string
	urlencode bool
}

func parseCurl(src string, l lang) (*ImportedRequest, error) {
	words, err := splitShellWords(src, l)
	if err != nil {
		return nil, err
	}
	if len(words) == 0 || !strings.EqualFold(words[0], "curl") {
		return nil, newLangError(l, "不是 curl 命令或 fetch 代码", "not a curl command or fetch code")
	}

	var (
		rawURL     string
		method     string
		methodSet  bool
		head       bool
		compressed bool
		userAgent  string
		referer    string
		cookie     string
		user       string
		headers    []string
		data       []curlData
		warnings   []string
	)

	for i := 1; i < len(words); i++ {
		w := words[i]
		next := func() (string, bool) {
			if i+1 < len(words) {
				i++
				return words[i], true
			}
			return "", false
		}
		switch w {
		case "-X", "--request":
			if v, ok := next(); ok {
				method = strings.ToUpper(v)
				methodSet = true
			}
		case "--url":
			if v, ok := next(); ok {
				rawURL = v
			}
		case "-H", "--header":
			if v, ok := next(); ok {
				headers = append(headers, v)
			}
		case "-d", "--data", "--data-raw", "--data-ascii", "--data-binary":
			if v, ok := next(); ok {
				data = append(data, curlData{value: v})
			}
		case "--data-urlencode":
			if v, ok := next(); ok {
				data = append(data, curlData{value: v, urlencode: true})
			}
		case "--json":
			if v, ok := next(); ok {
				data = append(data, curlData{value: v})
				headers = append(headers, "Content-Type: application/json", "Accept: application/json")
			}
		case "-I", "--head":
			head = true
		case "--compressed":
			compressed = true
		case "-A", "--user-agent":
			if v, ok := next(); ok {
				userAgent = v
			}
		case "-e", "--referer":
			if v, ok := next(); ok {
				referer = v
			}
		case "-b", "--cookie":
			if v, ok := next(); ok {
				cookie = v
			}
		case "-u", "--user":
			if v, ok := next(); ok {
				user = v
			}
		case "-F", "--form", "--form-string", "-T", "--upload-file":
			return nil, newLangErrorf(l,
				"暂不支持 curl 的 %s（multipart / 文件上传），请用 -d 提供 body 或在编辑器里手工构造",
				"curl %s (multipart / file upload) is not supported; use -d for the body or build the request manually",
				w)
		case "--http1.1", "--http1.0", "--http2", "--http2-prior-knowledge",
			"-k", "--insecure", "-L", "--location", "-s", "--silent", "-S", "--show-error",
			"-v", "--verbose", "-i", "--include", "-N", "--no-buffer", "-g", "--globoff",
			"-f", "--fail", "--path-as-is", "--no-progress-meter", "-#", "--progress-bar":
			// 无参数开关，不影响报文
		default:
			if curlFlagTakesValue(w) {
				_, _ = next()
				continue
			}
			if strings.HasPrefix(w, "-") {
				// 未知开关（含 --flag=value 形式）忽略
				continue
			}
			if rawURL == "" && looksLikeURL(w) {
				rawURL = w
			}
		}
	}

	if rawURL == "" {
		return nil, newLangError(l, "curl 命令中未找到 URL", "no URL found in the curl command")
	}
	scheme, host, authority, path, port, err := splitImportURL(rawURL, l)
	if err != nil {
		return nil, err
	}

	// body：多个 -d 以 & 连接；--data-urlencode 对值编码
	var bodyParts []string
	for _, d := range data {
		bodyParts = append(bodyParts, encodeCurlData(d))
	}
	body := strings.Join(bodyParts, "&")

	if !methodSet {
		switch {
		case head:
			method = "HEAD"
		case len(data) > 0:
			method = "POST"
		default:
			method = "GET"
		}
	}

	var extras []string
	if userAgent != "" {
		extras = append(extras, "User-Agent: "+userAgent)
	}
	if referer != "" {
		extras = append(extras, "Referer: "+referer)
	}
	if cookie != "" {
		if strings.Contains(cookie, "=") {
			extras = append(extras, "Cookie: "+cookie)
		} else {
			warnings = append(warnings, tr(l,
				"-b 指向 Cookie 文件，无法读取，已忽略",
				"-b points to a cookie file, which cannot be read; ignored"))
		}
	}
	if user != "" {
		extras = append(extras, "Authorization: Basic "+base64.StdEncoding.EncodeToString([]byte(user)))
	}
	if compressed {
		extras = append(extras, "Accept-Encoding: gzip, deflate, br")
	}

	text, mismatch, headerWarnings := buildImportedText(method, path, authority, headers, extras, body, l)
	warnings = append(warnings, headerWarnings...)
	return &ImportedRequest{
		Source:             "curl",
		RequestText:        text,
		Target:             ParsedTarget{Scheme: scheme, Host: host, Port: port},
		HostHeaderMismatch: mismatch,
		Warnings:           warnings,
	}, nil
}

func encodeCurlData(d curlData) string {
	if !d.urlencode {
		return d.value
	}
	if i := strings.Index(d.value, "="); i >= 0 {
		return d.value[:i] + "=" + url.QueryEscape(d.value[i+1:])
	}
	return url.QueryEscape(d.value)
}

func curlFlagTakesValue(flag string) bool {
	switch flag {
	case "-o", "--output", "-m", "--max-time", "--connect-timeout",
		"-x", "--proxy", "-w", "--write-out", "--cacert", "--capath",
		"--cert", "--key", "-E", "--cert-type", "--key-type", "--pass",
		"--resolve", "--connect-to", "-r", "--range",
		"-c", "--cookie-jar", "-D", "--dump-header", "--retry", "--retry-delay",
		"--retry-max-time", "--limit-rate", "--interface", "--dns-servers",
		"-U", "--proxy-user", "--oauth2-bearer", "--proxy-header", "-Q",
		"--request-target", "--unix-socket", "--abstract-unix-socket", "--noproxy",
		"--keepalive-time", "--local-port", "--max-filesize", "-Y", "--speed-limit",
		"-y", "--speed-time", "-z", "--time-cond", "--url-query", "--aws-sigv4",
		"--hostpubmd5", "--hostpubsha256", "--ciphers", "--tls-max", "-Z", "--parallel-max":
		return true
	}
	return false
}

// splitShellWords 按 POSIX shell 规则拆分 curl 命令（单/双引号、反斜杠续行、$'...'）
func splitShellWords(s string, l lang) ([]string, error) {
	var (
		words   []string
		current strings.Builder
		inWord  bool
	)
	flush := func() {
		if inWord {
			words = append(words, current.String())
			current.Reset()
			inWord = false
		}
	}

	for i := 0; i < len(s); i++ {
		c := s[i]
		switch c {
		case ' ', '\t', '\r', '\n':
			flush()
		case '\\':
			// 续行或转义
			if i+1 < len(s) {
				if s[i+1] == '\n' {
					i++
					continue
				}
				if s[i+1] == '\r' && i+2 < len(s) && s[i+2] == '\n' {
					i += 2
					continue
				}
				inWord = true
				current.WriteByte(s[i+1])
				i++
			}
		case '\'':
			inWord = true
			end := strings.IndexByte(s[i+1:], '\'')
			if end < 0 {
				return nil, newLangError(l, "curl 命令引号未闭合", "unterminated quote in the curl command")
			}
			current.WriteString(s[i+1 : i+1+end])
			i += end + 1
		case '"':
			inWord = true
			i++
			closed := false
			for ; i < len(s); i++ {
				if s[i] == '\\' && i+1 < len(s) {
					switch s[i+1] {
					case '"', '\\', '$', '`':
						current.WriteByte(s[i+1])
						i++
						continue
					case '\n':
						i++
						continue
					}
				}
				if s[i] == '"' {
					closed = true
					break
				}
				current.WriteByte(s[i])
			}
			if !closed {
				return nil, newLangError(l, "curl 命令引号未闭合", "unterminated quote in the curl command")
			}
		case '$':
			if i+1 < len(s) && s[i+1] == '\'' {
				// ANSI-C 引号 $'...'
				inWord = true
				i += 2
				closed := false
				for ; i < len(s); i++ {
					if s[i] == '\\' && i+1 < len(s) {
						switch s[i+1] {
						case 'n':
							current.WriteByte('\n')
						case 'r':
							current.WriteByte('\r')
						case 't':
							current.WriteByte('\t')
						case '\\':
							current.WriteByte('\\')
						case '\'':
							current.WriteByte('\'')
						default:
							current.WriteByte(s[i+1])
						}
						i++
						continue
					}
					if s[i] == '\'' {
						closed = true
						break
					}
					current.WriteByte(s[i])
				}
				if !closed {
					return nil, newLangError(l, "curl 命令引号未闭合", "unterminated quote in the curl command")
				}
				continue
			}
			inWord = true
			current.WriteByte(c)
		default:
			inWord = true
			current.WriteByte(c)
		}
	}
	flush()
	return words, nil
}

func looksLikeURL(s string) bool {
	return strings.Contains(s, "://") || strings.Contains(s, ".") || strings.HasPrefix(strings.ToLower(s), "localhost")
}

// ---- fetch ----

func parseFetch(src string, l lang) (*ImportedRequest, error) {
	idx := indexFetchCall(src)
	if idx < 0 {
		return nil, newLangError(l, "不是 curl 命令或 fetch 代码", "not a curl command or fetch code")
	}
	open := strings.IndexByte(src[idx:], '(')
	if open < 0 {
		return nil, newLangError(l, "fetch 调用格式不正确", "malformed fetch call")
	}
	open += idx
	closeIdx := scanBalanced(src, open)
	if closeIdx < 0 {
		return nil, newLangError(l, "fetch 调用括号不匹配", "unbalanced parentheses in the fetch call")
	}

	args := splitTopLevel(src[open+1:closeIdx], ',')
	if len(args) == 0 {
		return nil, newLangError(l, "fetch 缺少 URL 参数", "the fetch call is missing a URL argument")
	}
	rawURL, err := jsStringLiteral(strings.TrimSpace(args[0]))
	if err != nil || rawURL == "" {
		return nil, newLangError(l,
			"无法从 fetch 代码识别 URL（需为字符串字面量）",
			"could not identify a URL in the fetch code (a string literal is required)")
	}

	var warnings []string
	if strings.Contains(rawURL, "${") {
		warnings = append(warnings, tr(l,
			"URL 含模板变量 ${...}，已按字面导入",
			"the URL contains template variables ${...}; imported literally"))
	}
	scheme, host, authority, path, port, err := splitImportURL(rawURL, l)
	if err != nil {
		return nil, err
	}

	opts := ""
	if len(args) > 1 {
		opts = trimOuterBraces(strings.TrimSpace(args[1]))
	}

	method := "GET"
	methodGiven := false
	if m, ok := fetchStringProp(opts, "method"); ok && m != "" {
		method = strings.ToUpper(m)
		methodGiven = true
	}

	headers, headersFound := fetchHeaders(opts)
	body, hasBody := fetchBody(opts)
	if !hasBody {
		body = ""
	}
	if !methodGiven && (hasBody || body != "") {
		method = "POST"
	}

	if hasBody && body != "" && !headerListHas(headers, "Content-Type") {
		warnings = append(warnings, tr(l,
			"fetch 代码未设置 Content-Type，发送前请按需补充",
			"the fetch code does not set Content-Type; add it if needed"))
	}
	warnings = append(warnings, tr(l,
		"fetch 代码不含浏览器自动附加的 Cookie / Authorization 等头；需要完整报文请用浏览器的「Copy as cURL」",
		"fetch code omits headers added automatically by the browser (Cookie, Authorization, etc.); use the browser's \"Copy as cURL\" for the complete request"))
	if !headersFound && opts != "" {
		// 有 options 但没解析出 headers（可能是变量），给个提示
		warnings = append(warnings, tr(l,
			"未能解析 fetch 的 headers（可能引用了变量），已仅导入 URL/method/body",
			"could not parse fetch headers (they may reference variables); imported URL/method/body only"))
	}

	text, mismatch, headerWarnings := buildImportedText(method, path, authority, headers, nil, body, l)
	warnings = append(warnings, headerWarnings...)
	return &ImportedRequest{
		Source:             "fetch",
		RequestText:        text,
		Target:             ParsedTarget{Scheme: scheme, Host: host, Port: port},
		HostHeaderMismatch: mismatch,
		Warnings:           warnings,
	}, nil
}

func indexFetchCall(s string) int {
	for i := 0; i+5 <= len(s); i++ {
		if s[i:i+5] != "fetch" {
			continue
		}
		if i > 0 && isJSIdentChar(s[i-1]) {
			continue
		}
		j := i + 5
		for j < len(s) && isSpaceByte(s[j]) {
			j++
		}
		if j < len(s) && s[j] == '(' {
			return i
		}
	}
	return -1
}

func trimOuterBraces(s string) string {
	if strings.HasPrefix(s, "{") {
		if end := scanBalanced(s, 0); end >= 0 {
			return s[1:end]
		}
	}
	return s
}

// fetchStringProp 取 options 顶层字符串属性（如 method）
func fetchStringProp(opts, key string) (string, bool) {
	expr, ok := fetchPropExpr(opts, key)
	if !ok {
		return "", false
	}
	v, err := jsStringLiteral(expr)
	if err != nil {
		return "", false
	}
	return v, true
}

func fetchPropExpr(opts, key string) (string, bool) {
	idx := findTopLevelKey(opts, key)
	if idx < 0 {
		return "", false
	}
	rest := opts[idx+len(key):]
	colon := strings.IndexByte(rest, ':')
	if colon < 0 {
		return "", false
	}
	rest = strings.TrimSpace(rest[colon+1:])
	if rest == "" {
		return "", false
	}
	if end := indexTopLevel(rest, ','); end >= 0 {
		rest = rest[:end]
	}
	return strings.TrimSpace(rest), true
}

// findTopLevelKey 在对象内部（已去掉外层花括号）找顶层属性名
func findTopLevelKey(s, key string) int {
	depth := 0
	for i := 0; i+len(key) <= len(s); i++ {
		c := s[i]
		switch c {
		case '\'', '"', '`':
			end := skipJSString(s, i)
			if end < 0 {
				return -1
			}
			i = end
			continue
		case '(', '[', '{':
			depth++
			continue
		case ')', ']', '}':
			if depth > 0 {
				depth--
			}
			continue
		}
		if depth != 0 || !strings.HasPrefix(s[i:], key) {
			continue
		}
		if i > 0 && isJSIdentChar(s[i-1]) {
			continue
		}
		after := i + len(key)
		if after < len(s) && isJSIdentChar(s[after]) {
			continue
		}
		return i
	}
	return -1
}

func fetchHeaders(opts string) ([]string, bool) {
	expr, ok := fetchPropExpr(opts, "headers")
	if !ok {
		return nil, false
	}
	if strings.HasPrefix(expr, "new Headers") {
		open := strings.IndexByte(expr, '(')
		if open < 0 {
			return nil, false
		}
		end := scanBalanced(expr, open)
		if end < 0 {
			return nil, false
		}
		expr = strings.TrimSpace(expr[open+1 : end])
	}
	if !strings.HasPrefix(expr, "{") {
		return nil, false
	}
	end := scanBalanced(expr, 0)
	if end < 0 {
		return nil, false
	}
	inner := expr[1:end]

	var headers []string
	for _, part := range splitTopLevel(inner, ',') {
		part = strings.TrimSpace(part)
		if part == "" || strings.HasPrefix(part, "...") {
			continue
		}
		colon := indexTopLevel(part, ':')
		if colon < 0 {
			continue
		}
		keyExpr := strings.TrimSpace(part[:colon])
		valExpr := strings.TrimSpace(part[colon+1:])
		key, err := jsStringLiteral(keyExpr)
		if err != nil {
			if isJSIdentifier(keyExpr) {
				key = keyExpr
			} else {
				continue
			}
		}
		val, err := jsStringLiteral(valExpr)
		if err != nil {
			continue // 非字面量（变量、函数调用）跳过
		}
		headers = append(headers, key+": "+val)
	}
	return headers, true
}

func fetchBody(opts string) (string, bool) {
	expr, ok := fetchPropExpr(opts, "body")
	if !ok {
		return "", false
	}
	if strings.HasPrefix(expr, "JSON.stringify") {
		open := strings.IndexByte(expr, '(')
		if open < 0 {
			return expr, true
		}
		end := scanBalanced(expr, open)
		if end < 0 {
			return expr, true
		}
		inner := expr[open+1 : end]
		if comma := indexTopLevel(inner, ','); comma >= 0 {
			inner = inner[:comma]
		}
		inner = strings.TrimSpace(inner)
		if converted, ok := jsObjectToJSON(inner); ok {
			return converted, true
		}
		if s, err := jsStringLiteral(inner); err == nil {
			return s, true
		}
		return inner, true
	}
	if s, err := jsStringLiteral(expr); err == nil {
		return s, true
	}
	return expr, true
}

// jsObjectToJSON 尽力把 JS 对象字面量转成 JSON（单引号、裸键名、尾逗号）
func jsObjectToJSON(expr string) (string, bool) {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return "", false
	}
	var b strings.Builder
	for i := 0; i < len(expr); {
		c := expr[i]
		switch {
		case c == '\'' || c == '"' || c == '`':
			end := skipJSString(expr, i)
			if end < 0 {
				return "", false
			}
			val, err := jsStringLiteral(expr[i : end+1])
			if err != nil {
				return "", false
			}
			encoded, err := json.Marshal(val)
			if err != nil {
				return "", false
			}
			b.Write(encoded)
			i = end + 1
		case c == ',':
			j := i + 1
			for j < len(expr) && isSpaceByte(expr[j]) {
				j++
			}
			if j < len(expr) && (expr[j] == '}' || expr[j] == ']') {
				i++ // 去掉尾逗号
				continue
			}
			b.WriteByte(c)
			i++
		case isJSIdentStart(c):
			j := i
			for j < len(expr) && isJSIdentChar(expr[j]) {
				j++
			}
			word := expr[i:j]
			k := j
			for k < len(expr) && isSpaceByte(expr[k]) {
				k++
			}
			if k < len(expr) && expr[k] == ':' && word != "true" && word != "false" && word != "null" {
				b.WriteString(`"` + word + `"`)
			} else {
				b.WriteString(word)
			}
			i = j
		default:
			b.WriteByte(c)
			i++
		}
	}

	var v any
	if err := json.Unmarshal([]byte(b.String()), &v); err != nil {
		return "", false
	}
	compact, err := json.Marshal(v)
	if err != nil {
		return "", false
	}
	return string(compact), true
}

// ---- 共用工具 ----

func splitImportURL(raw string, l lang) (scheme, host, authority, path string, port int, err error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", "", "", "", 0, newLangError(l, "URL 为空", "URL is empty")
	}
	if !strings.Contains(raw, "://") {
		raw = "http://" + raw
	}
	u, parseErr := url.Parse(raw)
	if parseErr != nil || u.Hostname() == "" {
		return "", "", "", "", 0, newLangErrorf(l, "URL 格式不正确: %s", "invalid URL: %s", raw)
	}
	scheme = strings.ToLower(u.Scheme)
	if scheme != "http" && scheme != "https" {
		return "", "", "", "", 0, newLangErrorf(l, "暂只支持 http/https，当前为 %s", "only http/https are supported (got %s)", u.Scheme)
	}
	host = u.Hostname()
	if ps := u.Port(); ps != "" {
		port, _ = strconv.Atoi(ps)
	}
	authority = u.Host
	path = u.RequestURI()
	if path == "" {
		path = "/"
	}
	return scheme, host, authority, path, port, nil
}

// buildImportedText 组装报文；Host 缺失时补 URL authority，body 非空补 Content-Length
func buildImportedText(method, path, authority string, userHeaders, extras []string, body string, l lang) (string, bool, []string) {
	var warnings []string
	headers := make([]string, 0, len(userHeaders)+len(extras)+2)

	hostHeader := ""
	for _, h := range userHeaders {
		name, val, ok := splitHeaderLine(h)
		if !ok {
			warnings = append(warnings, tr(l, "忽略无法解析的头: ", "ignored unparseable header: ")+h)
			continue
		}
		if strings.EqualFold(name, "Host") {
			hostHeader = val
		}
		headers = append(headers, name+": "+val)
	}
	mismatch := hostHeader != "" && !strings.EqualFold(hostHeader, authority)
	if hostHeader == "" {
		headers = append([]string{"Host: " + authority}, headers...)
	}

	for _, e := range extras {
		name, _, ok := splitHeaderLine(e)
		if !ok || hasHeader(headers, name) {
			continue
		}
		headers = append(headers, e)
	}

	if body != "" && !hasHeader(headers, "Content-Length") {
		headers = append(headers, "Content-Length: "+strconv.Itoa(len(body)))
	}

	var b strings.Builder
	b.WriteString(method + " " + path + " HTTP/1.1\n")
	for _, h := range headers {
		b.WriteString(h + "\n")
	}
	b.WriteString("\n")
	b.WriteString(body)
	return b.String(), mismatch, warnings
}

func splitHeaderLine(line string) (name, value string, ok bool) {
	line = strings.TrimSpace(line)
	if line == "" {
		return "", "", false
	}
	if i := strings.Index(line, ":"); i > 0 {
		return strings.TrimSpace(line[:i]), strings.TrimSpace(line[i+1:]), true
	}
	// curl 的 "Header;" 表示空值头
	if strings.HasSuffix(line, ";") {
		if name = strings.TrimSpace(strings.TrimSuffix(line, ";")); name != "" {
			return name, "", true
		}
	}
	return "", "", false
}

func headerListHas(headers []string, name string) bool {
	for _, h := range headers {
		if n, _, ok := splitHeaderLine(h); ok && strings.EqualFold(n, name) {
			return true
		}
	}
	return false
}

// scanBalanced 返回与 s[open] 匹配的右括号下标（处理嵌套与 JS 字符串）
func scanBalanced(s string, open int) int {
	if open < 0 || open >= len(s) {
		return -1
	}
	var closeCh byte
	switch s[open] {
	case '(':
		closeCh = ')'
	case '[':
		closeCh = ']'
	case '{':
		closeCh = '}'
	default:
		return -1
	}
	openCh := s[open]
	depth := 0
	for i := open; i < len(s); i++ {
		c := s[i]
		switch c {
		case '\'', '"', '`':
			end := skipJSString(s, i)
			if end < 0 {
				return -1
			}
			i = end
			continue
		}
		if c == openCh {
			depth++
		} else if c == closeCh {
			depth--
			if depth == 0 {
				return i
			}
		}
	}
	return -1
}

func splitTopLevel(s string, sep byte) []string {
	var parts []string
	depth := 0
	start := 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch c {
		case '\'', '"', '`':
			end := skipJSString(s, i)
			if end < 0 {
				return append(parts, s[start:])
			}
			i = end
		case '(', '[', '{':
			depth++
		case ')', ']', '}':
			if depth > 0 {
				depth--
			}
		default:
			if depth == 0 && c == sep {
				parts = append(parts, s[start:i])
				start = i + 1
			}
		}
	}
	parts = append(parts, s[start:])
	return parts
}

func indexTopLevel(s string, sep byte) int {
	depth := 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch c {
		case '\'', '"', '`':
			end := skipJSString(s, i)
			if end < 0 {
				return -1
			}
			i = end
		case '(', '[', '{':
			depth++
		case ')', ']', '}':
			if depth > 0 {
				depth--
			}
		default:
			if depth == 0 && c == sep {
				return i
			}
		}
	}
	return -1
}

// skipJSString 返回以 s[start] 开头的 JS 字符串闭合引号下标
func skipJSString(s string, start int) int {
	if start >= len(s) {
		return -1
	}
	q := s[start]
	for i := start + 1; i < len(s); i++ {
		if s[i] == '\\' {
			i++
			continue
		}
		if s[i] == q {
			return i
		}
	}
	return -1
}

func jsStringLiteral(expr string) (string, error) {
	expr = strings.TrimSpace(expr)
	if len(expr) < 2 {
		return "", errors.New("not a string literal")
	}
	q := expr[0]
	if q != '\'' && q != '"' && q != '`' {
		return "", errors.New("not a string literal")
	}
	var b strings.Builder
	for i := 1; i < len(expr); i++ {
		c := expr[i]
		if c == '\\' && i+1 < len(expr) {
			switch expr[i+1] {
			case 'n':
				b.WriteByte('\n')
			case 'r':
				b.WriteByte('\r')
			case 't':
				b.WriteByte('\t')
			case 'b':
				b.WriteByte('\b')
			case 'f':
				b.WriteByte('\f')
			case 'v':
				b.WriteByte('\v')
			case '0':
				b.WriteByte(0)
			case '\\', '\'', '"', '`':
				b.WriteByte(expr[i+1])
			default:
				b.WriteByte(expr[i+1])
			}
			i++
			continue
		}
		if c == q {
			return b.String(), nil
		}
		b.WriteByte(c)
	}
	return "", errors.New("unterminated string literal")
}

func isJSIdentifier(s string) bool {
	if s == "" || !isJSIdentStart(s[0]) {
		return false
	}
	for i := 1; i < len(s); i++ {
		if !isJSIdentChar(s[i]) {
			return false
		}
	}
	return true
}

func isJSIdentStart(c byte) bool {
	return c == '_' || c == '$' || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

func isJSIdentChar(c byte) bool {
	return isJSIdentStart(c) || (c >= '0' && c <= '9')
}

func isSpaceByte(c byte) bool {
	return c == ' ' || c == '\t' || c == '\r' || c == '\n'
}
