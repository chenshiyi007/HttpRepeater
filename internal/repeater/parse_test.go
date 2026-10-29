package repeater

import (
	"strings"
	"testing"
)

func TestParseRawBasic(t *testing.T) {
	raw := "GET /api/list?a=1 HTTP/1.1\r\nHost: example.com\r\nAccept: */*\r\n\r\n"
	p, err := ParseRaw(raw)
	if err != nil {
		t.Fatalf("ParseRaw error: %v", err)
	}
	if p.Method != "GET" || p.Path != "/api/list?a=1" || p.Version != "HTTP/1.1" {
		t.Errorf("request line mismatch: %s %s %s", p.Method, p.Path, p.Version)
	}
	if p.Target.Scheme != "http" || p.Target.Host != "example.com" || p.Target.Port != 0 {
		t.Errorf("target mismatch: %+v", p.Target)
	}
	if len(p.Headers) != 2 || p.Headers[0] != [2]string{"Host", "example.com"} {
		t.Errorf("headers mismatch: %v", p.Headers)
	}
	if p.IsBareURL {
		t.Error("should not be bare URL")
	}
}

func TestParseRawHostPort443(t *testing.T) {
	p, err := ParseRaw("GET / HTTP/1.1\nHost: example.com:443\n\n")
	if err != nil {
		t.Fatalf("ParseRaw error: %v", err)
	}
	if p.Target.Scheme != "https" || p.Target.Port != 443 || p.Target.Host != "example.com" {
		t.Errorf("target mismatch: %+v", p.Target)
	}
}

func TestParseRawAbsoluteURL(t *testing.T) {
	raw := "GET https://example.com:8443/x?y=1 HTTP/1.1\r\nHost: example.com\r\n\r\n"
	p, err := ParseRaw(raw)
	if err != nil {
		t.Fatalf("ParseRaw error: %v", err)
	}
	if p.Target.Scheme != "https" || p.Target.Host != "example.com" || p.Target.Port != 8443 {
		t.Errorf("target mismatch: %+v", p.Target)
	}
	if p.Path != "/x?y=1" {
		t.Errorf("path should be origin-form, got %q", p.Path)
	}
}

func TestParseRawBareURL(t *testing.T) {
	p, err := ParseRaw("https://example.com/login?a=1")
	if err != nil {
		t.Fatalf("ParseRaw error: %v", err)
	}
	if !p.IsBareURL {
		t.Error("should be bare URL")
	}
	if p.Target.Scheme != "https" || p.Target.Host != "example.com" || p.Target.Port != 0 {
		t.Errorf("target mismatch: %+v", p.Target)
	}
	if !strings.Contains(p.Skeleton, "GET /login?a=1 HTTP/1.1") {
		t.Errorf("skeleton missing request line: %q", p.Skeleton)
	}
	if !strings.Contains(p.Skeleton, "Host: example.com") {
		t.Errorf("skeleton missing host: %q", p.Skeleton)
	}
}

func TestParseRawBareURLWithPort(t *testing.T) {
	p, err := ParseRaw("http://127.0.0.1:18080/json")
	if err != nil {
		t.Fatalf("ParseRaw error: %v", err)
	}
	if p.Target.Host != "127.0.0.1" || p.Target.Port != 18080 || p.Target.Scheme != "http" {
		t.Errorf("target mismatch: %+v", p.Target)
	}
	if !strings.Contains(p.Skeleton, "Host: 127.0.0.1:18080") {
		t.Errorf("skeleton host mismatch: %q", p.Skeleton)
	}
}

func TestParseRawNoHost(t *testing.T) {
	p, err := ParseRaw("GET / HTTP/1.1\n\n")
	if err != nil {
		t.Fatalf("ParseRaw error: %v", err)
	}
	if p.Target.Host != "" {
		t.Errorf("host should be empty, got %q", p.Target.Host)
	}
	if len(p.Headers) != 0 {
		t.Errorf("headers should be empty, got %v", p.Headers)
	}
}

func TestParseRawBody(t *testing.T) {
	raw := "POST /x HTTP/1.1\r\nHost: a.com\r\nContent-Length: 3\r\n\r\nabc"
	p, err := ParseRaw(raw)
	if err != nil {
		t.Fatalf("ParseRaw error: %v", err)
	}
	if p.Body != "abc" {
		t.Errorf("body mismatch: %q", p.Body)
	}
}

func TestParseRawInvalid(t *testing.T) {
	if _, err := ParseRaw(""); err == nil {
		t.Error("empty input should error")
	}
	if _, err := ParseRaw("garbage"); err == nil {
		t.Error("single-word input should error")
	}
}

func TestParseRawIPv6Host(t *testing.T) {
	p, err := ParseRaw("GET / HTTP/1.1\r\nHost: [::1]:8080\r\n\r\n")
	if err != nil {
		t.Fatalf("ParseRaw error: %v", err)
	}
	if p.Target.Host != "::1" || p.Target.Port != 8080 {
		t.Errorf("target mismatch: %+v", p.Target)
	}
}
