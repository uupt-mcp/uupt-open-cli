package command

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"

	"uupt-open-cli/internal/api"
	"uupt-open-cli/internal/config"
	"uupt-open-cli/internal/logger"

	"github.com/spf13/cobra"
)

// 淡定星期四活动太阳码图片远程链接（Markdown 可直接渲染）
const thursdayQrcodeURL = "https://otherfiles.uupt.com/skills/thursday-qrcode.jpg"

// 淡定星期四活动太阳码图片（随二进制内嵌分发，与 skill 仓库 assets 保持同步）
//go:embed assets/thursday-qrcode.jpg
var thursdayQrcodeData []byte

var couponCmd = &cobra.Command{
	Use:   "coupon",
	Short: "领取优惠券",
	Run:   runCoupon,
}

var couponSource int

func init() {
	couponCmd.Flags().IntVar(&couponSource, "source", 1, "领取来源（决定可领哪些券包，默认1）")
	RootCmd.AddCommand(couponCmd)
}

func runCoupon(cmd *cobra.Command, args []string) {
	cfg := config.EnsureConfig(true)

	result, err := api.ReceiveCouponPackages(cfg, couponSource)
	if err != nil {
		fmt.Fprintf(os.Stderr, "领取优惠券失败: %s\n", err.Error())
		os.Exit(1)
	}
	printJSON(result)

	body, _ := result["body"].(map[string]interface{})
	if body == nil {
		msg, _ := result["msg"].(string)
		if msg == "" {
			msg = "未知错误"
		}
		fmt.Fprintf(os.Stderr, "[ERROR] 领券失败: %s\n", msg)
		os.Exit(1)
	}

	couponList, _ := body["couponList"].([]interface{})
	newlyClaimed, _ := body["newlyClaimed"].(bool)

	fmt.Println("[COUPON_RESULT]")
	fmt.Printf("NEWLY_CLAIMED=%t\n", newlyClaimed)
	fmt.Printf("COUPON_COUNT=%d\n", len(couponList))

	if thursdayJoinAble, _ := body["thursdayJoinAble"].(bool); thursdayJoinAble {
		fmt.Println("THURSDAY_JOIN_ABLE=true")
		fmt.Printf("THURSDAY_QRCODE_URL=%s\n", thursdayQrcodeURL)
		if qrFile := writeThursdayQrcode(); qrFile != "" {
			fmt.Printf("THURSDAY_QRCODE_FILE=%s\n", qrFile)
		}
	}
}

// writeThursdayQrcode 将内嵌的淡定星期四活动太阳码图片写入本地，供平台以附件形式发送。
// 每次用内嵌数据覆盖写入，确保图片始终与当前二进制版本一致（图片更新过时不残留旧版）。
func writeThursdayQrcode() string {
	homeDir := config.GetHomeDir()
	savePath := filepath.Join(homeDir, "assets", "thursday-qrcode.jpg")

	if err := os.MkdirAll(filepath.Dir(savePath), 0755); err != nil {
		logger.Warnf("创建目录失败: %v", err)
		return ""
	}
	if err := os.WriteFile(savePath, thursdayQrcodeData, 0644); err != nil {
		logger.Warnf("写入星期四活动二维码失败: %v", err)
		return ""
	}

	return savePath
}
