package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

type Config struct {
	AppId     string `json:"appId"`
	AppSecret string `json:"appSecret"`
	OpenId    string `json:"openId"`
	ApiUrl    string `json:"apiUrl"`
}

func GetHomeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		// 降级：手动获取
		if runtime.GOOS == "windows" {
			home = os.Getenv("USERPROFILE")
		} else {
			home = os.Getenv("HOME")
		}
	}
	return filepath.Join(home, ".uupt-open-cli")
}

func LoadConfig() (*Config, error) {
	cfg := &Config{}
	homeDir := GetHomeDir()

	// Layer 1: defaults.json
	defaultsPath := filepath.Join(homeDir, "configs", "defaults.json")
	loadJSONFile(defaultsPath, cfg)

	// Layer 2: config.json (override non-empty fields)
	configPath := filepath.Join(homeDir, "configs", "config.json")
	overrideCfg := &Config{}
	if loadJSONFile(configPath, overrideCfg) == nil {
		if overrideCfg.AppId != "" {
			cfg.AppId = overrideCfg.AppId
		}
		if overrideCfg.AppSecret != "" {
			cfg.AppSecret = overrideCfg.AppSecret
		}
		if overrideCfg.OpenId != "" {
			cfg.OpenId = overrideCfg.OpenId
		}
		if overrideCfg.ApiUrl != "" {
			cfg.ApiUrl = overrideCfg.ApiUrl
		}
	}

	// Layer 3: environment variables
	if v := os.Getenv("UUPT_APP_ID"); v != "" {
		cfg.AppId = v
	}
	if v := os.Getenv("UUPT_APP_SECRET"); v != "" {
		cfg.AppSecret = v
	}
	if v := os.Getenv("UUPT_OPEN_ID"); v != "" {
		cfg.OpenId = v
	}
	if v := os.Getenv("UUPT_API_URL"); v != "" {
		cfg.ApiUrl = v
	}

	return cfg, nil
}

func loadJSONFile(path string, target interface{}) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	// Strip UTF-8 BOM if present
	data = bytes.TrimPrefix(data, []byte("\xef\xbb\xbf"))
	return json.Unmarshal(data, target)
}

func SaveConfig(updates map[string]string) error {
	homeDir := GetHomeDir()
	configDir := filepath.Join(homeDir, "configs")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	configPath := filepath.Join(configDir, "config.json")

	existing := make(map[string]string)
	data, err := os.ReadFile(configPath)
	if err == nil {
		json.Unmarshal(data, &existing)
	}

	for k, v := range updates {
		existing[k] = v
	}

	out, err := json.MarshalIndent(existing, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, out, 0644)
}

func EnsureConfig(needOpenId bool) *Config {
	cfg, _ := LoadConfig()

	if cfg.AppId == "" || cfg.AppSecret == "" {
		fmt.Println("[FATAL] 缺少必要配置，请设置 AppId 和 AppSecret")
		os.Exit(1)
	}

	if needOpenId && cfg.OpenId == "" {
		fmt.Println("[REGISTRATION_REQUIRED]\n请先注册获取 OpenId")
		os.Exit(1)
	}

	return cfg
}
