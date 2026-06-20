package command

import (
	"fmt"
	"os"

	"uupt-open-cli/internal/api"
	"uupt-open-cli/internal/config"
	"uupt-open-cli/internal/qrcode"

	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "创建订单（支持跑腿配送和帮忙服务）",
	Run:   runCreate,
}

var (
	createPriceToken    string
	createReceiverPhone string
	createChannel       string
	createNote          string
)

func init() {
	createCmd.Flags().StringVar(&createPriceToken, "price-token", "", "询价token（必填）")
	createCmd.Flags().StringVar(&createReceiverPhone, "receiver-phone", "", "收件人手机号（必填）")
	createCmd.Flags().StringVar(&createChannel, "channel", "", "支付渠道（可选，wechat表示微信）")
	createCmd.Flags().StringVar(&createNote, "note", "", "帮忙内容描述（帮忙订单时必填）")
	createCmd.MarkFlagRequired("price-token")
	createCmd.MarkFlagRequired("receiver-phone")
	RootCmd.AddCommand(createCmd)
}

func runCreate(cmd *cobra.Command, args []string) {
	cfg := config.EnsureConfig(true)

	result, err := api.CreateOrder(cfg, createPriceToken, createReceiverPhone, createChannel, createNote)
	if err != nil {
		fmt.Fprintf(os.Stderr, "创建订单失败: %s\n", err.Error())
		os.Exit(1)
	}

	// 检查是否需要支付（余额不足）
	body, _ := result["body"].(map[string]interface{})
	if body != nil {
		orderUrl, _ := body["orderUrl"].(string)
		if orderUrl != "" {
			orderCode, _ := body["orderCode"].(string)
			fmt.Println("[PAYMENT_REQUIRED]")
			fmt.Printf("ORDER_CODE=%s\n", orderCode)
			fmt.Printf("PAYMENT_URL=%s\n", orderUrl)
			if createChannel == "wechat" {
				qrPath := qrcode.DownloadQRCode(orderUrl)
				if qrPath != "" {
					fmt.Printf("QRCODE_FILE=%s\n", qrPath)
				}
			}
			return
		}
	}

	printJSON(result)
}
