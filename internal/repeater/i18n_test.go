package repeater

import (
	"strings"
	"testing"
)

func TestParseRawErrorMessageLanguage(t *testing.T) {
	cases := []struct {
		lang string
		want string
	}{
		{"", "请求内容为空"},
		{"zh", "请求内容为空"},
		{"en", "request is empty"},
		{"EN", "request is empty"},
	}
	for _, c := range cases {
		_, err := ParseRawLang("", c.lang)
		if err == nil {
			t.Fatalf("lang %q: expected an error", c.lang)
		}
		if err.Error() != c.want {
			t.Errorf("lang %q: got %q, want %q", c.lang, err.Error(), c.want)
		}
	}
}

func TestImportRequestErrorMessageLanguage(t *testing.T) {
	_, err := ImportRequestLang("GET / HTTP/1.1\r\n\r\n", "en")
	if err == nil {
		t.Fatal("expected an error")
	}
	if err.Error() != "not a curl command or fetch code" {
		t.Errorf("got %q", err.Error())
	}
}

func TestProxyURLErrorMessageLanguage(t *testing.T) {
	_, err := parseProxyURLLang("", langEN)
	if err == nil || !strings.Contains(err.Error(), "proxy is enabled") {
		t.Fatalf("got %v", err)
	}
	_, err = parseProxyURL("")
	if err == nil || !strings.Contains(err.Error(), "已启用代理") {
		t.Fatalf("got %v", err)
	}
}
