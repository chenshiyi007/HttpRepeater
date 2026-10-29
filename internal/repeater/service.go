package repeater

import (
	"context"
	"errors"
	"sync"
)

// Service Wails v2 绑定入口（前端经 wailsjs/go/repeater/Service 调用）
type Service struct {
	chatMu     sync.Mutex
	chatCancel context.CancelFunc
	chatSeq    uint64

	sendMu     sync.Mutex
	sendCancel context.CancelFunc
	sendSeq    uint64
}

func NewService() *Service {
	return &Service{}
}

// ParseRaw 解析原始报文，返回目标推断结果（表单联动）；lang 控制错误语言
func (s *Service) ParseRaw(raw string, lang string) (*ParsedRequest, error) {
	return ParseRawLang(raw, lang)
}

// Send 发送请求；网络/协议错误体现在 Response.Error 中，便于失败也生成历史
func (s *Service) Send(req SendRequest) (*SendResult, error) {
	ctx, cancel := context.WithCancel(context.Background())
	s.sendMu.Lock()
	if s.sendCancel != nil {
		s.sendCancel()
	}
	s.sendSeq++
	seq := s.sendSeq
	s.sendCancel = cancel
	s.sendMu.Unlock()

	defer func() {
		cancel()
		s.sendMu.Lock()
		if s.sendSeq == seq {
			s.sendCancel = nil
		}
		s.sendMu.Unlock()
	}()

	return sendWithContext(ctx, req), nil
}

// CancelSend 取消正在发送的请求（关闭连接，结果里会标记「已取消」）
func (s *Service) CancelSend() {
	s.sendMu.Lock()
	cancel := s.sendCancel
	s.sendCancel = nil
	s.sendSeq++
	s.sendMu.Unlock()
	if cancel != nil {
		cancel()
	}
}

// ImportRequest 把 curl 命令 / 浏览器 fetch 代码转换为可编辑报文；识别失败返回错误
func (s *Service) ImportRequest(text string, lang string) (*ImportedRequest, error) {
	return ImportRequestLang(text, lang)
}

// LoadSettings 读取 ~/.httpRepeater/setting.json；缺失时返回默认值
func (s *Service) LoadSettings() (*SettingsLoadResult, error) {
	return LoadSettings()
}

// SaveSettings 保存设置到 ~/.httpRepeater/setting.json
func (s *Service) SaveSettings(settings Settings) error {
	return SaveSettings(settings)
}

// LLMChat 与 LLM 多轮对话；设置每次从配置文件读取。可被 CancelLLMChat 中断。
func (s *Service) LLMChat(req ChatRequest) (*ChatResponse, error) {
	res, err := LoadSettings()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())
	s.chatMu.Lock()
	if s.chatCancel != nil {
		s.chatCancel() // 同一时间只保留一个请求
	}
	s.chatSeq++
	seq := s.chatSeq
	s.chatCancel = cancel
	s.chatMu.Unlock()

	defer func() {
		cancel()
		s.chatMu.Lock()
		if s.chatSeq == seq {
			s.chatCancel = nil
		}
		s.chatMu.Unlock()
	}()

	out, err := chatLLM(ctx, res.Settings.LLM, req, nil)
	if err != nil && errors.Is(ctx.Err(), context.Canceled) {
		return nil, newLangError(normalizeLang(req.Lang), "已取消", "cancelled")
	}
	return out, err
}

// CancelLLMChat 取消正在进行的 LLM 请求（用户点「停止」时调用）
func (s *Service) CancelLLMChat() {
	s.chatMu.Lock()
	cancel := s.chatCancel
	s.chatCancel = nil
	s.chatSeq++ // 让进行中请求的清理逻辑不再覆盖新的取消函数
	s.chatMu.Unlock()
	if cancel != nil {
		cancel()
	}
}

// BuildRequestURL 生成「复制 URL」文本（协议 + Host + 路径/查询参数）
func (s *Service) BuildRequestURL(req RequestToolInput) (*RequestToolResult, error) {
	return BuildRequestURL(req)
}

// BuildRequestCurl 生成「复制为 curl」命令
func (s *Service) BuildRequestCurl(req RequestToolInput) (*RequestToolResult, error) {
	return BuildRequestCurl(req)
}

// ConvertRequestMethod 仿 Burp 的 GET↔POST 请求方法转换
func (s *Service) ConvertRequestMethod(req RequestToolInput) (*RequestToolResult, error) {
	return ConvertRequestMethod(req)
}
