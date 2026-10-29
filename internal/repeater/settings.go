package repeater

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// DefaultProxyURL 新标签页的下游代理默认值
const DefaultProxyURL = "http://127.0.0.1:7890"

// Settings 持久化到 ~/.httpRepeater/setting.json
type Settings struct {
	Language        string      `json:"language"` // zh | en
	Theme           string      `json:"theme"`    // dark | light
	DefaultProxyURL string      `json:"defaultProxyUrl"`
	LLM             LLMSettings `json:"llm"`
}

// SettingsLoadResult 设置读取结果；文件缺失/损坏时返回默认值并带提示
type SettingsLoadResult struct {
	Settings            Settings `json:"settings"`
	Path                string   `json:"path"`
	Exists              bool     `json:"exists"`
	Error               string   `json:"error"` // 非致命：读取/解析失败说明
	DefaultSystemPrompt string   `json:"defaultSystemPrompt"`
}

var settingsMu sync.Mutex

// DefaultSettings 首次运行或字段缺失时使用的默认配置
func DefaultSettings() Settings {
	return Settings{
		Language:        "zh",
		Theme:           "dark",
		DefaultProxyURL: DefaultProxyURL,
		LLM: LLMSettings{
			BaseURL:         "https://api.openai.com/v1",
			Model:           "",
			APIType:         "completions",
			ReasoningEffort: "",
			SystemPrompt:    DefaultSystemPrompt,
		},
	}
}

// NormalizeSettings 补齐/纠正枚举值，避免脏配置影响请求
func NormalizeSettings(s *Settings) {
	if strings.EqualFold(strings.TrimSpace(s.Language), "en") {
		s.Language = "en"
	} else {
		s.Language = "zh"
	}
	if strings.EqualFold(strings.TrimSpace(s.Theme), "light") {
		s.Theme = "light"
	} else {
		s.Theme = "dark"
	}
	s.DefaultProxyURL = strings.TrimSpace(s.DefaultProxyURL)
	s.LLM.BaseURL = strings.TrimSpace(s.LLM.BaseURL)
	s.LLM.APIKey = strings.TrimSpace(s.LLM.APIKey)
	s.LLM.Model = strings.TrimSpace(s.LLM.Model)
	s.LLM.APIType = normalizeAPIType(s.LLM.APIType)
	s.LLM.ReasoningEffort = normalizeReasoningEffort(s.LLM.ReasoningEffort)
	s.LLM.SystemPrompt = strings.TrimSpace(s.LLM.SystemPrompt)
	if s.LLM.SystemPrompt == "" {
		s.LLM.SystemPrompt = DefaultSystemPrompt
	}
}

// SettingsPath 返回配置文件路径（~/.httpRepeater/setting.json）
func SettingsPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot resolve the user home directory: %w", err)
	}
	return filepath.Join(home, ".httpRepeater", "setting.json"), nil
}

// LoadSettings 读取设置；文件不存在时返回默认值，不视为错误
func LoadSettings() (*SettingsLoadResult, error) {
	path, err := SettingsPath()
	if err != nil {
		return nil, err
	}
	settingsMu.Lock()
	defer settingsMu.Unlock()
	return loadSettingsAt(path), nil
}

func loadSettingsAt(path string) *SettingsLoadResult {
	res := &SettingsLoadResult{
		Settings:            DefaultSettings(),
		Path:                path,
		DefaultSystemPrompt: DefaultSystemPrompt,
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			res.Error = err.Error()
		}
		return res
	}
	res.Exists = true

	settings := DefaultSettings()
	if err := json.Unmarshal(data, &settings); err != nil {
		res.Error = err.Error()
		return res
	}
	NormalizeSettings(&settings)
	res.Settings = settings
	return res
}

// SaveSettings 原子写入配置文件（权限 0600，含 API Key）
func SaveSettings(settings Settings) error {
	path, err := SettingsPath()
	if err != nil {
		return err
	}
	NormalizeSettings(&settings)
	settingsMu.Lock()
	defer settingsMu.Unlock()
	return saveSettingsAt(path, settings)
}

func saveSettingsAt(path string, settings Settings) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("cannot create the settings directory: %w", err)
	}
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("cannot encode settings: %w", err)
	}
	data = append(data, '\n')

	tmp, err := os.CreateTemp(dir, ".setting-*.json")
	if err != nil {
		return fmt.Errorf("cannot create a temporary settings file: %w", err)
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	_ = tmp.Chmod(0o600)

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return fmt.Errorf("cannot write settings: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return fmt.Errorf("cannot flush settings: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("cannot close the settings file: %w", err)
	}
	if err := os.Rename(tmpName, path); err != nil {
		return fmt.Errorf("cannot replace the settings file: %w", err)
	}
	return nil
}
