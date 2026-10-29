package repeater

// Service Wails v2 绑定入口（前端经 wailsjs/go/repeater/Service 调用）
type Service struct{}

func NewService() *Service {
	return &Service{}
}

// ParseRaw 解析原始报文，返回目标推断结果（表单联动）；lang 控制错误语言
func (s *Service) ParseRaw(raw string, lang string) (*ParsedRequest, error) {
	return ParseRawLang(raw, lang)
}

// Send 发送请求；网络/协议错误体现在 Response.Error 中，便于失败也生成历史
func (s *Service) Send(req SendRequest) (*SendResult, error) {
	return Send(req), nil
}

// ImportRequest 把 curl 命令 / 浏览器 fetch 代码转换为可编辑报文；识别失败返回错误
func (s *Service) ImportRequest(text string, lang string) (*ImportedRequest, error) {
	return ImportRequestLang(text, lang)
}
