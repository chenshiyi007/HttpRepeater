package repeater

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestLoadSettingsMissingFileReturnsDefaults(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".httpRepeater", "setting.json")
	res := loadSettingsAt(path)
	if res.Exists {
		t.Error("Exists should be false")
	}
	if res.Error != "" {
		t.Errorf("unexpected error: %q", res.Error)
	}
	if res.Settings.Language != "zh" || res.Settings.Theme != "dark" {
		t.Errorf("defaults = %+v", res.Settings)
	}
	if res.Settings.DefaultProxyURL != DefaultProxyURL {
		t.Errorf("proxy = %q", res.Settings.DefaultProxyURL)
	}
	if res.Settings.LLM.APIType != "completions" {
		t.Errorf("api type = %q", res.Settings.LLM.APIType)
	}
	if res.DefaultSystemPrompt == "" || res.Settings.LLM.SystemPrompt != DefaultSystemPrompt {
		t.Error("default system prompt missing")
	}
}

func TestSaveAndLoadSettings(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".httpRepeater", "setting.json")
	in := Settings{
		Language:        "en",
		Theme:           "light",
		DefaultProxyURL: "http://127.0.0.1:8888",
		LLM: LLMSettings{
			BaseURL:         "https://api.example.com/v1",
			APIKey:          "sk-test",
			Model:           "test-model",
			APIType:         "responses",
			ReasoningEffort: "high",
			SystemPrompt:    "custom",
		},
	}
	if err := saveSettingsAt(path, in); err != nil {
		t.Fatalf("save failed: %v", err)
	}

	res := loadSettingsAt(path)
	if !res.Exists || res.Error != "" {
		t.Fatalf("load result = %+v", res)
	}
	got := res.Settings
	if got.Language != "en" || got.Theme != "light" || got.DefaultProxyURL != in.DefaultProxyURL {
		t.Errorf("settings = %+v", got)
	}
	if got.LLM.BaseURL != in.LLM.BaseURL || got.LLM.APIKey != "sk-test" ||
		got.LLM.Model != "test-model" || got.LLM.APIType != "responses" ||
		got.LLM.ReasoningEffort != "high" || got.LLM.SystemPrompt != "custom" {
		t.Errorf("llm = %+v", got.LLM)
	}

	if runtime.GOOS != "windows" {
		info, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		if info.Mode().Perm() != 0o600 {
			t.Errorf("mode = %o, want 600", info.Mode().Perm())
		}
	}
}

func TestLoadSettingsMergesDefaults(t *testing.T) {
	path := filepath.Join(t.TempDir(), "setting.json")
	if err := os.WriteFile(path, []byte(`{"language":"en","llm":{"model":"m"}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	res := loadSettingsAt(path)
	if res.Error != "" {
		t.Fatalf("unexpected error: %q", res.Error)
	}
	if res.Settings.Theme != "dark" || res.Settings.DefaultProxyURL != DefaultProxyURL {
		t.Errorf("defaults not merged: %+v", res.Settings)
	}
	if res.Settings.LLM.Model != "m" || res.Settings.LLM.APIType != "completions" {
		t.Errorf("llm merge = %+v", res.Settings.LLM)
	}
	if res.Settings.LLM.SystemPrompt != DefaultSystemPrompt {
		t.Error("system prompt should fall back to default")
	}
}

func TestLoadSettingsCorruptFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "setting.json")
	if err := os.WriteFile(path, []byte("{not json"), 0o600); err != nil {
		t.Fatal(err)
	}
	res := loadSettingsAt(path)
	if res.Error == "" {
		t.Fatal("expected a non-fatal parse error")
	}
	if res.Settings.Language != "zh" {
		t.Errorf("should fall back to defaults, got %+v", res.Settings)
	}
}

func TestNormalizeSettings(t *testing.T) {
	s := Settings{
		Language: "EN",
		Theme:    "LIGHT",
		LLM: LLMSettings{
			APIType:         "unknown",
			ReasoningEffort: "HIGH",
			SystemPrompt:    "  ",
		},
	}
	NormalizeSettings(&s)
	if s.Language != "en" || s.Theme != "light" {
		t.Errorf("language/theme = %q/%q", s.Language, s.Theme)
	}
	if s.LLM.APIType != "completions" || s.LLM.ReasoningEffort != "high" {
		t.Errorf("llm = %+v", s.LLM)
	}
	if !strings.Contains(s.LLM.SystemPrompt, "HTTP request assistant") {
		t.Errorf("system prompt = %q", s.LLM.SystemPrompt)
	}
}
