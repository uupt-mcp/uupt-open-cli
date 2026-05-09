package qrcode

import (
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"uupt-open-cli/internal/config"
	"uupt-open-cli/internal/logger"
)

// DownloadQRCode 下载微信支付二维码
func DownloadQRCode(paymentUrl string) string {
	encoded := url.QueryEscape(paymentUrl)
	qrURL := "https://api.qrserver.com/v1/create-qr-code/?size=300x300&data=" + encoded

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(qrURL)
	if err != nil {
		logger.Warnf("下载二维码失败: %v", err)
		return ""
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Warnf("读取二维码数据失败: %v", err)
		return ""
	}

	homeDir := config.GetHomeDir()
	savePath := filepath.Join(homeDir, "payment_qrcode.png")

	if err := os.MkdirAll(homeDir, 0755); err != nil {
		logger.Warnf("创建目录失败: %v", err)
		return ""
	}

	if err := os.WriteFile(savePath, data, 0644); err != nil {
		logger.Warnf("保存二维码失败: %v", err)
		return ""
	}

	return savePath
}
