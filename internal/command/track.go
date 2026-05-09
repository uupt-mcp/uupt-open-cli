package command

import (
	"fmt"
	"os"

	"uupt-open-cli/internal/api"
	"uupt-open-cli/internal/config"

	"github.com/spf13/cobra"
)

var trackCmd = &cobra.Command{
	Use:   "track",
	Short: "跑男实时追踪",
	Run:   runTrack,
}

var trackOrderCode string

func init() {
	trackCmd.Flags().StringVar(&trackOrderCode, "order-code", "", "订单编号（必填）")
	trackCmd.MarkFlagRequired("order-code")
	RootCmd.AddCommand(trackCmd)
}

func runTrack(cmd *cobra.Command, args []string) {
	cfg := config.EnsureConfig(true)

	result, err := api.DriverTrack(cfg, trackOrderCode)
	if err != nil {
		fmt.Fprintf(os.Stderr, "查询跑男追踪失败: %s\n", err.Error())
		os.Exit(1)
	}
	printJSON(result)
}
