package command

import (
	"encoding/json"
	"fmt"
	"os"

	"uupt-open-cli/internal/api"
	"uupt-open-cli/internal/config"

	"github.com/spf13/cobra"
)

var priceCmd = &cobra.Command{
	Use:   "price",
	Short: "订单询价（支持跑腿配送和帮忙服务）",
	Run:   runPrice,
}

var (
	priceFromAddress string
	priceToAddress   string
	priceCity        string
	priceOrderType   string
)

func init() {
	priceCmd.Flags().StringVar(&priceFromAddress, "from-address", "", "发货地址（必填，帮忙订单时填写帮忙地点）")
	priceCmd.Flags().StringVar(&priceToAddress, "to-address", "", "收货地址（帮忙订单时无需填写，自动使用发货地址）")
	priceCmd.Flags().StringVar(&priceCity, "city", "郑州市", "城市名称")
	priceCmd.Flags().StringVar(&priceOrderType, "order-type", "send", "订单类型：send=跑腿配送，help=帮忙服务")
	priceCmd.MarkFlagRequired("from-address")
	RootCmd.AddCommand(priceCmd)
}

func runPrice(cmd *cobra.Command, args []string) {
	cfg := config.EnsureConfig(true)

	// 帮忙订单时，to-address 使用 from-address 的值
	toAddr := priceToAddress
	if priceOrderType == "help" && toAddr == "" {
		toAddr = priceFromAddress
	}
	if toAddr == "" {
		fmt.Fprintln(os.Stderr, "错误: 跑腿配送订单必须指定 --to-address 参数")
		os.Exit(1)
	}

	result, err := api.OrderPrice(cfg, priceFromAddress, toAddr, priceCity, priceOrderType)
	if err != nil {
		fmt.Fprintf(os.Stderr, "询价失败: %s\n", err.Error())
		os.Exit(1)
	}
	printJSON(result)
}

func printJSON(data interface{}) {
	output, err := json.Marshal(data)
	if err != nil {
		fmt.Fprintf(os.Stderr, "JSON序列化错误: %s\n", err.Error())
		os.Exit(1)
	}
	fmt.Println(string(output))
}
