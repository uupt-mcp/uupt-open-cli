package command

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"uupt-open-cli/internal/config"

	"github.com/spf13/cobra"
)

const (
	latestVersionURL = "https://raw.githubusercontent.com/uupt-mcp/uupt-open-cli/refs/heads/main/latest"
	releaseBaseURL   = "https://github.com/uupt-mcp/uupt-open-cli/releases/download"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "检查并更新到最新版本",
	Run:   runUpdate,
}

func init() {
	RootCmd.AddCommand(updateCmd)
}

func runUpdate(cmd *cobra.Command, args []string) {
	currentVersion := Version
	if currentVersion == "dev" {
		fmt.Println("[INFO] 当前为开发版本，无法检查更新")
		return
	}

	// 1. 获取最新版本号
	fmt.Println("[INFO] 正在检查更新...")
	latestVersion, err := getLatestVersion()
	if err != nil {
		fmt.Printf("[ERROR] 获取最新版本失败: %s\n", err.Error())
		os.Exit(1)
	}

	fmt.Printf("[INFO] 当前版本: %s\n", currentVersion)
	fmt.Printf("[INFO] 最新版本: %s\n", latestVersion)

	if latestVersion == currentVersion {
		fmt.Println("[OK] 已是最新版本，无需更新")
		return
	}

	// 2. 下载并替换二进制
	fmt.Printf("[INFO] 正在更新 %s → %s ...\n", currentVersion, latestVersion)
	if err := selfUpdate(latestVersion); err != nil {
		fmt.Printf("[ERROR] 更新失败: %s\n", err.Error())
		fmt.Println("[提示] 可通过以下命令手动更新:")
		if runtime.GOOS == "windows" {
			fmt.Println("  irm https://raw.githubusercontent.com/uupt-mcp/uupt-open-cli/main/scripts/install.ps1 | iex")
		} else {
			fmt.Println("  curl -fsSL https://raw.githubusercontent.com/uupt-mcp/uupt-open-cli/main/scripts/install.sh | bash")
		}
		os.Exit(1)
	}

	fmt.Printf("[OK] 更新成功！%s → %s\n", currentVersion, latestVersion)
}

func getLatestVersion() (string, error) {
	resp, err := http.Get(latestVersionURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	version := strings.TrimSpace(string(body))
	if version == "" {
		return "", fmt.Errorf("版本号为空")
	}

	return version, nil
}

func selfUpdate(version string) error {
	platform, arch, ext, archiveExt := getPlatformInfo()
	installDir := config.GetHomeDir()

	// 构建下载 URL
	binaryName := fmt.Sprintf("uupt-open-cli-%s-%s-%s", version, platform, arch)
	if ext != "" {
		binaryName += ext
	}
	archiveName := fmt.Sprintf("uupt-open-cli-%s-%s-%s.%s", version, platform, arch, archiveExt)
	downloadURL := fmt.Sprintf("%s/v%s/%s", releaseBaseURL, version, archiveName)

	// 下载压缩包
	fmt.Printf("[INFO] 下载 %s ...\n", archiveName)
	tmpDir, err := os.MkdirTemp("", "uupt-update-*")
	if err != nil {
		return fmt.Errorf("创建临时目录失败: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	archivePath := filepath.Join(tmpDir, archiveName)
	if err := downloadFile(downloadURL, archivePath); err != nil {
		return fmt.Errorf("下载失败: %w", err)
	}

	// 解压
	fmt.Println("[INFO] 解压文件...")
	if err := extractArchive(archivePath, tmpDir, archiveExt); err != nil {
		return fmt.Errorf("解压失败: %w", err)
	}

	// 查找解压后的二进制文件
	newBinary := filepath.Join(tmpDir, binaryName)
	if _, err := os.Stat(newBinary); os.IsNotExist(err) {
		// 可能在子目录中
		matches, _ := filepath.Glob(filepath.Join(tmpDir, "**", "uupt-open-cli*"))
		for _, m := range matches {
			if !strings.HasSuffix(m, ".json") && !strings.HasSuffix(m, ".tar.gz") && !strings.HasSuffix(m, ".zip") {
				newBinary = m
				break
			}
		}
	}

	if _, err := os.Stat(newBinary); os.IsNotExist(err) {
		return fmt.Errorf("未找到二进制文件: %s", newBinary)
	}

	// 替换当前二进制
	currentExe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("获取当前路径失败: %w", err)
	}

	// Windows: 先重命名旧文件，再复制新文件
	targetPath := filepath.Join(installDir, "uupt-open-cli")
	if runtime.GOOS == "windows" {
		targetPath += ".exe"
	}

	if runtime.GOOS == "windows" {
		// Windows 下正在运行的可执行文件无法直接覆盖
		// 使用 rename + copy 策略
		oldPath := currentExe + ".old"
		os.Remove(oldPath)
		if err := os.Rename(currentExe, oldPath); err != nil {
			return fmt.Errorf("重命名旧文件失败: %w", err)
		}
		defer os.Remove(oldPath)

		if err := copyFile(newBinary, targetPath); err != nil {
			// 回滚
			os.Rename(oldPath, currentExe)
			return fmt.Errorf("复制新文件失败: %w", err)
		}
	} else {
		if err := os.Rename(newBinary, targetPath); err != nil {
			return fmt.Errorf("替换二进制失败: %w", err)
		}
		os.Chmod(targetPath, 0755)
	}

	return nil
}

func getPlatformInfo() (platform, arch, ext, archiveExt string) {
	switch runtime.GOOS {
	case "darwin":
		platform = "macos"
	case "windows":
		platform = "windows"
	default:
		platform = runtime.GOOS
	}

	arch = runtime.GOARCH
	if arch == "arm64" && platform == "macos" {
		arch = "arm64"
	}

	if platform == "windows" {
		ext = ".exe"
		archiveExt = "zip"
	} else {
		ext = ""
		archiveExt = "tar.gz"
	}

	return
}

func downloadFile(url, dest string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	f, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, resp.Body)
	return err
}

func extractArchive(archivePath, destDir, archiveExt string) error {
	if archiveExt == "zip" {
		// 使用系统 tar 解压 zip（Windows 10+ 自带 tar 支持 zip）
		cmd := exec.Command("tar", "-xf", archivePath, "-C", destDir)
		return cmd.Run()
	}
	// tar.gz
	cmd := exec.Command("tar", "-xzf", archivePath, "-C", destDir)
	return cmd.Run()
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}

	// 复制可执行权限
	info, _ := os.Stat(src)
	if info != nil {
		os.Chmod(dst, info.Mode())
	}

	return nil
}

// checkForUpdate 检查是否有新版本（用于其他命令的提示）
func checkForUpdate() {
	currentVersion := Version
	if currentVersion == "dev" {
		return
	}

	latest, err := getLatestVersion()
	if err != nil || latest == currentVersion {
		return
	}

	data := map[string]string{
		"current": currentVersion,
		"latest":  latest,
	}
	hint, _ := json.Marshal(data)
	fmt.Printf("\n[UPDATE_AVAILABLE] %s\n", string(hint))
	fmt.Printf("提示: 运行 uupt-open-cli update 更新到 %s\n\n", latest)
}
