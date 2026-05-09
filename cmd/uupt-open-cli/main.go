package main

import (
	"os"

	"uupt-open-cli/internal/command"
	"uupt-open-cli/internal/logger"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	// 初始化日志
	logger.Init()

	// 设置版本信息
	command.Version = version
	command.Commit = commit
	command.Date = date

	// 执行命令
	if err := command.RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
