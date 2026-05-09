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
	Long:  "卸载 uupt-open-cli，删除安装目录、PATH配置和go install二进制文件",
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
		fmt.Printf("将删除以下内容:\n")
		fmt.Printf("  安装目录: %s\n", installDir)
		if goBinPath := getGoInstallBinaryPath(); goBinPath != "" {
			if _, err := os.Stat(goBinPath); err == nil {
				fmt.Printf("  go install 二进制: %s\n", goBinPath)
			}
		}
		fmt.Println()
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

	// 2. 删除 go install 二进制文件
	removeGoInstallBinary()

	// 3. 删除安装目录
	removeInstallDir(installDir)

	fmt.Println("[OK] 卸载完成！请重新打开终端使 PATH 生效。")
}

// getGoInstallBinaryPath 返回 go install 安装的二进制文件路径
func getGoInstallBinaryPath() string {
	gobin := os.Getenv("GOBIN")
	if gobin == "" {
		gopath := os.Getenv("GOPATH")
		if gopath == "" {
			home, _ := os.UserHomeDir()
			gopath = filepath.Join(home, "go")
		}
		gobin = filepath.Join(gopath, "bin")
	}

	binaryName := "uupt-open-cli"
	if runtime.GOOS == "windows" {
		binaryName = "uupt-open-cli.exe"
	}

	return filepath.Join(gobin, binaryName)
}

// removeGoInstallBinary 删除通过 go install 安装的二进制文件
func removeGoInstallBinary() {
	binaryPath := getGoInstallBinaryPath()
	if binaryPath == "" {
		return
	}

	if _, err := os.Stat(binaryPath); err != nil {
		return // 文件不存在，跳过
	}

	// 检查是否为当前运行的二进制
	if isRunningFromPath(binaryPath) {
		fmt.Printf("[WARN] go install 二进制文件正在运行，无法删除: %s\n", binaryPath)
		fmt.Printf("[提示] 请关闭后手动删除: %s\n", binaryPath)
		return
	}

	if err := os.Remove(binaryPath); err != nil {
		fmt.Printf("[WARN] 无法删除 go install 二进制文件: %s\n", err.Error())
		fmt.Printf("[提示] 请手动删除: %s\n", binaryPath)
	} else {
		fmt.Printf("[INFO] 已删除 go install 二进制文件: %s\n", binaryPath)
	}
}

// isRunningFromPath 检查当前运行的二进制是否在指定路径
func isRunningFromPath(path string) bool {
	exe, err := os.Executable()
	if err != nil {
		return false
	}
	return strings.EqualFold(filepath.Clean(exe), filepath.Clean(path))
}

// isRunningFromDir 检查当前运行的二进制是否在指定目录内
func isRunningFromDir(dir string) bool {
	exe, err := os.Executable()
	if err != nil {
		return false
	}
	return strings.HasPrefix(strings.ToLower(filepath.Clean(exe)), strings.ToLower(filepath.Clean(dir))+string(os.PathSeparator))
}

// removeInstallDir 删除安装目录，处理 Windows 下自删除问题
func removeInstallDir(installDir string) {
	if isRunningFromDir(installDir) && runtime.GOOS == "windows" {
		removeInstallDirSelfWindows(installDir)
		return
	}

	if err := os.RemoveAll(installDir); err != nil {
		fmt.Printf("[ERROR] 删除目录失败: %s\n", err.Error())
		os.Exit(1)
	}
	fmt.Println("[INFO] 已删除安装目录")
}

// removeInstallDirSelfWindows 处理 Windows 下从安装目录运行时的删除
// 通过启动后台 PowerShell 进程，在当前进程退出后自动删除安装目录
func removeInstallDirSelfWindows(installDir string) {
	exe, err := os.Executable()
	if err != nil {
		fmt.Printf("[WARN] 无法确定当前运行路径，尝试直接删除...\n")
		if err := os.RemoveAll(installDir); err != nil {
			fmt.Printf("[ERROR] 删除目录失败: %s\n", err.Error())
		}
		return
	}

	exePath := filepath.Clean(exe)

	// 尽力删除安装目录中除当前运行二进制外的所有文件
	entries, err := os.ReadDir(installDir)
	if err != nil {
		fmt.Printf("[ERROR] 读取安装目录失败: %s\n", err.Error())
		return
	}

	for _, entry := range entries {
		entryPath := filepath.Join(installDir, entry.Name())
		// 跳过当前运行的二进制文件
		if strings.EqualFold(entryPath, exePath) {
			continue
		}
		os.RemoveAll(entryPath)
	}

	// 创建临时清理脚本，在当前进程退出后自动删除安装目录
	pid := os.Getpid()
	tmpScript := filepath.Join(os.TempDir(), fmt.Sprintf("uupt-cleanup-%d.ps1", pid))
	scriptContent := fmt.Sprintf(
		"$p = Get-Process -Id %d -ErrorAction SilentlyContinue\nif ($p) { Wait-Process -Id %d -ErrorAction SilentlyContinue }\nRemove-Item -Recurse -Force -LiteralPath '%s' -ErrorAction SilentlyContinue\nRemove-Item -Force -LiteralPath '%s' -ErrorAction SilentlyContinue\n",
		pid, pid,
		strings.ReplaceAll(installDir, "'", "''"),
		strings.ReplaceAll(tmpScript, "'", "''"),
	)

	if err := os.WriteFile(tmpScript, []byte(scriptContent), 0644); err != nil {
		fmt.Println("[WARN] 当前正在从安装目录运行，无法删除正在使用的二进制文件")
		fmt.Printf("[提示] 请关闭终端后手动删除: %s\n", installDir)
		return
	}

	psExe := "pwsh"
	if _, err := exec.LookPath("pwsh"); err != nil {
		psExe = "powershell"
	}

	cmd := exec.Command(psExe, "-NoProfile", "-WindowStyle", "Hidden", "-ExecutionPolicy", "Bypass", "-File", tmpScript)
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.Stdin = nil

	if err := cmd.Start(); err != nil {
		os.Remove(tmpScript) // 清理临时文件
		fmt.Println("[WARN] 当前正在从安装目录运行，无法删除正在使用的二进制文件")
		fmt.Printf("[提示] 请关闭终端后手动删除: %s\n", installDir)
		return
	}

	// 释放子进程，使其独立运行
	cmd.Process.Release()

	fmt.Println("[INFO] 已安排在进程退出后自动清理安装目录")
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

	// 优先使用 PowerShell Core (pwsh)，降级到 Windows PowerShell (powershell)
	psExe := "pwsh"
	if _, err := exec.LookPath("pwsh"); err != nil {
		psExe = "powershell"
	}

	cmd := exec.Command(psExe, "-NoProfile", "-Command", psScript)
	if err := cmd.Run(); err != nil {
		fmt.Printf("[WARN] 无法自动移除 PATH: %s\n", err.Error())
		fmt.Println("[提示] 请手动从系统环境变量 PATH 中移除:")
		fmt.Printf("  %s\n", installDir)
		return
	}

	fmt.Println("[INFO] 已从用户 PATH 中移除安装目录")
}

func removeFromUnixPath(installDir string) {
	shellConfigs := []string{
		filepath.Join(getUnixHome(), ".bashrc"),
		filepath.Join(getUnixHome(), ".zshrc"),
		filepath.Join(getUnixHome(), ".config", "fish", "config.fish"),
	}

	for _, configPath := range shellConfigs {
		data, err := os.ReadFile(configPath)
		if err != nil {
			continue
		}

		lines := strings.Split(string(data), "\n")
		linesToRemove := make(map[int]bool)

		// 标记包含安装目录的 PATH 行
		for i, line := range lines {
			if strings.Contains(strings.TrimSpace(line), installDir) {
				linesToRemove[i] = true
				// 仅标记紧邻 PATH 行上方的关联注释
				if i > 0 && strings.TrimSpace(lines[i-1]) == "# UU跑腿 CLI" {
					linesToRemove[i-1] = true
				}
			}
		}

		if len(linesToRemove) == 0 {
			continue
		}

		var newLines []string
		for i, line := range lines {
			if !linesToRemove[i] {
				newLines = append(newLines, line)
			}
		}

		output := strings.Join(newLines, "\n")
		output = strings.TrimRight(output, "\n") + "\n"
		if err := os.WriteFile(configPath, []byte(output), 0644); err != nil {
			fmt.Printf("[WARN] 无法更新 %s: %s\n", configPath, err.Error())
		} else {
			fmt.Printf("[INFO] 已从 %s 中移除 PATH 配置\n", configPath)
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
