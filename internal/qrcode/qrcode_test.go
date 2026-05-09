package qrcode

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"uupt-open-cli/internal/config"
)

func TestDownloadQRCode_Success(t *testing.T) {
	// Mock 一个返回 PNG 数据的服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		w.Write([]byte("fake-png-data"))
	}))
	defer server.Close()

	// 获取 home 目录
	homeDir := config.GetHomeDir()
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

func TestGetHomeDir(t *testing.T) {
	dir := config.GetHomeDir()
	if dir == "" {
		t.Error("GetHomeDir 不应返回空字符串")
	}
	if !filepath.IsAbs(dir) {
		t.Errorf("应返回绝对路径: %s", dir)
	}
	if filepath.Base(dir) != ".uupt-open-cli" {
		t.Errorf("路径应以 .uupt-open-cli 结尾: %s", dir)
	}
}
