package repeater

import (
	"bufio"
	"context"
	"encoding/base64"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// proxyConfig 下游代理解析结果（发包出口，非抓包上游）
type proxyConfig struct {
	addr     string // host:port
	authLine string // 完整 "Proxy-Authorization: Basic ..." 行；空串 = 无认证
	rawURL   string // 展示用（含认证时隐藏密码）
}

// parseProxyURL 解析下游代理地址（默认中文错误）
func parseProxyURL(raw string) (*proxyConfig, error) {
	return parseProxyURLLang(raw, langZH)
}

// parseProxyURLLang 解析下游代理地址：缺 scheme 按 http 处理，仅支持 HTTP 代理
func parseProxyURLLang(raw string, l lang) (*proxyConfig, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, newLangError(l,
			"已启用代理但代理地址为空，请填写代理地址或关闭代理开关",
			"proxy is enabled but the proxy URL is empty; set a URL or turn the proxy off")
	}
	if !strings.Contains(raw, "://") {
		raw = "http://" + raw
	}
	u, err := url.Parse(raw)
	if err != nil || u.Hostname() == "" {
		return nil, newLangErrorf(l, "代理地址格式不正确: %q", "invalid proxy URL: %q", raw)
	}
	if scheme := strings.ToLower(u.Scheme); scheme != "http" {
		return nil, newLangErrorf(l, "暂不支持的代理协议 %q（当前仅支持 http 代理）", "unsupported proxy scheme %q (only HTTP proxies are supported)", u.Scheme)
	}
	port := 80
	if ps := u.Port(); ps != "" {
		p, err := strconv.Atoi(ps)
		if err != nil || p <= 0 {
			return nil, newLangErrorf(l, "代理端口不正确: %q", "invalid proxy port: %q", ps)
		}
		port = p
	}
	authLine := ""
	display := u.Scheme + "://" + u.Host
	if u.User != nil {
		pass, _ := u.User.Password()
		cred := u.User.Username() + ":" + pass
		authLine = "Proxy-Authorization: Basic " + base64.StdEncoding.EncodeToString([]byte(cred))
	}
	return &proxyConfig{
		addr:     net.JoinHostPort(u.Hostname(), strconv.Itoa(port)),
		authLine: authLine,
		rawURL:   display,
	}, nil
}

// dialViaProxy 连接下游代理；HTTPS 目标再发 CONNECT 建立隧道，返回可直用的原始连接
func dialViaProxy(ctx context.Context, cfg *proxyConfig, scheme, host string, port int, timeout time.Duration, l lang) (net.Conn, error) {
	dialer := &net.Dialer{Timeout: timeout}
	conn, err := dialer.DialContext(ctx, "tcp", cfg.addr)
	if err != nil {
		return nil, newLangErrorf(l, "连接代理失败（%s）: %w", "failed to connect to proxy (%s): %w", cfg.rawURL, err)
	}
	_ = conn.SetDeadline(time.Now().Add(timeout))

	// CONNECT 交换期间也要能被取消
	done := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			_ = conn.Close()
		case <-done:
		}
	}()
	defer close(done)

	if scheme != "https" {
		return conn, nil // HTTP 目标直接绝对 URL 请求式转发，无需握手
	}

	target := net.JoinHostPort(host, strconv.Itoa(port))
	var msg strings.Builder
	msg.WriteString("CONNECT " + target + " HTTP/1.1\r\n")
	msg.WriteString("Host: " + target + "\r\n")
	if cfg.authLine != "" {
		msg.WriteString(cfg.authLine + "\r\n")
	}
	msg.WriteString("\r\n")
	if _, err := conn.Write([]byte(msg.String())); err != nil {
		conn.Close()
		return nil, newLangErrorf(l, "向代理发送 CONNECT 失败: %w", "failed to send CONNECT to proxy: %w", err)
	}

	// CONNECT 的 2xx 响应无 body（http.ReadResponse 对 CONNECT+2xx 返回空 body）
	resp, err := http.ReadResponse(bufio.NewReader(conn), &http.Request{Method: http.MethodConnect})
	if err != nil {
		conn.Close()
		return nil, newLangErrorf(l, "代理 CONNECT 响应无法解析: %w", "cannot parse proxy CONNECT response: %w", err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		status := resp.Status
		conn.Close()
		return nil, newLangErrorf(l,
			"代理拒绝建立隧道（CONNECT %s），请检查代理地址/端口/认证",
			"proxy refused the CONNECT tunnel (%s); check the proxy address, port and credentials",
			status)
	}
	return conn, nil
}
