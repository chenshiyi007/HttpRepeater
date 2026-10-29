package repeater

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/responses"
	"github.com/openai/openai-go/v3/shared"
)

// DefaultSystemPrompt 指导模型产出可直接发送的原始 HTTP 报文
const DefaultSystemPrompt = `You are an HTTP request assistant embedded in HttpRepeater, a desktop HTTP replay tool.
Your job is to help the user edit raw HTTP/1.1 requests and craft payloads.

Rules:
- Reply in the same language the user uses.
- When you produce or modify a request, output the complete raw HTTP/1.1 request in a single ` + "```http" + ` fenced block: request line, headers, one blank line, then the body.
- The request must be directly sendable: keep or add Host, Content-Type and Content-Length when relevant.
- Keep the existing target, headers and formatting unless the user asks to change them.
- Do not send the request; the user applies your output manually.
- If the user only asks for a payload or an explanation, answer normally and put code/data in fenced blocks.`

const defaultLLMTimeout = 180 * time.Second

// LLMSettings 兼容 OpenAI Responses API 与 Chat Completions API
type LLMSettings struct {
	BaseURL         string `json:"baseUrl"`
	APIKey          string `json:"apiKey"`
	Model           string `json:"model"`
	APIType         string `json:"apiType"`         // responses | completions
	ReasoningEffort string `json:"reasoningEffort"` // "", minimal, low, medium, high
	SystemPrompt    string `json:"systemPrompt"`
}

// ChatMessage 对话消息（由前端维护多轮历史）
type ChatMessage struct {
	Role    string `json:"role"` // user | assistant
	Content string `json:"content"`
}

// ChatRequest 一次对话请求；当前编辑器内容随每条消息一起提供给模型
type ChatRequest struct {
	Messages    []ChatMessage `json:"messages"`
	RequestText string        `json:"requestText"`
	Scheme      string        `json:"scheme"`
	Host        string        `json:"host"`
	Port        int           `json:"port"`
	Lang        string        `json:"lang"`
}

// ChatResponse 模型回复
type ChatResponse struct {
	Content string `json:"content"`
}

// chatLLM 使用官方 OpenAI Go SDK 调用模型；client 仅用于测试注入
func chatLLM(ctx context.Context, cfg LLMSettings, req ChatRequest, httpClient *http.Client) (*ChatResponse, error) {
	l := normalizeLang(req.Lang)
	if strings.TrimSpace(cfg.BaseURL) == "" || strings.TrimSpace(cfg.Model) == "" {
		return nil, newLangError(l,
			"LLM 未配置：请在设置中填写 Base URL 和模型",
			"LLM is not configured: set the base URL and model in Settings")
	}
	if len(req.Messages) == 0 {
		return nil, newLangError(l, "对话消息为空", "the conversation is empty")
	}

	baseURL, err := resolveLLMBaseURL(cfg.BaseURL)
	if err != nil {
		return nil, newLangErrorf(l, "LLM Base URL 不正确: %v", "invalid LLM base URL: %v", err)
	}

	opts := []option.RequestOption{
		option.WithAPIKey(strings.TrimSpace(cfg.APIKey)),
		option.WithBaseURL(baseURL),
		option.WithMaxRetries(0), // 取消/失败都直接返回，避免静默重试
	}
	if httpClient == nil {
		httpClient = &http.Client{Timeout: defaultLLMTimeout}
	}
	opts = append(opts, option.WithHTTPClient(httpClient))
	client := openai.NewClient(opts...)

	instructions := buildLLMInstructions(cfg, req)
	apiType := normalizeAPIType(cfg.APIType)
	effort := normalizeReasoningEffort(cfg.ReasoningEffort)

	if apiType == "responses" {
		return chatResponses(ctx, client, l, cfg.Model, instructions, effort, req)
	}
	return chatCompletions(ctx, client, l, cfg.Model, instructions, effort, req)
}

func chatCompletions(
	ctx context.Context,
	client openai.Client,
	l lang,
	model, instructions, effort string,
	req ChatRequest,
) (*ChatResponse, error) {
	messages := make([]openai.ChatCompletionMessageParamUnion, 0, len(req.Messages)+1)
	messages = append(messages, openai.SystemMessage(instructions))
	appended := 0
	for _, m := range req.Messages {
		if strings.TrimSpace(m.Content) == "" {
			continue
		}
		if strings.EqualFold(strings.TrimSpace(m.Role), "assistant") {
			messages = append(messages, openai.AssistantMessage(m.Content))
		} else {
			messages = append(messages, openai.UserMessage(m.Content))
		}
		appended++
	}
	if appended == 0 {
		return nil, newLangError(l, "对话消息为空", "the conversation is empty")
	}

	params := openai.ChatCompletionNewParams{
		Model:    shared.ChatModel(model),
		Messages: messages,
	}
	if effort != "" {
		params.ReasoningEffort = shared.ReasoningEffort(effort)
	}

	completion, err := client.Chat.Completions.New(ctx, params)
	if err != nil {
		return nil, llmRequestError(l, err)
	}
	for _, choice := range completion.Choices {
		if content := strings.TrimSpace(choice.Message.Content); content != "" {
			return &ChatResponse{Content: content}, nil
		}
		if refusal := strings.TrimSpace(choice.Message.Refusal); refusal != "" {
			return &ChatResponse{Content: refusal}, nil
		}
	}
	return nil, newLangError(l, "LLM 返回了空内容", "the LLM returned an empty response")
}

func chatResponses(
	ctx context.Context,
	client openai.Client,
	l lang,
	model, instructions, effort string,
	req ChatRequest,
) (*ChatResponse, error) {
	input := make(responses.ResponseInputParam, 0, len(req.Messages))
	for _, m := range req.Messages {
		if strings.TrimSpace(m.Content) == "" {
			continue
		}
		role := responses.EasyInputMessageRoleUser
		if strings.EqualFold(strings.TrimSpace(m.Role), "assistant") {
			role = responses.EasyInputMessageRoleAssistant
		}
		input = append(input, responses.ResponseInputItemParamOfMessage(m.Content, role))
	}
	if len(input) == 0 {
		return nil, newLangError(l, "对话消息为空", "the conversation is empty")
	}

	params := responses.ResponseNewParams{
		Model:        shared.ResponsesModel(model),
		Instructions: openai.String(instructions),
		Input:        responses.ResponseNewParamsInputUnion{OfInputItemList: input},
		Store:        openai.Bool(false),
	}
	if effort != "" {
		params.Reasoning = shared.ReasoningParam{Effort: shared.ReasoningEffort(effort)}
	}

	response, err := client.Responses.New(ctx, params)
	if err != nil {
		return nil, llmRequestError(l, err)
	}
	if content := strings.TrimSpace(response.OutputText()); content != "" {
		return &ChatResponse{Content: content}, nil
	}
	return nil, newLangError(l, "LLM 返回了空内容", "the LLM returned an empty response")
}

func llmRequestError(l lang, err error) error {
	var apiErr *openai.Error
	if errors.As(err, &apiErr) {
		message := strings.TrimSpace(apiErr.Message)
		if message == "" {
			message = apiErr.Error()
		}
		return newLangErrorf(l,
			"LLM 请求失败（HTTP %d）：%s",
			"LLM request failed (HTTP %d): %s",
			apiErr.StatusCode, message)
	}
	return newLangErrorf(l, "LLM 请求失败: %v", "LLM request failed: %v", err)
}

func normalizeAPIType(value string) string {
	if strings.EqualFold(strings.TrimSpace(value), "responses") {
		return "responses"
	}
	return "completions"
}

func normalizeReasoningEffort(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "minimal", "low", "medium", "high":
		return strings.ToLower(strings.TrimSpace(value))
	default:
		return ""
	}
}

// resolveLLMBaseURL 把用户填写的根地址 / /v1 地址 / 完整 endpoint 归一成 SDK 需要的 base URL。
// 例如 https://api.openai.com、https://api.openai.com/v1、.../v1/chat/completions
// 都会得到 https://api.openai.com/v1/。
func resolveLLMBaseURL(baseURL string) (string, error) {
	raw := strings.TrimSpace(baseURL)
	if raw == "" {
		return "", errors.New("base URL is empty")
	}
	if !strings.Contains(raw, "://") {
		lower := strings.ToLower(raw)
		if strings.HasPrefix(lower, "localhost") || strings.HasPrefix(lower, "127.") ||
			strings.HasPrefix(lower, "[::1]") || strings.HasPrefix(lower, "0.0.0.0") {
			raw = "http://" + raw
		} else {
			raw = "https://" + raw
		}
	}
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" {
		return "", fmt.Errorf("cannot parse %q", baseURL)
	}

	path := strings.TrimRight(u.Path, "/")
	lowerPath := strings.ToLower(path)
	for _, suffix := range []string{"/chat/completions", "/responses"} {
		if strings.HasSuffix(lowerPath, suffix) {
			path = path[:len(path)-len(suffix)]
			break
		}
	}
	if path == "" {
		path = "/v1"
	}
	u.Path = path + "/"
	u.RawQuery = ""
	return u.String(), nil
}

func buildLLMInstructions(cfg LLMSettings, req ChatRequest) string {
	systemPrompt := strings.TrimSpace(cfg.SystemPrompt)
	if systemPrompt == "" {
		systemPrompt = DefaultSystemPrompt
	}

	scheme := strings.TrimSpace(req.Scheme)
	host := strings.TrimSpace(req.Host)
	port := req.Port
	if host == "" {
		if p, err := parseRequestLang(req.RequestText, normalizeLang(req.Lang)); err == nil {
			scheme, host, port = p.Target.Scheme, p.Target.Host, p.Target.Port
		}
	}

	var b strings.Builder
	b.WriteString(systemPrompt)
	b.WriteString("\n\n")
	if host != "" {
		target := strings.TrimSuffix(composeURL(scheme, host, port, "/"), "/")
		b.WriteString("Current target: " + target + "\n")
	}
	b.WriteString("Current request in the editor:\n")
	raw := strings.TrimSpace(req.RequestText)
	if raw == "" {
		b.WriteString("(empty)\n")
	} else {
		b.WriteString("```http\n")
		b.WriteString(raw)
		b.WriteString("\n```\n")
	}
	return b.String()
}
