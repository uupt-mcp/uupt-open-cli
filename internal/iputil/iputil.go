package iputil

import (
	"encoding/json"
	"io"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"
)

var ipv4Re = regexp.MustCompile(`\b(?:\d{1,3}\.){3}\d{1,3}\b`)

// GetPublicIP 获取公网IP。并发探测多个源，避免单个国外接口卡住。
func GetPublicIP() string {
	services := []struct {
		url   string
		field string
	}{
		{"https://myip.ipip.net", ""},
		{"https://ip.3322.net", ""},
		{"https://api.ipify.org?format=json", "ip"},
		{"https://api64.ipify.org?format=json", "ip"},
		{"https://httpbin.org/ip", "origin"},
		{"https://ipinfo.io/json", "ip"},
	}

	client := &http.Client{Timeout: 2 * time.Second}
	ch := make(chan string, 1)
	var wg sync.WaitGroup
	for _, svc := range services {
		wg.Add(1)
		go func(url, field string) {
			defer wg.Done()
			ip := fetchIP(client, url, field)
			if ip == "" {
				return
			}
			select {
			case ch <- ip:
			default:
			}
		}(svc.url, svc.field)
	}
	go func() {
		wg.Wait()
		close(ch)
	}()

	if ip, ok := <-ch; ok {
		return ip
	}
	return ""
}

func fetchIP(client *http.Client, url string, field string) string {
	resp, err := client.Get(url)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if err != nil {
		return ""
	}
	text := strings.TrimSpace(string(body))
	if text == "" {
		return ""
	}

	if field != "" {
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
		if strings.Contains(ip, ",") {
			ip = strings.TrimSpace(strings.Split(ip, ",")[0])
		}
		return strings.TrimSpace(ip)
	}

	if m := ipv4Re.FindString(text); m != "" {
		return m
	}
	return ""
}
