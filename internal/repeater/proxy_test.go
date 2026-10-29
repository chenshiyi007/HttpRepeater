package repeater

import (
	"bufio"
	"bytes"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"
)

// startForwardProxy 启动最小 HTTP 正向代理：普通请求走绝对 URL 转发，CONNECT 走隧道
func startForwardProxy(t *testing.T, plainCaptured *[]byte, connectTarget *string) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("proxy listen: %v", err)
	}
	t.Cleanup(func() { ln.Close() })
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go handleProxyConn(conn, plainCaptured, connectTarget)
		}
	}()
	return ln.Addr().String()
}

func handleProxyConn(client net.Conn, plainCaptured *[]byte, connectTarget *string) {
	defer client.Close()
	_ = client.SetDeadline(time.Now().Add(5 * time.Second))
	req := readOneRequest(client)
	headEnd := bytes.Index(req, []byte("\r\n\r\n"))
	if headEnd < 0 {
		return
	}
	head := string(req[:headEnd+4])
	body := req[headEnd+4:]
	lines := strings.Split(strings.TrimSuffix(head, "\r\n\r\n"), "\r\n")
	fields := strings.Fields(lines[0])
	if len(fields) < 2 {
		return
	}

	if fields[0] == "CONNECT" {
		if connectTarget != nil {
			*connectTarget = fields[1]
		}
		if _, err := client.Write([]byte("HTTP/1.1 200 Connection established\r\n\r\n")); err != nil {
			return
		}
		up, err := net.Dial("tcp", fields[1])
		if err != nil {
			return
		}
		defer up.Close()
		_ = client.SetDeadline(time.Time{})
		_ = up.SetDeadline(time.Time{})
		go func() { _, _ = io.Copy(up, client) }()
		_, _ = io.Copy(client, up)
		return
	}

	if plainCaptured != nil {
		*plainCaptured = req
	}
	u, err := url.Parse(fields[1])
	if err != nil || u.Host == "" {
		return
	}
	originReq := strings.Replace(head, fields[1], u.RequestURI(), 1) + string(body)
	up, err := net.Dial("tcp", u.Host)
	if err != nil {
		return
	}
	defer up.Close()
	_ = up.SetDeadline(time.Now().Add(5 * time.Second))
	if _, err := up.Write([]byte(originReq)); err != nil {
		return
	}
	_ = client.SetDeadline(time.Time{})
	_, _ = io.Copy(client, up)
}

func TestParseProxyURL(t *testing.T) {
	cases := []struct {
		name    string
		raw     string
		addr    string
		auth    string
		wantErr string
	}{
		{name: "标准 http", raw: "http://127.0.0.1:8080", addr: "127.0.0.1:8080"},
		{name: "缺 scheme 默认 http", raw: "127.0.0.1:8080", addr: "127.0.0.1:8080"},
		{name: "缺端口默认 80", raw: "http://proxy.local", addr: "proxy.local:80"},
		{name: "认证信息", raw: "http://user:pass@127.0.0.1:3128", addr: "127.0.0.1:3128", auth: "Proxy-Authorization: Basic dXNlcjpwYXNz"},
		{name: "空地址", raw: "  ", wantErr: "代理地址为空"},
		{name: "非 http 协议", raw: "socks5://127.0.0.1:1080", wantErr: "暂不支持"},
		{name: "非法端口", raw: "http://127.0.0.1:abc", wantErr: "代理地址格式不正确"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			cfg, err := parseProxyURL(c.raw)
			if c.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), c.wantErr) {
					t.Fatalf("want err contains %q, got %v", c.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
			if cfg.addr != c.addr {
				t.Errorf("addr = %q, want %q", cfg.addr, c.addr)
			}
			if cfg.authLine != c.auth {
				t.Errorf("authLine = %q, want %q", cfg.authLine, c.auth)
			}
		})
	}
}

func TestSendViaProxyHTTP(t *testing.T) {
	resp := []byte("HTTP/1.1 200 OK\r\nContent-Length: 4\r\n\r\npong")
	addr := startRawServer(t, resp, nil)

	var proxyCaptured []byte
	proxyAddr := startForwardProxy(t, &proxyCaptured, nil)

	req := formFor(addr)
	req.UseProxy = true
	req.ProxyURL = "http://" + proxyAddr
	req.Raw = "GET /x?y=1 HTTP/1.1\r\nHost: old.example.com\r\n\r\n"
	res := Send(req)

	if res.Response.Error != "" {
		t.Fatalf("unexpected error: %s", res.Response.Error)
	}
	if res.Response.BodyDecoded != "pong" {
		t.Errorf("proxy response mismatch: %+v", res.Response)
	}
	head := string(proxyCaptured)
	// 绝对 URL 请求行 + Host 跟随目标
	if !strings.HasPrefix(head, "GET http://"+addr+"/x?y=1 HTTP/1.1\r\n") {
		t.Errorf("absolute-form request line expected, captured: %q", head)
	}
	if !strings.Contains(head, "Host: "+addr) {
		t.Errorf("Host header should follow target, captured: %q", head)
	}
}

func TestSendViaProxyAuth(t *testing.T) {
	resp := []byte("HTTP/1.1 200 OK\r\nContent-Length: 2\r\n\r\nok")
	addr := startRawServer(t, resp, nil)

	var proxyCaptured []byte
	proxyAddr := startForwardProxy(t, &proxyCaptured, nil)

	req := formFor(addr)
	req.UseProxy = true
	req.ProxyURL = "http://user:pass@" + proxyAddr
	req.Raw = "GET / HTTP/1.1\r\nHost: x\r\n\r\n"
	res := Send(req)

	if res.Response.Error != "" {
		t.Fatalf("unexpected error: %s", res.Response.Error)
	}
	if !strings.Contains(string(proxyCaptured), "Proxy-Authorization: Basic dXNlcjpwYXNz") {
		t.Errorf("proxy auth header missing, captured: %q", string(proxyCaptured))
	}
}

func TestSendViaProxyHTTPSConnect(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("tunnel-ok"))
	}))
	defer srv.Close()

	addr := strings.TrimPrefix(srv.URL, "https://")
	host, portStr, _ := net.SplitHostPort(addr)
	port, _ := strconv.Atoi(portStr)

	var connectTarget string
	proxyAddr := startForwardProxy(t, nil, &connectTarget)

	req := SendRequest{
		Raw:               "GET / HTTP/1.1\r\nHost: x\r\n\r\n",
		Scheme:            "https",
		Host:              host,
		Port:              port,
		UseProxy:          true,
		ProxyURL:          "http://" + proxyAddr,
		FollowHostHeader:  true,
		AutoContentLength: true,
		VerifyTLS:         false,
		TimeoutMs:         5000,
	}
	res := Send(req)

	if res.Response.Error != "" {
		t.Fatalf("unexpected error: %s", res.Response.Error)
	}
	if res.Response.BodyDecoded != "tunnel-ok" {
		t.Errorf("tunnel response mismatch: %+v", res.Response)
	}
	if connectTarget != net.JoinHostPort(host, portStr) {
		t.Errorf("CONNECT target = %q, want %q", connectTarget, net.JoinHostPort(host, portStr))
	}
}

func TestSendProxyUnreachable(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	proxyAddr := ln.Addr().String()
	_ = ln.Close() // 占住再释放，端口不可达

	req := formFor("127.0.0.1:1")
	req.UseProxy = true
	req.ProxyURL = "http://" + proxyAddr
	req.Raw = "GET / HTTP/1.1\r\nHost: x\r\n\r\n"
	res := Send(req)

	if res.Response.Error == "" {
		t.Fatal("expected proxy dial error")
	}
	if !strings.Contains(res.Response.Error, "连接代理失败") {
		t.Errorf("unfriendly error: %q", res.Response.Error)
	}
	if res.Response.StatusCode != 0 {
		t.Errorf("status should be 0 on proxy error, got %d", res.Response.StatusCode)
	}
}

func TestSendProxyConnectRejected(t *testing.T) {
	// 代理对 CONNECT 一律回 407
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { ln.Close() })
	go func() {
		for {
			conn, err := ln.Accept()
			if err != nil {
				return
			}
			go func(c net.Conn) {
				defer c.Close()
				_ = c.SetDeadline(time.Now().Add(5 * time.Second))
				br := bufio.NewReader(c)
				_, _ = br.ReadString('\n') // CONNECT 行，直接拒绝
				_, _ = c.Write([]byte("HTTP/1.1 407 Proxy Authentication Required\r\nContent-Length: 0\r\n\r\n"))
			}(conn)
		}
	}()

	req := SendRequest{
		Raw:               "GET / HTTP/1.1\r\nHost: x\r\n\r\n",
		Scheme:            "https",
		Host:              "example.com",
		Port:              443,
		UseProxy:          true,
		ProxyURL:          "http://" + ln.Addr().String(),
		FollowHostHeader:  true,
		AutoContentLength: true,
		TimeoutMs:         5000,
	}
	res := Send(req)

	if res.Response.Error == "" {
		t.Fatal("expected CONNECT failure")
	}
	if !strings.Contains(res.Response.Error, "代理拒绝建立隧道") || !strings.Contains(res.Response.Error, "407") {
		t.Errorf("unfriendly error: %q", res.Response.Error)
	}
}

func TestSendProxyBadURL(t *testing.T) {
	req := formFor("127.0.0.1:1")
	req.UseProxy = true
	req.Raw = "GET / HTTP/1.1\r\nHost: x\r\n\r\n"

	req.ProxyURL = ""
	res := Send(req)
	if !strings.Contains(res.Response.Error, "代理地址为空") {
		t.Errorf("empty proxyURL error: %q", res.Response.Error)
	}

	req.ProxyURL = "socks5://127.0.0.1:1080"
	res = Send(req)
	if !strings.Contains(res.Response.Error, "暂不支持") {
		t.Errorf("bad scheme error: %q", res.Response.Error)
	}
}
