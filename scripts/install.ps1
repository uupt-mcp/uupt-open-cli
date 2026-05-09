# UU跑腿 CLI Windows 安装脚本
$ErrorActionPreference = "Stop"

# 核心变量
$BaseUrl = "https://github.com/uupt-mcp/uupt-open-cli/releases/download"
$LatestVersionUrl = "https://raw.githubusercontent.com/uupt-mcp/uupt-open-cli/refs/heads/main/latest"
$InstallDir = "$env:USERPROFILE\.uupt-open-cli"
$DownloadDir = "$InstallDir\downloads"
$BinaryName = "uupt-open-cli.exe"

function Write-Info { param([string]$Message) Write-Host "[INFO] $Message" -ForegroundColor Green }
function Write-Warn { param([string]$Message) Write-Host "[WARN] $Message" -ForegroundColor Yellow }
function Write-Err { param([string]$Message) Write-Host "[ERROR] $Message" -ForegroundColor Red; exit 1 }

# 参数验证
function Get-TargetVersion {
    param([string]$RequestedVersion)

    if ($RequestedVersion) {
        if ($RequestedVersion -notmatch '^\d+\.\d+\.\d+$') {
            Write-Err "无效的版本号格式: $RequestedVersion (期望格式: X.Y.Z)"
        }
        Write-Info "使用指定版本: v$RequestedVersion"
        return $RequestedVersion
    }

    Write-Info "获取最新版本..."
    try {
        $version = (Invoke-WebRequest -Uri $LatestVersionUrl -UseBasicParsing).Content.Trim()
        if ([string]::IsNullOrWhiteSpace($version)) {
            Write-Err "无法获取最新版本号"
        }
        Write-Info "最新版本: v$version"
        return $version
    } catch {
        Write-Err "无法获取最新版本: $_"
    }
}

# 下载
function Download-Release {
    param([string]$Version)

    $filename = "uupt-open-cli-$Version-windows-amd64.zip"
    $url = "$BaseUrl/v$Version/$filename"
    $dest = "$DownloadDir\$filename"

    if (-not (Test-Path $DownloadDir)) {
        New-Item -ItemType Directory -Path $DownloadDir -Force | Out-Null
    }

    Write-Info "下载 $filename..."
    try {
        Invoke-WebRequest -Uri $url -OutFile $dest -UseBasicParsing
    } catch {
        Write-Err "下载失败: $url`n$_"
    }

    Write-Info "下载完成: $dest"
    return $dest
}

# 解压安装
function Install-Release {
    param([string]$ArchivePath, [string]$Version)

    $tmpDir = Join-Path $env:TEMP "uupt-open-cli-install-$(Get-Random)"
    New-Item -ItemType Directory -Path $tmpDir -Force | Out-Null

    Write-Info "解压文件..."
    # 使用 Expand-Archive 解压 .zip
    Expand-Archive -Path $ArchivePath -DestinationPath $tmpDir -Force

    # 查找二进制文件
    $binary = Get-ChildItem -Path $tmpDir -Filter "uupt-open-cli*.exe" -Recurse | Select-Object -First 1
    if (-not $binary) {
        Write-Err "未找到二进制文件"
    }

    $isUpgrade = (Test-Path "$InstallDir\$BinaryName")

    if ($isUpgrade) {
        # 升级安装：仅替换二进制，保留 configs/config.json 和 logs/
        Write-Info "检测到已有安装，执行升级..."
        Copy-Item -Path $binary.FullName -Destination "$InstallDir\$BinaryName" -Force
    } else {
        # 全新安装
        Write-Info "执行全新安装..."
        if (-not (Test-Path $InstallDir)) {
            New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
        }
        if (-not (Test-Path "$InstallDir\configs")) {
            New-Item -ItemType Directory -Path "$InstallDir\configs" -Force | Out-Null
        }
        if (-not (Test-Path "$InstallDir\logs")) {
            New-Item -ItemType Directory -Path "$InstallDir\logs" -Force | Out-Null
        }

        Copy-Item -Path $binary.FullName -Destination "$InstallDir\$BinaryName" -Force

        # 复制 configs
        $configsDir = Join-Path $tmpDir "configs"
        if (Test-Path $configsDir) {
            Copy-Item -Path "$configsDir\*" -Destination "$InstallDir\configs\" -Force -Recurse
        }
    }

    # 清理临时目录
    Remove-Item -Path $tmpDir -Recurse -Force

    Write-Info "安装到: $InstallDir\$BinaryName"
}

# 配置 PATH
function Configure-Path {
    $currentPath = [Environment]::GetEnvironmentVariable("PATH", "User")

    if ($currentPath -like "*$InstallDir*") {
        Write-Info "PATH 已包含安装目录"
        return
    }

    $newPath = "$InstallDir;$currentPath"
    [Environment]::SetEnvironmentVariable("PATH", $newPath, "User")
    $env:PATH = "$InstallDir;$env:PATH"
    Write-Info "已将 $InstallDir 添加到用户 PATH"
    Write-Warn "新的终端窗口将自动生效"
}

# 验证安装
function Verify-Install {
    try {
        $ver = & "$InstallDir\$BinaryName" --version 2>&1
        Write-Info "安装验证成功: $ver"
    } catch {
        Write-Warn "安装完成，但版本验证未通过（请重新打开终端）"
    }
}

# 主流程
function Main {
    param([string]$RequestedVersion)

    Write-Host ""
    Write-Host "========================================="
    Write-Host "  UU跑腿 CLI Windows 安装程序"
    Write-Host "========================================="
    Write-Host ""

    $version = Get-TargetVersion -RequestedVersion $RequestedVersion
    Write-Info "平台: windows/amd64"

    $archive = Download-Release -Version $version
    Install-Release -ArchivePath $archive -Version $version
    Configure-Path
    Verify-Install

    Write-Host ""
    Write-Info "安装完成！运行 'uupt-open-cli --help' 开始使用"
    Write-Host ""
}

# 执行
Main -RequestedVersion $args[0]
