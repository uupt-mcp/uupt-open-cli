package command

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

var showVersion bool
var showList bool

var RootCmd = &cobra.Command{
	Use:   "uupt-open-cli",
	Short: "UU跑腿开放平台CLI工具",
	Long:  "UU跑腿开放平台CLI工具 - 为AI智能体提供同城即时配送能力",
	Run: func(cmd *cobra.Command, args []string) {
		if showVersion {
			fmt.Printf("uupt-open-cli %s (commit: %s, built: %s)\n", Version, Commit, Date)
			return
		}
		if showList {
			fmt.Println("支持的命令:")
			fmt.Println("  register    手机号注册/获取授权")
			fmt.Println("  price       订单询价")
			fmt.Println("  create      创建订单")
			fmt.Println("  detail      查询订单详情")
			fmt.Println("  cancel      取消订单")
			fmt.Println("  track       跑男实时追踪")
			fmt.Println("  update      检查并更新到最新版本")
			return
		}
		cmd.Help()
	},
}

func init() {
	RootCmd.Flags().BoolVar(&showVersion, "version", false, "显示版本号")
	RootCmd.Flags().BoolVar(&showList, "list", false, "列出所有支持的命令")
}
