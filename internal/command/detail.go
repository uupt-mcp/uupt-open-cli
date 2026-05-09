package command

import (
	"fmt"
	"os"

	"uupt-open-cli/internal/api"
	"uupt-open-cli/internal/config"

	"github.com/spf13/cobra"
)

var detailCmd = &cobra.Command{
	Use:   "detail",
	Short: "查询订单详情",
	Run:   runDetail,
}

var detailOrderCode string

func init() {
	detailCmd.Flags().StringVar(&detailOrderCode, "order-code", "", "订单编号（必填）")
	detailCmd.MarkFlagRequired("order-code")
	RootCmd.AddCommand(detailCmd)
}

func runDetail(cmd *cobra.Command, args []string) {
	cfg := config.EnsureConfig(true)

	result, err := api.OrderDetail(cfg, detailOrderCode)
	if err != nil {
		fmt.Fprintf(os.Stderr, "查询订单详情失败: %s\n", err.Error())
		os.Exit(1)
	}
	printJSON(result)
}
