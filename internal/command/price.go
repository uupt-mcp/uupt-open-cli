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
	Short: "订单询价",
	Run:   runPrice,
}

var (
	priceFromAddress string
	priceToAddress   string
	priceCity        string
)

func init() {
	priceCmd.Flags().StringVar(&priceFromAddress, "from-address", "", "发货地址（必填）")
	priceCmd.Flags().StringVar(&priceToAddress, "to-address", "", "收货地址（必填）")
	priceCmd.Flags().StringVar(&priceCity, "city", "郑州市", "城市名称")
	priceCmd.MarkFlagRequired("from-address")
	priceCmd.MarkFlagRequired("to-address")
	RootCmd.AddCommand(priceCmd)
}

func runPrice(cmd *cobra.Command, args []string) {
	cfg := config.EnsureConfig(true)

	result, err := api.OrderPrice(cfg, priceFromAddress, priceToAddress, priceCity)
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
