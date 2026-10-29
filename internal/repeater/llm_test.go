package repeater

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestResolveLLMBaseURL(t *testing.T) {
	cases := []struct {
		base string
		want string
	}{
		{"https://api.openai.com", "https://api.openai.com/v1/"},
		{"https://api.openai.com/v1", "https://api.openai.com/v1/"},
		{"https://api.openai.com/v1/", "https://api.openai.com/v1/"},
		{"https://openrouter.ai/api/v1", "https://openrouter.ai/api/v1/"},
		{"https://host/custom/openai", "https://host/custom/openai/"},
		{"https://host/v1/chat/completions", "https://host/v1/"},
		{"https://host/v1/responses", "https://host/v1/"},
		{"https://host/openai/deployments/x", "https://host/openai/deployments/x/"},
		{"127.0.0.1:8080/v1", "http://127.0.0.1:8080/v1/"},
		{"localhost:8080", "http://localhost:8080/v1/"},
	}
	for _, tc := range cases {
		got, err := resolveLLMBaseURL(tc.base)
		if err != nil {
			t.Errorf("resolve(%q): %v", tc.base, err)
			continue
		}
		if got != tc.want {
			t.Errorf("resolve(%q) = %q, want %q", tc.base, got, tc.want)
		}
	}
}

func TestChatLLMCompletions(t *testing.T) {
	var gotPath, gotAuth string
	var payload map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &payload)
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"choices":[{"message":{"content":"hello"}}]}`)
	}))
	defer srv.Close()

	cfg := LLMSettings{
		BaseURL:         srv.URL + "/v1",
		APIKey:          "sk-test",
		Model:           "test-model",
		APIType:         "completions",
		ReasoningEffort: "medium",
		SystemPrompt:    "custom system",
	}
	req := ChatRequest{
		Messages:    []ChatMessage{{Role: "user", Content: "hi"}},
		RequestText: "GET /x HTTP/1.1\nHost: example.com\n\n",
		Scheme:      "https",
		Host:        "example.com",
		Lang:        "en",
	}
	res, err := chatLLM(context.Background(), cfg, req, srv.Client())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Content != "hello" {
		t.Errorf("content = %q", res.Content)
	}
	if gotPath != "/v1/chat/completions" {
		t.Errorf("path = %q", gotPath)
	}
	if gotAuth != "Bearer sk-test" {
		t.Errorf("authorization = %q", gotAuth)
	}
	if payload["model"] != "test-model" {
		t.Errorf("model = %v", payload["model"])
	}
	if payload["reasoning_effort"] != "medium" {
		t.Errorf("reasoning_effort = %v", payload["reasoning_effort"])
	}
	messages, ok := payload["messages"].([]any)
	if !ok || len(messages) != 2 {
		t.Fatalf("messages = %#v", payload["messages"])
	}
	system, _ := messages[0].(map[string]any)
	if content, _ := system["content"].(string); !strings.Contains(content, "custom system") ||
		!strings.Contains(content, "GET /x HTTP/1.1") {
		t.Errorf("system content = %q", content)
	}
	user, _ := messages[1].(map[string]any)
	if user["role"] != "user" || user["content"] != "hi" {
		t.Errorf("user message = %#v", user)
	}
}

func TestChatLLMResponses(t *testing.T) {
	var gotPath, gotAuth string
	var payload map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &payload)
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"output":[{"type":"message","content":[{"type":"output_text","text":"world"}]}]}`)
	}))
	defer srv.Close()

	cfg := LLMSettings{
		BaseURL:         srv.URL,
		APIKey:          "sk-test",
		Model:           "gpt-5",
		APIType:         "responses",
		ReasoningEffort: "low",
		SystemPrompt:    "sys",
	}
	req := ChatRequest{
		Messages: []ChatMessage{
			{Role: "user", Content: "first"},
			{Role: "assistant", Content: "answer"},
			{Role: "user", Content: "second"},
		},
		RequestText: "POST /y HTTP/1.1\nHost: a.test\n\n",
		Lang:        "en",
	}
	res, err := chatLLM(context.Background(), cfg, req, srv.Client())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Content != "world" {
		t.Errorf("content = %q", res.Content)
	}
	if gotPath != "/v1/responses" {
		t.Errorf("path = %q", gotPath)
	}
	if gotAuth != "Bearer sk-test" {
		t.Errorf("authorization = %q", gotAuth)
	}
	if payload["store"] != false {
		t.Errorf("store = %v", payload["store"])
	}
	instructions, _ := payload["instructions"].(string)
	if !strings.Contains(instructions, "sys") || !strings.Contains(instructions, "POST /y HTTP/1.1") {
		t.Errorf("instructions = %q", instructions)
	}
	input, ok := payload["input"].([]any)
	if !ok || len(input) != 3 {
		t.Fatalf("input = %#v", payload["input"])
	}
	last, _ := input[2].(map[string]any)
	if last["role"] != "user" || last["content"] != "second" {
		t.Errorf("last input = %#v", last)
	}
	reasoning, _ := payload["reasoning"].(map[string]any)
	if reasoning["effort"] != "low" {
		t.Errorf("reasoning = %#v", payload["reasoning"])
	}
}

func TestChatLLMHTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = io.WriteString(w, `{"error":{"message":"bad key"}}`)
	}))
	defer srv.Close()

	cfg := LLMSettings{BaseURL: srv.URL, Model: "m", APIType: "completions"}
	req := ChatRequest{Messages: []ChatMessage{{Role: "user", Content: "hi"}}, Lang: "en"}
	_, err := chatLLM(context.Background(), cfg, req, srv.Client())
	if err == nil {
		t.Fatal("expected an error")
	}
	if !strings.Contains(err.Error(), "bad key") || !strings.Contains(err.Error(), "401") {
		t.Errorf("error = %v", err)
	}
}

func TestChatLLMNotConfigured(t *testing.T) {
	_, err := chatLLM(context.Background(), LLMSettings{}, ChatRequest{
		Messages: []ChatMessage{{Role: "user", Content: "hi"}},
		Lang:     "en",
	}, nil)
	if err == nil || !strings.Contains(err.Error(), "not configured") {
		t.Fatalf("error = %v", err)
	}
}

func TestChatLLMCancel(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(started)
		<-release
	}))
	defer func() {
		close(release)
		srv.Close()
	}()

	cfg := LLMSettings{BaseURL: srv.URL, Model: "m", APIType: "completions"}
	req := ChatRequest{Messages: []ChatMessage{{Role: "user", Content: "hi"}}, Lang: "en"}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		_, err := chatLLM(ctx, cfg, req, srv.Client())
		done <- err
	}()

	select {
	case <-started:
	case <-time.After(3 * time.Second):
		t.Fatal("request did not start")
	}
	cancel()
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("expected a cancellation error")
		}
	case <-time.After(3 * time.Second):
		t.Fatal("request was not cancelled")
	}
}

func TestChatLLMNoAPIKey(t *testing.T) {
	var auth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"choices":[{"message":{"content":"local"}}]}`)
	}))
	defer srv.Close()

	cfg := LLMSettings{BaseURL: srv.URL, Model: "local-model", APIType: "completions"}
	req := ChatRequest{Messages: []ChatMessage{{Role: "user", Content: "hi"}}, Lang: "en"}
	res, err := chatLLM(context.Background(), cfg, req, srv.Client())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Content != "local" {
		t.Errorf("content = %q", res.Content)
	}
	if auth != "" {
		t.Errorf("authorization should be empty for a keyless endpoint, got %q", auth)
	}
}
