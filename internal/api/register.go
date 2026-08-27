package api

import (
	"uupt-open-cli/internal/config"
)

// SendSmsCode 发送短信验证码
func SendSmsCode(cfg *config.Config, mobile string, ip string, imageCode string) (map[string]interface{}, error) {
	bizParams := map[string]interface{}{
		"userMobile": mobile,
		"userIp":     ip,
	}
	if imageCode != "" {
		bizParams["imageCode"] = imageCode
	}

	return UnauthorizedRequest(cfg, "/openapi/v3/user/unauthorized/sendSmsCode", bizParams)
}

// Auth 完成授权获取openId
func Auth(cfg *config.Config, mobile string, ip string, smsCode string) (map[string]interface{}, error) {
	bizParams := map[string]interface{}{
		"userMobile": mobile,
		"userIp":     ip,
		"smsCode":    smsCode,
		"cityName":   "郑州市",
		"countyName": "",
	}

	return UnauthorizedRequest(cfg, "/openapi/v3/user/unauthorized/auth", bizParams)
}
