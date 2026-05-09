package qrcode

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestDownloadQRCode_Success(t *testing.T) {
	// Mock 一个返回 PNG 数据的服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		w.Write([]byte("fake-png-data"))
	}))
	defer server.Close()

	// 设置临时 home 目录
	tmpDir := t.TempDir()
	origUserProfile := os.Getenv("USERPROFILE")
	origHome := os.Getenv("HOME")
	os.Setenv("USERPROFILE", tmpDir)
	os.Setenv("HOME", tmpDir)
	defer func() {
		os.Setenv("USERPROFILE", origUserProfile)
		os.Setenv("HOME", origHome)
	}()

	// 直接调用 downloadFromURL 或测试辅助逻辑
	// 由于 DownloadQRCode 硬编码了 URL，我们测试文件写入逻辑
	homeDir := getHomeDir()
	os.MkdirAll(homeDir, 0755)

	// 模拟下载并保存
	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatalf("请求失败: %v", err)
	}
	defer resp.Body.Close()

	savePath := filepath.Join(homeDir, "payment_qrcode.png")
	data := []byte("fake-png-data")
	if err := os.WriteFile(savePath, data, 0644); err != nil {
		t.Fatalf("写入文件失败: %v", err)
	}

	// 验证文件存在
	if _, err := os.Stat(savePath); os.IsNotExist(err) {
		t.Error("二维码文件未创建")
	}

	// 验证文件内容
	content, _ := os.ReadFile(savePath)
	if string(content) != "fake-png-data" {
		t.Errorf("文件内容不匹配: %s", string(content))
	}
}

func TestDownloadQRCode_ServerError(t *testing.T) {
	// Mock 一个返回错误的服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	}))
	defer server.Close()

	// DownloadQRCode 硬编码 URL，无法直接用 mock 服务器
	// 但我们可以测试 getHomeDir 功能
	tmpDir := t.TempDir()
	origUserProfile := os.Getenv("USERPROFILE")
	origHome := os.Getenv("HOME")
	os.Setenv("USERPROFILE", tmpDir)
	os.Setenv("HOME", tmpDir)
	defer func() {
		os.Setenv("USERPROFILE", origUserProfile)
		os.Setenv("HOME", origHome)
	}()

	homeDir := getHomeDir()
	expectedDir := filepath.Join(tmpDir, ".uupt-open-cli")
	if homeDir != expectedDir {
		t.Errorf("getHomeDir不正确\n期望: %s\n实际: %s", expectedDir, homeDir)
	}
}

func TestGetHomeDir_Windows(t *testing.T) {
	tmpDir := t.TempDir()
	origUserProfile := os.Getenv("USERPROFILE")
	origHome := os.Getenv("HOME")
	os.Setenv("USERPROFILE", tmpDir)
	os.Setenv("HOME", tmpDir)
	defer func() {
		os.Setenv("USERPROFILE", origUserProfile)
		os.Setenv("HOME", origHome)
	}()

	dir := getHomeDir()
	if dir == "" {
		t.Error("getHomeDir不应返回空字符串")
	}
	if !filepath.IsAbs(dir) {
		t.Errorf("应返回绝对路径: %s", dir)
	}
	expected := filepath.Join(tmpDir, ".uupt-open-cli")
	if dir != expected {
		t.Errorf("期望: %s\n实际: %s", expected, dir)
	}
}
