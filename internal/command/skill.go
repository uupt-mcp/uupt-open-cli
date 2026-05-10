package command

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

const skillName = "uupt"

// agentDirs lists all supported agent skill directories (relative to root).
// Parent-directory gate: only install into directories whose parent exists.
var agentDirs = []string{
	".agents/skills",
	".claude/skills",
	".cursor/skills",
	".gemini/skills",
	".codex/skills",
	".github/skills",
	".windsurf/skills",
	".augment/skills",
	".cline/skills",
	".amp/skills",
	".kiro/skills",
	".trae/skills",
	".openclaw/skills",
	".hermes/skills",
	".qoder/skills",
}

var (
	skillRoot string
)

var skillCmd = &cobra.Command{
	Use:   "skill",
	Short: "管理 Agent Skill",
	Long:  "管理 AI 智能体 Skill 文件的安装和卸载",
}

var skillInstallCmd = &cobra.Command{
	Use:   "install",
	Short: "安装 Agent Skill 到智能体目录",
	Long:  "从 GitHub Releases 下载 uupt-skills.zip 并安装到本地智能体目录",
	Run:   runSkillInstall,
}

var skillUninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "卸载 Agent Skill",
	Long:  "从所有智能体目录中移除 uupt skill 文件",
	Run:   runSkillUninstall,
}

func init() {
	skillInstallCmd.Flags().StringVar(&skillRoot, "root", "", "安装根目录（默认为当前目录）")
	skillCmd.AddCommand(skillInstallCmd)
	skillCmd.AddCommand(skillUninstallCmd)
	RootCmd.AddCommand(skillCmd)
}

func runSkillInstall(cmd *cobra.Command, args []string) {
	root := skillRoot
	if root == "" {
		root, _ = os.Getwd()
	}

	// 1. 获取版本
	fmt.Println("[INFO] 正在获取最新版本...")
	version, err := getLatestVersion()
	if err != nil {
		fmt.Printf("[ERROR] 获取最新版本失败: %s\n", err.Error())
		os.Exit(1)
	}
	fmt.Printf("[INFO] 最新版本: %s\n", version)

	// 2. 下载 skills zip
	assetURL := fmt.Sprintf("%s/v%s/uupt-skills.zip", releaseBaseURL, version)
	fmt.Printf("[INFO] 正在下载 skills (%s)...\n", version)

	tmpDir, err := os.MkdirTemp("", "uupt-skills-*")
	if err != nil {
		fmt.Printf("[ERROR] 创建临时目录失败: %s\n", err.Error())
		os.Exit(1)
	}
	defer os.RemoveAll(tmpDir)

	zipPath := filepath.Join(tmpDir, "uupt-skills.zip")
	if err := downloadFile(assetURL, zipPath); err != nil {
		fmt.Printf("[ERROR] 下载 skills 失败: %s\n", err.Error())
		fmt.Println("[提示] 请检查网络连接或稍后重试")
		os.Exit(1)
	}

	// 3. 解压
	extractDir := filepath.Join(tmpDir, "extracted")
	if err := extractZip(zipPath, extractDir); err != nil {
		fmt.Printf("[ERROR] 解压失败: %s\n", err.Error())
		os.Exit(1)
	}

	// 4. 定位 skill 源目录
	skillSrc := extractDir
	nestedSkillMD := filepath.Join(extractDir, skillName, "SKILL.md")
	if _, err := os.Stat(nestedSkillMD); err == nil {
		skillSrc = filepath.Join(extractDir, skillName)
	}
	skillMD := filepath.Join(skillSrc, "SKILL.md")
	if _, err := os.Stat(skillMD); os.IsNotExist(err) {
		fmt.Println("[ERROR] Release 中未找到 SKILL.md 文件")
		os.Exit(1)
	}

	// 5. 安装到各智能体目录
	fmt.Printf("[INFO] 安装到根目录: %s\n", root)
	installed := installSkillsToRoot(skillSrc, root)

	if installed == 0 {
		fmt.Println("[OK] Skill 安装完成（未检测到已安装的智能体目录，已安装到 .agents/skills/uupt）")
	} else {
		fmt.Printf("[OK] Skill 安装完成，共安装到 %d 个智能体目录\n", installed)
	}

	fmt.Println()
	fmt.Println("  📖 Skill 包含:")
	fmt.Println("     • SKILL.md — 主 skill 文件（产品概览与意图路由）")
	fmt.Println("     • references/ — 详细的产品命令参考文档")
	fmt.Println()
	fmt.Println("  ⚡ 前提条件: uupt-open-cli 已安装并在 PATH 中")
	fmt.Println()
}

func runSkillUninstall(cmd *cobra.Command, args []string) {
	root := skillRoot
	if root == "" {
		root, _ = os.Getwd()
	}

	removed := 0
	for _, agentDir := range agentDirs {
		skillPath := filepath.Join(root, agentDir, skillName)
		if _, err := os.Stat(skillPath); err == nil {
			if err := os.RemoveAll(skillPath); err != nil {
				fmt.Printf("[WARN] 删除失败: %s (%s)\n", skillPath, err.Error())
			} else {
				fmt.Printf("[INFO] 已移除: %s\n", skillPath)
				removed++
			}
		}
	}

	// 也检查 HOME 目录
	home, _ := os.UserHomeDir()
	if home != "" && home != root {
		for _, agentDir := range agentDirs {
			skillPath := filepath.Join(home, agentDir, skillName)
			if _, err := os.Stat(skillPath); err == nil {
				if err := os.RemoveAll(skillPath); err != nil {
					fmt.Printf("[WARN] 删除失败: %s (%s)\n", skillPath, err.Error())
				} else {
					fmt.Printf("[INFO] 已移除: %s\n", skillPath)
					removed++
				}
			}
		}
	}

	if removed == 0 {
		fmt.Println("[INFO] 未找到已安装的 uupt skill")
	} else {
		fmt.Printf("[OK] 已从 %d 个位置移除 skill\n", removed)
	}
}

// installSkillsToRoot installs skills to all detected agent directories under root.
// Returns the number of directories installed to.
func installSkillsToRoot(skillSrc, root string) int {
	installed := 0

	for idx, agentDir := range agentDirs {
		baseDir := filepath.Join(root, agentDir)
		parentGate := filepath.Dir(baseDir)

		// Parent-directory gate: skip if parent doesn't exist (except for first entry)
		if idx > 0 {
			if _, err := os.Stat(parentGate); os.IsNotExist(err) {
				continue
			}
		}

		dest := filepath.Join(baseDir, skillName)
		if err := copySkillDir(skillSrc, dest); err != nil {
			fmt.Printf("[WARN] 安装到 %s 失败: %s\n", dest, err.Error())
			continue
		}

		fileCount := countFiles(dest)
		label := dest
		if strings.HasPrefix(dest, root) {
			rel, _ := filepath.Rel(root, dest)
			label = rel
		}
		fmt.Printf("  ✅ Skills → %s (%d 个文件)\n", label, fileCount)

		// First target: list contents
		if installed == 0 {
			entries, err := os.ReadDir(dest)
			if err == nil {
				for _, entry := range entries {
					if entry.IsDir() {
						subPath := filepath.Join(dest, entry.Name())
						subCount := countFiles(subPath)
						fmt.Printf("     📁 %s/ (%d 个文件)\n", entry.Name(), subCount)
					} else {
						fmt.Printf("     📄 %s\n", entry.Name())
					}
				}
			}
		}

		installed++
	}

	// Fallback: if no agent dirs found, install to .agents/skills/uupt
	if installed == 0 {
		fallbackDest := filepath.Join(root, ".agents", "skills", skillName)
		if err := copySkillDir(skillSrc, fallbackDest); err != nil {
			fmt.Printf("[ERROR] 安装失败: %s\n", err.Error())
			os.Exit(1)
		}
		fileCount := countFiles(fallbackDest)
		fmt.Printf("  ✅ Skills → .agents/skills/%s (%d 个文件)\n", skillName, fileCount)
		installed = 1
	}

	return installed
}

// copySkillDir recursively copies a directory, replacing the destination if it exists.
func copySkillDir(src, dst string) error {
	// Remove existing
	os.RemoveAll(dst)

	// Create destination
	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}

	return filepath.WalkDir(src, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		dstPath := filepath.Join(dst, relPath)

		if d.IsDir() {
			return os.MkdirAll(dstPath, 0755)
		}

		return copyFile(path, dstPath)
	})
}

// countFiles counts all files in a directory recursively.
func countFiles(dir string) int {
	count := 0
	filepath.WalkDir(dir, func(_ string, d os.DirEntry, _ error) error {
		if d != nil && !d.IsDir() {
			count++
		}
		return nil
	})
	return count
}
