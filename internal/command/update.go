package command

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"uupt-open-cli/internal/config"

	"github.com/spf13/cobra"
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
			fmt.Println("  curl.exe -fsSL --tlsv1.2 -o $env:TEMP\\uupt-install.ps1 https://cdn.jsdelivr.net/gh/uupt-mcp/uupt-open-cli@main/scripts/install.ps1")
			fmt.Println("  powershell -NoProfile -ExecutionPolicy Bypass -File $env:TEMP\\uupt-install.ps1")
		} else {
			fmt.Println("  curl -fsSL https://cdn.jsdelivr.net/gh/uupt-mcp/uupt-open-cli@main/scripts/install.sh | bash")
		}
		os.Exit(1)
	}

	fmt.Printf("[OK] 更新成功！%s → %s\n", currentVersion, latestVersion)
}

func getLatestVersion() (string, error) {
	var lastErr error
	for _, u := range latestVersionURLs() {
		body, err := fetchURL(u, 20*time.Second)
		if err != nil {
			lastErr = err
			continue
		}
		version := strings.TrimSpace(string(body))
		if version == "" {
			lastErr = fmt.Errorf("版本号为空")
			continue
		}
		return version, nil
	}
	if lastErr == nil {
		return "", fmt.Errorf("版本号为空")
	}
	return "", lastErr
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

	// 下载压缩包（优先 GitHub 加速代理）
	fmt.Printf("[INFO] 下载 %s ...\n", archiveName)
	tmpDir, err := os.MkdirTemp("", "uupt-update-*")
	if err != nil {
		return fmt.Errorf("创建临时目录失败: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	archivePath := filepath.Join(tmpDir, archiveName)
	if err := downloadFileWithFallback(githubReleaseURLs(downloadURL), archivePath); err != nil {
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
		// 在子目录中递归查找
		filepath.WalkDir(tmpDir, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if d.IsDir() {
				return nil
			}
			base := filepath.Base(path)
			if strings.HasPrefix(base, "uupt-open-cli") &&
				!strings.HasSuffix(path, ".json") &&
				!strings.HasSuffix(path, ".tar.gz") &&
				!strings.HasSuffix(path, ".zip") {
				newBinary = path
				return filepath.SkipAll
			}
			return nil
		})
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

func extractArchive(archivePath, destDir, archiveExt string) error {
	if archiveExt == "zip" {
		return extractZip(archivePath, destDir)
	}
	return extractTarGz(archivePath, destDir)
}

func extractZip(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		fpath := filepath.Join(dest, f.Name)

		// 安全检查：防止 zip slip 路径遍历
		if !strings.HasPrefix(filepath.Clean(fpath), filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("非法的文件路径: %s", f.Name)
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, 0755)
			continue
		}

		if err := os.MkdirAll(filepath.Dir(fpath), 0755); err != nil {
			return err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func extractTarGz(src, dest string) error {
	f, err := os.Open(src)
	if err != nil {
		return err
	}
	defer f.Close()

	gzr, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		fpath := filepath.Join(dest, hdr.Name)

		// 安全检查：防止 tar slip 路径遍历
		if !strings.HasPrefix(filepath.Clean(fpath), filepath.Clean(dest)+string(os.PathSeparator)) {
			return fmt.Errorf("非法的文件路径: %s", hdr.Name)
		}

		switch hdr.Typeflag {
		case tar.TypeDir:
			os.MkdirAll(fpath, 0755)
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(fpath), 0755); err != nil {
				return err
			}
			outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.FileMode(hdr.Mode))
			if err != nil {
				return err
			}
			if _, err := io.Copy(outFile, tr); err != nil {
				outFile.Close()
				return err
			}
			outFile.Close()
		}
	}
	return nil
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
