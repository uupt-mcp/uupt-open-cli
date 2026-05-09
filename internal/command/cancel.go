package command

import (
	"fmt"
	"os"

	"uupt-open-cli/internal/api"
	"uupt-open-cli/internal/config"

	"github.com/spf13/cobra"
)

var cancelCmd = &cobra.Command{
	Use:   "cancel",
	Short: "取消订单",
	Run:   runCancel,
}

var (
	cancelOrderCode string
	cancelReason    string
)

func init() {
	cancelCmd.Flags().StringVar(&cancelOrderCode, "order-code", "", "订单编号（必填）")
	cancelCmd.Flags().StringVar(&cancelReason, "reason", "", "取消原因（可选）")
	cancelCmd.MarkFlagRequired("order-code")
	RootCmd.AddCommand(cancelCmd)
}

func runCancel(cmd *cobra.Command, args []string) {
	cfg := config.EnsureConfig(true)

	result, err := api.CancelOrder(cfg, cancelOrderCode, cancelReason)
	if err != nil {
		fmt.Fprintf(os.Stderr, "取消订单失败: %s\n", err.Error())
		os.Exit(1)
	}
	printJSON(result)
}
