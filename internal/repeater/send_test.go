package repeater

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"
)

// startRawServer 启动一次性 raw TCP 服务：读取一个完整请求后写死响应
func startRawServer(t *testing.T, response []byte, captured *[]byte) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	t.Cleanup(func() { ln.Close() })
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		_ = conn.SetDeadline(time.Now().Add(5 * time.Second))
		req := readOneRequest(conn)
		if captured != nil {
			*captured = req
		}
		_, _ = conn.Write(response)
	}()
	return ln.Addr().String()
}

// readOneRequest 读取请求头 + Content-Length 指定的 body（不等 EOF，避免死锁）
func readOneRequest(r io.Reader) []byte {
	br := bufio.NewReader(r)
	var head bytes.Buffer
	for {
		line, err := br.ReadString('\n')
		if err != nil {
			break
		}
		head.WriteString(line)
		if line == "\r\n" || line == "\n" {
			break
		}
	}
	cl := 0
	for _, line := range strings.Split(head.String(), "\n") {
		if strings.HasPrefix(strings.ToLower(line), "content-length:") {
			v := strings.TrimSpace(strings.SplitN(line, ":", 2)[1])
			cl, _ = strconv.Atoi(v)
		}
	}
	body := make([]byte, cl)
	_, _ = io.ReadFull(br, body)
	return append(head.Bytes(), body...)
}

func formFor(addr string) SendRequest {
	host, portStr, _ := net.SplitHostPort(addr)
	port, _ := strconv.Atoi(portStr)
	return SendRequest{
		Scheme:            "http",
		Host:              host,
		Port:              port,
		FollowHostHeader:  true,
		AutoContentLength: true,
		TimeoutMs:         3000,
	}
}

func TestSendContentLengthResponse(t *testing.T) {
	var captured []byte
	resp := []byte("HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: 5\r\n\r\nhello")
	addr := startRawServer(t, resp, &captured)

	req := formFor(addr)
	req.Raw = "GET /x HTTP/1.1\r\nHost: old.example.com\r\n\r\n"
	res := Send(req)

	if res.Response.Error != "" {
		t.Fatalf("unexpected error: %s", res.Response.Error)
	}
	if res.Response.StatusCode != 200 || res.Response.BodyDecoded != "hello" {
		t.Errorf("unexpected response: %+v", res.Response)
	}
	if !strings.Contains(res.Response.HeadersRaw, "Content-Type: text/plain") {
		t.Errorf("headersRaw missing content-type: %q", res.Response.HeadersRaw)
	}
	// Host 头跟随目标重写
	if !strings.Contains(string(captured), "Host: "+addr) {
		t.Errorf("Host header not rewritten, captured: %q", string(captured))
	}
	if strings.Contains(string(captured), "old.example.com") {
		t.Errorf("original Host should be replaced, captured: %q", string(captured))
	}
	if res.Response.DurationMs < 0 || res.Response.Bytes <= 0 {
		t.Errorf("bad metrics: %+v", res.Response)
	}
}

func TestSendChunked(t *testing.T) {
	var captured []byte
	resp := []byte("HTTP/1.1 200 OK\r\nTransfer-Encoding: chunked\r\n\r\n4\r\nWiki\r\n5\r\npedia\r\n0\r\n\r\n")
	addr := startRawServer(t, resp, &captured)

	req := formFor(addr)
	req.Raw = "GET / HTTP/1.1\r\nHost: x\r\n\r\n"
	res := Send(req)
	if res.Response.Error != "" {
		t.Fatalf("unexpected error: %s", res.Response.Error)
	}
	if res.Response.BodyDecoded != "Wikipedia" {
		t.Errorf("chunked decode mismatch: %q", res.Response.BodyDecoded)
	}
}

func TestSendChunkedWithExtension(t *testing.T) {
	var captured []byte
	resp := []byte("HTTP/1.1 200 OK\r\nTransfer-Encoding: chunked\r\n\r\n5;foo=bar\r\nhello\r\n0\r\n\r\n")
	addr := startRawServer(t, resp, &captured)

	req := formFor(addr)
	req.Raw = "GET / HTTP/1.1\r\nHost: x\r\n\r\n"
	res := Send(req)
	if res.Response.Error != "" {
		t.Fatalf("unexpected error: %s", res.Response.Error)
	}
	if res.Response.BodyDecoded != "hello" {
		t.Errorf("chunked extension decode mismatch: %q", res.Response.BodyDecoded)
	}
}

func TestSendEOFBody(t *testing.T) {
	var captured []byte
	resp := []byte("HTTP/1.0 200 OK\r\nContent-Type: text/plain\r\n\r\nstream-end")
	addr := startRawServer(t, resp, &captured)

	req := formFor(addr)
	req.Raw = "GET / HTTP/1.1\r\nHost: x\r\n\r\n"
	res := Send(req)
	if res.Response.Error != "" {
		t.Fatalf("unexpected error: %s", res.Response.Error)
	}
	if res.Response.BodyDecoded != "stream-end" {
		t.Errorf("EOF body mismatch: %q", res.Response.BodyDecoded)
	}
}

func TestSendGzipDecode(t *testing.T) {
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	_, _ = gw.Write([]byte(`{"msg":"gzip ok"}`))
	_ = gw.Close()

	var captured []byte
	resp := append([]byte("HTTP/1.1 200 OK\r\nContent-Encoding: gzip\r\nContent-Length: "+strconv.Itoa(buf.Len())+"\r\n\r\n"), buf.Bytes()...)
	addr := startRawServer(t, resp, &captured)

	req := formFor(addr)
	req.Raw = "GET / HTTP/1.1\r\nHost: x\r\n\r\n"
	res := Send(req)
	if res.Response.Error != "" {
		t.Fatalf("unexpected error: %s", res.Response.Error)
	}
	if res.Response.BodyDecoded != `{"msg":"gzip ok"}` {
		t.Errorf("gzip decode mismatch: %q", res.Response.BodyDecoded)
	}
	if res.Response.BodyRaw == "" {
		t.Error("raw body should be preserved as base64")
	}
}

func TestSendContentLengthRecompute(t *testing.T) {
	var captured []byte
	resp := []byte("HTTP/1.1 200 OK\r\nContent-Length: 2\r\n\r\nok")
	addr := startRawServer(t, resp, &captured)

	req := formFor(addr)
	req.Raw = "POST /upload HTTP/1.1\r\nHost: x\r\nContent-Length: 1\r\n\r\nabcdef"
	res := Send(req)
	if res.Response.Error != "" {
		t.Fatalf("unexpected error: %s", res.Response.Error)
	}
	if !strings.Contains(string(captured), "Content-Length: 6") {
		t.Errorf("Content-Length not recomputed, captured: %q", string(captured))
	}
	if strings.Contains(string(captured), "Content-Length: 1\r\n") {
		t.Errorf("stale Content-Length still present, captured: %q", string(captured))
	}
}

func TestSendAddsMissingContentLength(t *testing.T) {
	var captured []byte
	resp := []byte("HTTP/1.1 200 OK\r\nContent-Length: 2\r\n\r\nok")
	addr := startRawServer(t, resp, &captured)

	req := formFor(addr)
	req.Raw = "POST /x HTTP/1.1\r\nHost: x\r\n\r\nabc"
	_ = Send(req)
	if !strings.Contains(string(captured), "Content-Length: 3") {
		t.Errorf("missing Content-Length not added, captured: %q", string(captured))
	}
}

func TestSendFollowHostHeaderOff(t *testing.T) {
	var captured []byte
	resp := []byte("HTTP/1.1 200 OK\r\nContent-Length: 2\r\n\r\nok")
	addr := startRawServer(t, resp, &captured)

	req := formFor(addr)
	req.FollowHostHeader = false
	req.Raw = "GET / HTTP/1.1\r\nHost: original.com\r\n\r\n"
	_ = Send(req)
	if !strings.Contains(string(captured), "Host: original.com") {
		t.Errorf("Host should be kept as-is, captured: %q", string(captured))
	}
}

func TestSendTLSInsecure(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte("tls-ok"))
	}))
	defer srv.Close()

	addr := strings.TrimPrefix(srv.URL, "https://")
	host, portStr, _ := net.SplitHostPort(addr)
	port, _ := strconv.Atoi(portStr)

	req := SendRequest{
		Raw:               "GET / HTTP/1.1\r\nHost: ignored\r\n\r\n",
		Scheme:            "https",
		Host:              host,
		Port:              port,
		FollowHostHeader:  true,
		AutoContentLength: true,
		VerifyTLS:         false,
		TimeoutMs:         5000,
	}
	res := Send(req)
	if res.Response.Error != "" {
		t.Fatalf("unexpected error: %s", res.Response.Error)
	}
	if res.Response.StatusCode != 200 || res.Response.BodyDecoded != "tls-ok" {
		t.Errorf("TLS response mismatch: %+v", res.Response)
	}
}

func TestSendConnectionRefused(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := ln.Addr().String()
	_ = ln.Close() // 关闭后端口不可达

	req := formFor(addr)
	req.Raw = "GET / HTTP/1.1\r\nHost: x\r\n\r\n"
	res := Send(req)
	if res.Response.Error == "" {
		t.Fatal("expected connection error")
	}
	if !strings.Contains(res.Response.Error, "连接失败") {
		t.Errorf("unexpected error message: %q", res.Response.Error)
	}
	if res.Response.StatusCode != 0 {
		t.Errorf("status should be 0 on transport error, got %d", res.Response.StatusCode)
	}
}

func TestSendCancelled(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer ln.Close()

	started := make(chan struct{})
	go func() {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		buf := make([]byte, 1024)
		_, _ = conn.Read(buf)
		close(started)
		_, _ = io.Copy(io.Discard, conn)
	}()

	port := ln.Addr().(*net.TCPAddr).Port
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan *SendResult, 1)
	go func() {
		done <- sendWithContext(ctx, SendRequest{
			Raw:       "GET / HTTP/1.1\nHost: 127.0.0.1\n\n",
			Scheme:    "http",
			Host:      "127.0.0.1",
			Port:      port,
			TimeoutMs: 10000,
		})
	}()

	select {
	case <-started:
	case <-time.After(3 * time.Second):
		t.Fatal("request did not start")
	}
	cancel()

	select {
	case res := <-done:
		if res.Response == nil {
			t.Fatal("missing response")
		}
		if res.Response.Error != "已取消" {
			t.Fatalf("error = %q, want 已取消", res.Response.Error)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("send was not cancelled")
	}
}
