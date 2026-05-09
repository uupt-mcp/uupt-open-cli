package command

import (
	"fmt"
	"os"

	"uupt-open-cli/internal/api"
	"uupt-open-cli/internal/config"
	"uupt-open-cli/internal/iputil"
	"uupt-open-cli/internal/logger"

	"github.com/spf13/cobra"
)

var registerCmd = &cobra.Command{
	Use:   "register",
	Short: "手机号注册/获取授权",
	Run:   runRegister,
}

var (
	regMobile    string
	regSmsCode   string
	regImageCode string
	regIP        string
)

func init() {
	registerCmd.Flags().StringVar(&regMobile, "mobile", "", "手机号（必填）")
	registerCmd.Flags().StringVar(&regSmsCode, "sms-code", "", "短信验证码")
	registerCmd.Flags().StringVar(&regImageCode, "image-code", "", "图片验证码")
	registerCmd.Flags().StringVar(&regIP, "ip", "", "公网IP（默认自动检测）")
	registerCmd.MarkFlagRequired("mobile")
	RootCmd.AddCommand(registerCmd)
}

func runRegister(cmd *cobra.Command, args []string) {
	cfg := config.EnsureConfig(false)

	ip := regIP
	if ip == "" {
		ip = iputil.GetPublicIP()
		logger.Infof("检测到公网IP: %s", ip)
	}

	if regSmsCode != "" {
		// 步骤二：提交验证码完成授权
		result, err := api.Auth(cfg, regMobile, ip, regSmsCode)
		if err != nil {
			fmt.Printf("[REGISTRATION_FAILED]\n%s\n", err.Error())
			os.Exit(1)
		}
		// 提取 openId
		body, _ := result["body"].(map[string]interface{})
		if body != nil {
			if openId, ok := body["openId"].(string); ok && openId != "" {
				config.SaveConfig(map[string]string{"openId": openId})
				fmt.Printf("[REGISTRATION_SUCCESS]\n[成功] 注册成功！openId 已保存到配置文件。\n   openId: %s\n", openId)
				os.Exit(0)
			}
		}
		fmt.Printf("[REGISTRATION_FAILED]\n授权失败，未获取到openId\n")
		os.Exit(1)
	} else {
		// 步骤一：发送短信验证码
		result, err := api.SendSmsCode(cfg, regMobile, ip, regImageCode)
		if err != nil {
			fmt.Printf("发送验证码失败: %s\n", err.Error())
			os.Exit(1)
		}
		// UU API: code=1 表示成功, code=0 也视为成功
		code, _ := result["code"].(float64)
		if int(code) == 88100106 {
			body, _ := result["body"].(map[string]interface{})
			imageData := ""
			if body != nil {
				imageData, _ = body["imageData"].(string)
			}
			fmt.Printf("[IMAGE_CAPTCHA_REQUIRED]\nIMAGE_DATA=%s\n", imageData)
			os.Exit(2)
		}
		if int(code) != 0 && int(code) != 1 {
			msg, _ := result["msg"].(string)
			fmt.Printf("发送验证码失败: code=%d, msg=%s\n", int(code), msg)
			os.Exit(1)
		}
		fmt.Println("[SMS_SENT]")
	}
}
