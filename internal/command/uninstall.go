package command

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"uupt-open-cli/internal/config"

	"github.com/spf13/cobra"
)

var uninstallForce bool

var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "卸载 uupt-open-cli",
	Long:  "卸载 uupt-open-cli，删除安装目录和PATH配置",
	Run:   runUninstall,
}

func init() {
	uninstallCmd.Flags().BoolVarP(&uninstallForce, "force", "y", false, "跳过确认提示")
	RootCmd.AddCommand(uninstallCmd)
}

func runUninstall(cmd *cobra.Command, args []string) {
	installDir := config.GetHomeDir()

	// 确认
	if !uninstallForce {
		fmt.Printf("将删除以下目录及所有内容:\n  %s\n\n", installDir)
		fmt.Print("确认卸载? [y/N] ")
		var answer string
		fmt.Scanln(&answer)
		if strings.ToLower(strings.TrimSpace(answer)) != "y" {
			fmt.Println("已取消")
			return
		}
	}

	// 1. 从 PATH 中移除
	removeFromPath(installDir)

	// 2. 删除安装目录
	if err := os.RemoveAll(installDir); err != nil {
		fmt.Printf("[ERROR] 删除目录失败: %s\n", err.Error())
		os.Exit(1)
	}

	fmt.Println("[OK] 卸载完成！请重新打开终端使 PATH 生效。")
}

func removeFromPath(installDir string) {
	switch runtime.GOOS {
	case "windows":
		removeFromWindowsPath(installDir)
	default:
		removeFromUnixPath(installDir)
	}
}

func removeFromWindowsPath(installDir string) {
	currentPath := os.Getenv("PATH")
	if !strings.Contains(strings.ToLower(currentPath), strings.ToLower(installDir)) {
		return
	}

	// 通过 PowerShell 从用户 PATH 环境变量中移除
	psScript := fmt.Sprintf(
		`$p = [Environment]::GetEnvironmentVariable('PATH','User'); $p = ($p -split ';' | Where-Object { $_ -ne '%s' }) -join ';'; [Environment]::SetEnvironmentVariable('PATH',$p,'User')`,
		installDir,
	)

	cmd := exec.Command("powershell", "-NoProfile", "-Command", psScript)
	if err := cmd.Run(); err != nil {
		fmt.Printf("[WARN] 无法自动移除 PATH: %s\n", err.Error())
		fmt.Println("[提示] 请手动从系统环境变量 PATH 中移除:")
		fmt.Printf("  %s\n", installDir)
		return
	}

	fmt.Println("[INFO] 已从用户 PATH 中移除安装目录")
}

func removeFromUnixPath(installDir string) {
	shellConfigs := []struct {
		path    string
		content string
	}{
		{filepath.Join(getUnixHome(), ".bashrc"), fmt.Sprintf(`export PATH="%s:$PATH"`, installDir)},
		{filepath.Join(getUnixHome(), ".zshrc"), fmt.Sprintf(`export PATH="%s:$PATH"`, installDir)},
		{filepath.Join(getUnixHome(), ".config", "fish", "config.fish"), fmt.Sprintf(`set -gx PATH %s $PATH`, installDir)},
	}

	for _, sc := range shellConfigs {
		data, err := os.ReadFile(sc.path)
		if err != nil {
			continue
		}

		lines := strings.Split(string(data), "\n")
		var newLines []string
		modified := false

		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			// 跳过包含安装目录 PATH 的行，以及紧邻的注释行
			if strings.Contains(trimmed, installDir) {
				modified = true
				continue
			}
			// 跳过 "# UU跑腿 CLI" 注释行
			if trimmed == "# UU跑腿 CLI" {
				modified = true
				continue
			}
			newLines = append(newLines, line)
		}

		if modified {
			output := strings.Join(newLines, "\n")
			// 清理末尾多余空行
			output = strings.TrimRight(output, "\n") + "\n"
			if err := os.WriteFile(sc.path, []byte(output), 0644); err != nil {
				fmt.Printf("[WARN] 无法更新 %s: %s\n", sc.path, err.Error())
			} else {
				fmt.Printf("[INFO] 已从 %s 中移除 PATH 配置\n", sc.path)
			}
		}
	}
}

func getUnixHome() string {
	home := os.Getenv("HOME")
	if home == "" {
		home, _ = os.UserHomeDir()
	}
	return home
}
