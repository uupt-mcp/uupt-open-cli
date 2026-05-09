package iputil

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"
)

// GetPublicIP 获取公网IP
func GetPublicIP() string {
	services := []struct {
		url   string
		field string
	}{
		{"https://httpbin.org/ip", "origin"},
		{"https://ipinfo.io/json", "ip"},
		{"https://api64.ipify.org?format=json", "ip"},
		{"https://api.ipify.org?format=json", "ip"},
	}

	client := &http.Client{Timeout: 5 * time.Second}

	for _, svc := range services {
		ip := fetchIP(client, svc.url, svc.field)
		if ip != "" {
			return ip
		}
	}

	return ""
}

func fetchIP(client *http.Client, url string, field string) string {
	resp, err := client.Get(url)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return ""
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return ""
	}

	val, ok := result[field]
	if !ok {
		return ""
	}

	ip, ok := val.(string)
	if !ok || ip == "" {
		return ""
	}

	// httpbin.org origin 可能包含逗号分隔的多个IP，取第一个
	if strings.Contains(ip, ",") {
		ip = strings.TrimSpace(strings.Split(ip, ",")[0])
	}

	return strings.TrimSpace(ip)
}
