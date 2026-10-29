package repeater

import (
	"strconv"
	"strings"
	"testing"
)

func TestBuildRequestURL(t *testing.T) {
	raw := "GET /api/list?a=1&b=2 HTTP/1.1\nHost: example.com\n\n"
	got, err := BuildRequestURL(RequestToolInput{RequestText: raw, Scheme: "https"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Text != "https://example.com/api/list?a=1&b=2" {
		t.Errorf("url = %q", got.Text)
	}
}

func TestBuildRequestURLFormTargetWins(t *testing.T) {
	raw := "GET /x?q=1 HTTP/1.1\nHost: old.test\n\n"
	got, err := BuildRequestURL(RequestToolInput{
		RequestText: raw,
		Scheme:      "http",
		Host:        "127.0.0.1",
		Port:        8080,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Text != "http://127.0.0.1:8080/x?q=1" {
		t.Errorf("url = %q", got.Text)
	}
}

func TestBuildRequestURLIPv6(t *testing.T) {
	raw := "GET / HTTP/1.1\nHost: [::1]:8080\n\n"
	got, err := BuildRequestURL(RequestToolInput{RequestText: raw})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Text != "http://[::1]:8080/" {
		t.Errorf("url = %q", got.Text)
	}
}

func TestBuildRequestCurlPosix(t *testing.T) {
	raw := "POST /api?a=1 HTTP/1.1\n" +
		"Host: example.com\n" +
		"Content-Type: application/json\n" +
		"Content-Length: 7\n" +
		"User-Agent: UA\n\n" +
		`{"a":1}`
	got, err := buildRequestCurl(RequestToolInput{
		RequestText: raw,
		Scheme:      "https",
		Host:        "example.com",
	}, "unix")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "curl --url 'https://example.com/api?a=1' \\\n" +
		"  -X 'POST' \\\n" +
		"  -H 'Content-Type: application/json' \\\n" +
		"  -H 'User-Agent: UA' \\\n" +
		"  --data-raw '{\"a\":1}'"
	if got != want {
		t.Errorf("curl = %q, want %q", got, want)
	}
}

func TestBuildRequestCurlWindowsRoundTrip(t *testing.T) {
	body := "{\"a\":\"x&y|z<q>^r%20s\\t\",\"b\":\"(hi)\",\"c\":\"line1\nline2\",\"d\":\"中文\"}"
	raw := "POST /api?x=1&y=2 HTTP/1.1\n" +
		"Host: example.com\n" +
		"Content-Type: application/json\n" +
		"User-Agent: test\n\n" + body

	curl, err := buildRequestCurl(RequestToolInput{
		RequestText: raw,
		Scheme:      "https",
		Host:        "example.com",
	}, "windows")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(curl, `^"`) || !strings.Contains(curl, "^\n") {
		t.Fatalf("not a cmd-style curl: %q", curl)
	}

	res, err := ImportRequest(curl)
	if err != nil {
		t.Fatalf("re-import failed: %v\ncurl: %s", err, curl)
	}
	want := "POST /api?x=1&y=2 HTTP/1.1\n" +
		"Host: example.com\n" +
		"Content-Type: application/json\n" +
		"User-Agent: test\n" +
		"Content-Length: " + strconv.Itoa(len(body)) + "\n\n" +
		body
	if res.RequestText != want {
		t.Errorf("round trip mismatch:\n got %q\nwant %q", res.RequestText, want)
	}
}

func TestConvertMethodGetToPost(t *testing.T) {
	raw := "GET /search?q=a&b=2 HTTP/1.1\nHost: example.com\nUser-Agent: x\n\n"
	got, err := ConvertRequestMethod(RequestToolInput{RequestText: raw})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "POST /search HTTP/1.1\n" +
		"Host: example.com\n" +
		"User-Agent: x\n" +
		"Content-Type: application/x-www-form-urlencoded\n" +
		"Content-Length: 7\n\n" +
		"q=a&b=2"
	if got.Text != want {
		t.Errorf("converted = %q, want %q", got.Text, want)
	}
}

func TestConvertMethodPostFormToGet(t *testing.T) {
	raw := "POST /login HTTP/1.1\n" +
		"Host: example.com\n" +
		"Content-Type: application/x-www-form-urlencoded\n" +
		"Content-Length: 7\n\n" +
		"a=1&b=2"
	got, err := ConvertRequestMethod(RequestToolInput{RequestText: raw})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "GET /login?a=1&b=2 HTTP/1.1\nHost: example.com\n\n"
	if got.Text != want {
		t.Errorf("converted = %q, want %q", got.Text, want)
	}
}

func TestConvertMethodPostJSONKeepsBody(t *testing.T) {
	raw := "POST /api HTTP/1.1\n" +
		"Host: x\n" +
		"Content-Type: application/json\n" +
		"Content-Length: 7\n\n" +
		`{"a":1}`
	got, err := ConvertRequestMethod(RequestToolInput{RequestText: raw})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "GET /api HTTP/1.1\n" +
		"Host: x\n" +
		"Content-Type: application/json\n" +
		"Content-Length: 7\n\n" +
		`{"a":1}`
	if got.Text != want {
		t.Errorf("converted = %q, want %q", got.Text, want)
	}
}

func TestConvertMethodUnsupported(t *testing.T) {
	_, err := ConvertRequestMethod(RequestToolInput{RequestText: "PUT / HTTP/1.1\nHost: x\n\n"})
	if err == nil || !strings.Contains(err.Error(), "仅支持") {
		t.Fatalf("want unsupported method error, got %v", err)
	}
}

func TestCurlQuoteWindowsEscapesPercentBeforeVariable(t *testing.T) {
	got := quoteCurlWindows("%PATH%")
	if !strings.Contains(got, "%^P") {
		t.Errorf("percent should be escaped before a variable name: %q", got)
	}
	res := normalizeWindowsCmd(got)
	if res != "\"%PATH%\"" {
		t.Errorf("normalized = %q", res)
	}
}
