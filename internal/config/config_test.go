package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig_DefaultsOnly(t *testing.T) {
	// 创建临时目录模拟 home
	tmpDir := t.TempDir()

	// 临时设置环境变量让 GetHomeDir 返回我们的目录
	origUserProfile := os.Getenv("USERPROFILE")
	origHome := os.Getenv("HOME")
	// 我们无法直接覆盖 GetHomeDir，所以测试 LoadConfig 的环境变量逻辑
	// 清除可能干扰的环境变量
	os.Unsetenv("UUPT_APP_ID")
	os.Unsetenv("UUPT_APP_SECRET")
	os.Unsetenv("UUPT_OPEN_ID")
	os.Unsetenv("UUPT_API_URL")
	defer func() {
		os.Setenv("USERPROFILE", origUserProfile)
		os.Setenv("HOME", origHome)
	}()

	// 直接测试 loadJSONFile 功能
	configDir := filepath.Join(tmpDir, "configs")
	os.MkdirAll(configDir, 0755)

	defaults := Config{
		AppId:     "default-app-id",
		AppSecret: "default-secret",
		ApiUrl:    "https://api.test.com",
	}
	data, _ := json.Marshal(defaults)
	os.WriteFile(filepath.Join(configDir, "defaults.json"), data, 0644)

	// 测试 loadJSONFile
	loaded := &Config{}
	err := loadJSONFile(filepath.Join(configDir, "defaults.json"), loaded)
	if err != nil {
		t.Fatalf("加载defaults.json失败: %v", err)
	}
	if loaded.AppId != "default-app-id" {
		t.Errorf("AppId不匹配: %s", loaded.AppId)
	}
	if loaded.AppSecret != "default-secret" {
		t.Errorf("AppSecret不匹配: %s", loaded.AppSecret)
	}
	if loaded.ApiUrl != "https://api.test.com" {
		t.Errorf("ApiUrl不匹配: %s", loaded.ApiUrl)
	}
}

func TestLoadConfig_EnvOverride(t *testing.T) {
	os.Setenv("UUPT_APP_ID", "test-app-id")
	os.Setenv("UUPT_APP_SECRET", "test-secret")
	os.Setenv("UUPT_OPEN_ID", "test-open-id")
	defer func() {
		os.Unsetenv("UUPT_APP_ID")
		os.Unsetenv("UUPT_APP_SECRET")
		os.Unsetenv("UUPT_OPEN_ID")
	}()

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig失败: %v", err)
	}
	if cfg.AppId != "test-app-id" {
		t.Errorf("环境变量AppId覆盖失败: %s", cfg.AppId)
	}
	if cfg.AppSecret != "test-secret" {
		t.Errorf("环境变量AppSecret覆盖失败: %s", cfg.AppSecret)
	}
	if cfg.OpenId != "test-open-id" {
		t.Errorf("环境变量OpenId覆盖失败: %s", cfg.OpenId)
	}
}

func TestSaveConfig_MergeWrite(t *testing.T) {
	// 保存原始环境
	origUserProfile := os.Getenv("USERPROFILE")
	origHome := os.Getenv("HOME")

	// 设置临时目录为 home
	tmpDir := t.TempDir()
	os.Setenv("USERPROFILE", tmpDir)
	os.Setenv("HOME", tmpDir)
	defer func() {
		os.Setenv("USERPROFILE", origUserProfile)
		os.Setenv("HOME", origHome)
	}()

	// 第一次写入
	err := SaveConfig(map[string]string{"appId": "id1"})
	if err != nil {
		t.Fatalf("第一次SaveConfig失败: %v", err)
	}

	// 第二次写入另一个key
	err = SaveConfig(map[string]string{"appSecret": "secret1"})
	if err != nil {
		t.Fatalf("第二次SaveConfig失败: %v", err)
	}

	// 验证两个都存在
	configPath := filepath.Join(GetHomeDir(), "configs", "config.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("读取config.json失败: %v", err)
	}

	var result map[string]string
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("解析config.json失败: %v", err)
	}

	if result["appId"] != "id1" {
		t.Errorf("appId丢失或不正确: %s", result["appId"])
	}
	if result["appSecret"] != "secret1" {
		t.Errorf("appSecret丢失或不正确: %s", result["appSecret"])
	}

	if err := SaveConfig(map[string]string{"openId": "oid-1"}); err != nil {
		t.Fatalf("写入openId失败: %v", err)
	}
	if err := ClearOpenId(); err != nil {
		t.Fatalf("ClearOpenId失败: %v", err)
	}
	data, err = os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("再次读取config.json失败: %v", err)
	}
	result = map[string]string{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("再次解析config.json失败: %v", err)
	}
	if _, ok := result["openId"]; ok {
		t.Errorf("ClearOpenId 后仍残留 openId: %v", result)
	}
	if result["appId"] != "id1" {
		t.Errorf("ClearOpenId 不应删除其他字段")
	}
	if err := ClearOpenId(); err != nil {
		t.Fatalf("重复 ClearOpenId 应幂等: %v", err)
	}
}

func TestLoadConfig_FileNotExist(t *testing.T) {
	// 清除环境变量
	os.Unsetenv("UUPT_APP_ID")
	os.Unsetenv("UUPT_APP_SECRET")
	os.Unsetenv("UUPT_OPEN_ID")
	os.Unsetenv("UUPT_API_URL")

	// LoadConfig 即使文件不存在也应返回空 Config 而非 error
	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("文件不存在时LoadConfig不应报错: %v", err)
	}
	if cfg == nil {
		t.Fatal("cfg不应为nil")
	}
}

func TestGetHomeDir(t *testing.T) {
	dir := GetHomeDir()
	if dir == "" {
		t.Error("GetHomeDir不应返回空字符串")
	}
	if !filepath.IsAbs(dir) {
		t.Errorf("GetHomeDir应返回绝对路径: %s", dir)
	}
}
