package command

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const (
	jsdelivrLatestURL = "https://cdn.jsdelivr.net/gh/uupt-mcp/uupt-open-cli@main/latest"
	latestVersionURL  = "https://raw.githubusercontent.com/uupt-mcp/uupt-open-cli/refs/heads/main/latest"
	releaseBaseURL    = "https://github.com/uupt-mcp/uupt-open-cli/releases/download"
)

func githubReleaseURLs(githubURL string) []string {
	// 国内实测 ghfast.top 明显快于 ghproxy；直连放第二，慢代理垫底。
	return []string{
		"https://ghfast.top/" + githubURL,
		githubURL,
		"https://ghproxy.net/" + githubURL,
		"https://mirror.ghproxy.com/" + githubURL,
	}
}

func latestVersionURLs() []string {
	return []string{
		jsdelivrLatestURL,
		"https://ghfast.top/" + latestVersionURL,
		"https://ghproxy.net/" + latestVersionURL,
		latestVersionURL,
	}
}

func fetchURL(url string, timeout time.Duration) ([]byte, error) {
	client := &http.Client{Timeout: timeout}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}

func downloadFileWithFallback(urls []string, dest string) error {
	var lastErr error
	for _, u := range urls {
		fmt.Printf("[INFO] 尝试 %s\n", u)
		if err := downloadFile(u, dest); err != nil {
			lastErr = err
			continue
		}
		return nil
	}
	if lastErr == nil {
		return fmt.Errorf("没有可用的下载地址")
	}
	return lastErr
}

func downloadFile(url, dest string) error {
	client := &http.Client{Timeout: 180 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	return err
}
