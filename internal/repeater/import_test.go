package repeater

import (
	"strings"
	"testing"
)

func TestImportCurlBasicGET(t *testing.T) {
	res, err := ImportRequest(`curl 'https://example.com/a?b=1' -H 'Accept: application/json'`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Source != "curl" {
		t.Errorf("source = %q, want curl", res.Source)
	}
	if res.Target.Scheme != "https" || res.Target.Host != "example.com" || res.Target.Port != 0 {
		t.Errorf("target = %+v", res.Target)
	}
	want := "GET /a?b=1 HTTP/1.1\nHost: example.com\nAccept: application/json\n\n"
	if res.RequestText != want {
		t.Errorf("requestText = %q, want %q", res.RequestText, want)
	}
}

func TestImportCurlPostJSON(t *testing.T) {
	res, err := ImportRequest(`curl 'http://127.0.0.1:8080/api' -X POST -H 'Content-Type: application/json' --data-raw '{"a":1}'`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := res.RequestText
	if !strings.HasPrefix(text, "POST /api HTTP/1.1\n") {
		t.Errorf("unexpected request line: %q", text)
	}
	if !strings.Contains(text, "Host: 127.0.0.1:8080\n") {
		t.Errorf("missing Host: %q", text)
	}
	if !strings.Contains(text, "Content-Length: 7\n") {
		t.Errorf("missing Content-Length: %q", text)
	}
	if !strings.HasSuffix(text, "\n\n{\"a\":1}") {
		t.Errorf("unexpected body: %q", text)
	}
	if res.Target.Port != 8080 {
		t.Errorf("port = %d, want 8080", res.Target.Port)
	}
}

func TestImportCurlDefaultPostAndContinuation(t *testing.T) {
	src := "curl 'http://example.com/x' \\\n  -H 'A: 1' \\\n  -d 'hi'"
	res, err := ImportRequest(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasPrefix(res.RequestText, "POST /x HTTP/1.1\n") {
		t.Errorf("unquoted -d should imply POST: %q", res.RequestText)
	}
	if !strings.Contains(res.RequestText, "A: 1\n") {
		t.Errorf("header lost: %q", res.RequestText)
	}
	if !strings.HasSuffix(res.RequestText, "\n\nhi") {
		t.Errorf("body mismatch: %q", res.RequestText)
	}
}

func TestImportCurlJsonFlag(t *testing.T) {
	res, err := ImportRequest(`curl http://api.test/v1 --json '{"k":1}'`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := res.RequestText
	if !strings.HasPrefix(text, "POST /v1 HTTP/1.1\n") {
		t.Errorf("--json should imply POST: %q", text)
	}
	if !strings.Contains(text, "Content-Type: application/json\n") || !strings.Contains(text, "Accept: application/json\n") {
		t.Errorf("json headers missing: %q", text)
	}
	if !strings.HasSuffix(text, "\n\n{\"k\":1}") {
		t.Errorf("body mismatch: %q", text)
	}
}

func TestImportCurlAuthCookieCompressed(t *testing.T) {
	res, err := ImportRequest(`curl https://example.com -u user:pass -b 'a=b; c=d' --compressed`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := res.RequestText
	for _, want := range []string{
		"Authorization: Basic dXNlcjpwYXNz\n",
		"Cookie: a=b; c=d\n",
		"Accept-Encoding: gzip, deflate, br\n",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("missing %q in %q", want, text)
		}
	}
}

func TestImportCurlHead(t *testing.T) {
	res, err := ImportRequest(`curl -I https://example.com`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasPrefix(res.RequestText, "HEAD / HTTP/1.1\n") {
		t.Errorf("HEAD expected: %q", res.RequestText)
	}
}

func TestImportCurlDataURLEncode(t *testing.T) {
	res, err := ImportRequest(`curl http://example.com --data-urlencode 'q=a b'`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasSuffix(res.RequestText, "\n\nq=a+b") {
		t.Errorf("urlencoded body mismatch: %q", res.RequestText)
	}
}

func TestImportCurlHostHeaderMismatch(t *testing.T) {
	res, err := ImportRequest(`curl http://example.com/ -H 'Host: other.test'`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.HostHeaderMismatch {
		t.Error("hostHeaderMismatch should be true")
	}
	if !strings.Contains(res.RequestText, "Host: other.test\n") {
		t.Errorf("custom Host lost: %q", res.RequestText)
	}
	if strings.Contains(res.RequestText, "Host: example.com") {
		t.Errorf("Host should not be duplicated: %q", res.RequestText)
	}
}

func TestImportCurlUnsupportedForm(t *testing.T) {
	_, err := ImportRequest(`curl -F 'file=@a.txt' http://example.com`)
	if err == nil || !strings.Contains(err.Error(), "暂不支持") {
		t.Fatalf("want unsupported -F error, got %v", err)
	}
}

func TestImportCurlPromptPrefix(t *testing.T) {
	res, err := ImportRequest("$ curl https://example.com/path")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Target.Host != "example.com" {
		t.Errorf("host = %q", res.Target.Host)
	}
}

func TestImportDetectNonImportable(t *testing.T) {
	if _, err := ImportRequest("GET / HTTP/1.1\r\nHost: x\r\n\r\n"); err == nil {
		t.Fatal("raw HTTP should not be treated as importable")
	}
}

func TestImportFetchPostJSON(t *testing.T) {
	src := `const res = await fetch("https://api.example.com/v1/items", {
  method: "POST",
  headers: {
    "Content-Type": "application/json",
    "X-Token": "abc",
  },
  body: JSON.stringify({ id: 1, name: 'test', tags: ["a", "b"] }),
  credentials: "include",
})`
	res, err := ImportRequest(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Source != "fetch" {
		t.Errorf("source = %q", res.Source)
	}
	text := res.RequestText
	if !strings.HasPrefix(text, "POST /v1/items HTTP/1.1\n") {
		t.Errorf("request line: %q", text)
	}
	if !strings.Contains(text, "X-Token: abc\n") || !strings.Contains(text, "Content-Type: application/json\n") {
		t.Errorf("headers missing: %q", text)
	}
	if !strings.HasSuffix(text, `{"id":1,"name":"test","tags":["a","b"]}`) {
		t.Errorf("body mismatch: %q", text)
	}
	if len(res.Warnings) == 0 || !strings.Contains(strings.Join(res.Warnings, "；"), "Cookie") {
		t.Errorf("expected cookie warning, got %v", res.Warnings)
	}
}

func TestImportFetchNewHeaders(t *testing.T) {
	src := `fetch('https://x.test/p', { method: 'PUT', headers: new Headers({ 'X-A': 'b' }), body: 'hello' })`
	res, err := ImportRequest(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	text := res.RequestText
	if !strings.HasPrefix(text, "PUT /p HTTP/1.1\n") {
		t.Errorf("method: %q", text)
	}
	if !strings.Contains(text, "X-A: b\n") {
		t.Errorf("header missing: %q", text)
	}
	if !strings.Contains(text, "Content-Length: 5\n") || !strings.HasSuffix(text, "\n\nhello") {
		t.Errorf("body/CL mismatch: %q", text)
	}
}

func TestImportFetchSimpleGET(t *testing.T) {
	res, err := ImportRequest(`fetch('https://x.test/y')`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.RequestText != "GET /y HTTP/1.1\nHost: x.test\n\n" {
		t.Errorf("requestText = %q", res.RequestText)
	}
	if res.Target.Scheme != "https" || res.Target.Host != "x.test" {
		t.Errorf("target = %+v", res.Target)
	}
}

func TestImportFetchTemplateURL(t *testing.T) {
	res, err := ImportRequest("fetch(`https://x.test/${id}/info`)")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(strings.Join(res.Warnings, "；"), "模板变量") {
		t.Errorf("expected template warning, got %v", res.Warnings)
	}
}

// Chrome / Edge 在 Windows 上「Copy as cURL (cmd)」：参数用 ^"..."^" 包裹，
// ^ 续行，内部引号为 ^\^"，特殊字符带 ^ 转义。
func TestImportCurlWindowsCmdChromeStyle(t *testing.T) {
	src := `curl --url ^"https://example.com/api?x=1^&y=2^" ^
  -X ^"POST^" ^
  -H ^"accept: application/json^" ^
  -H ^"content-type: application/json^" ^
  --data-raw ^"^{^\^"id^\^":1,^\^"name^\^":^\^"test^\^"^}"`
	res, err := ImportRequest(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Target.Scheme != "https" || res.Target.Host != "example.com" {
		t.Errorf("target = %+v", res.Target)
	}
	want := "POST /api?x=1&y=2 HTTP/1.1\n" +
		"Host: example.com\n" +
		"accept: application/json\n" +
		"content-type: application/json\n" +
		"Content-Length: 22\n\n" +
		`{"id":1,"name":"test"}`
	if res.RequestText != want {
		t.Errorf("requestText = %q, want %q", res.RequestText, want)
	}
}

// 旧版 Chrome 的 cmd 格式：普通双引号 + ^ 续行
func TestImportCurlWindowsCmdPlainQuotes(t *testing.T) {
	src := `curl.exe "https://example.com/p" ^
  -H "Content-Type: application/json" ^
  --data-raw "{\"a\":1,\"b\":2}"`
	res, err := ImportRequest(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "POST /p HTTP/1.1\n" +
		"Host: example.com\n" +
		"Content-Type: application/json\n" +
		"Content-Length: 13\n\n" +
		`{"a":1,"b":2}`
	if res.RequestText != want {
		t.Errorf("requestText = %q, want %q", res.RequestText, want)
	}
}

func TestImportCurlWindowsCmdSpecialChars(t *testing.T) {
	src := `curl --url ^"http://127.0.0.1:8080/a?q=^$^(x^)^" --data-raw ^"a^&b^|c^<d^>e^^f^"`
	res, err := ImportRequest(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasPrefix(res.RequestText, "POST /a?q=$(x) HTTP/1.1\n") {
		t.Errorf("path mismatch: %q", res.RequestText)
	}
	if !strings.HasSuffix(res.RequestText, "\n\na&b|c<d>e^f") {
		t.Errorf("body mismatch: %q", res.RequestText)
	}
}
