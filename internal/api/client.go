package api

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	jsoniter "github.com/json-iterator/go"

	"uupt-open-cli/internal/config"
	"uupt-open-cli/internal/logger"
)

var jsonAPI = jsoniter.Config{
	EscapeHTML:             false,
	SortMapKeys:            true,
	ValidateJsonRawMessage: true,
}.Froze()

func GenerateSign(bizParams map[string]interface{}, appSecret string, timestamp int) string {
	bizJson, _ := jsonAPI.MarshalToString(bizParams)
	signStr := bizJson + appSecret + strconv.Itoa(timestamp)
	hash := md5.Sum([]byte(signStr))
	return strings.ToUpper(fmt.Sprintf("%x", hash))
}

func AuthorizedRequest(cfg *config.Config, apiPath string, bizParams map[string]interface{}) (map[string]interface{}, error) {
	timestamp := int(time.Now().Unix())
	bizJson, _ := jsonAPI.MarshalToString(bizParams)
	sign := GenerateSign(bizParams, cfg.AppSecret, timestamp)

	body := map[string]interface{}{
		"openId":    cfg.OpenId,
		"timestamp": timestamp,
		"biz":       bizJson,
		"sign":      sign,
	}

	return doRequest(cfg, apiPath, body)
}

func UnauthorizedRequest(cfg *config.Config, apiPath string, bizParams map[string]interface{}) (map[string]interface{}, error) {
	timestamp := int(time.Now().Unix())
	bizJson, _ := jsonAPI.MarshalToString(bizParams)
	sign := GenerateSign(bizParams, cfg.AppSecret, timestamp)

	body := map[string]interface{}{
		"timestamp": timestamp,
		"biz":       bizJson,
		"sign":      sign,
	}

	return doRequest(cfg, apiPath, body)
}

// buildRequestURL 构造请求完整 URL。
// 基址默认为不带路径前缀的域名（如 https://api-open.uupt.com），接口路径在调用处写全；
// 同时兼容旧配置中带 /openapi/v3/ 后缀的 apiUrl。
func buildRequestURL(cfg *config.Config, apiPath string) string {
	base := strings.TrimRight(cfg.ApiUrl, "/")
	base = strings.TrimSuffix(base, "/openapi/v3")
	if !strings.HasPrefix(apiPath, "/") {
		apiPath = "/" + apiPath
	}
	return base + apiPath
}

func doRequest(cfg *config.Config, apiPath string, body map[string]interface{}) (map[string]interface{}, error) {
	url := buildRequestURL(cfg, apiPath)

	jsonBody, err := jsonAPI.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	logger.Infof("POST %s", url)
	logger.Debugf("Request body: %s", string(jsonBody))

	req, err := http.NewRequest("POST", url, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("X-App-Id", cfg.AppId)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	logger.Debugf("Response [%d]: %s", resp.StatusCode, string(respBody))

	var result map[string]interface{}
	if err := jsonAPI.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response JSON: %w", err)
	}

	return result, nil
}
