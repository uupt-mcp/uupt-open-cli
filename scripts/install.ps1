# UU跑腿 CLI Windows 安装脚本
$ErrorActionPreference = "Stop"

# 旧版 Windows / PowerShell 5 默认 TLS 1.0，访问 GitHub 会报
# “未能创建 SSL/TLS 安全通道”。必须先启用 TLS 1.2，下载优先走 curl.exe。
try {
    [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12
} catch {}

# 核心变量
$BaseUrl = "https://github.com/uupt-mcp/uupt-open-cli/releases/download"
$LatestVersionUrl = "https://raw.githubusercontent.com/uupt-mcp/uupt-open-cli/refs/heads/main/latest"
$InstallDir = "$env:USERPROFILE\.uupt-open-cli"
$DownloadDir = "$InstallDir\downloads"
$BinaryName = "uupt-open-cli.exe"

function Write-Info { param([string]$Message) Write-Host "[INFO] $Message" -ForegroundColor Green }
function Write-Warn { param([string]$Message) Write-Host "[WARN] $Message" -ForegroundColor Yellow }
function Write-Err { param([string]$Message) Write-Host "[ERROR] $Message" -ForegroundColor Red; exit 1 }

function Get-CurlExe {
    $candidates = @(
        "$env:SystemRoot\System32\curl.exe",
        "$env:SystemRoot\SysWOW64\curl.exe"
    )
    foreach ($c in $candidates) {
        if (Test-Path $c) { return $c }
    }
    $cmd = Get-Command curl.exe -ErrorAction SilentlyContinue
    if ($cmd) { return $cmd.Source }
    return $null
}

function Get-DownloadUrls {
    param([string]$Url)
    $urls = New-Object System.Collections.Generic.List[string]
    if ($Url -match '^https://raw\.githubusercontent\.com/([^/]+)/([^/]+)/(.+)$') {
        $owner = $Matches[1]
        $repo = $Matches[2]
        $rest = $Matches[3] -replace '^refs/heads/', ''
        $slash = $rest.IndexOf('/')
        if ($slash -gt 0) {
            $ref = $rest.Substring(0, $slash)
            $path = $rest.Substring($slash + 1)
            [void]$urls.Add("https://cdn.jsdelivr.net/gh/$owner/${repo}@${ref}/$path")
        }
        [void]$urls.Add("https://ghproxy.net/$Url")
        [void]$urls.Add($Url)
        return $urls
    }

    [void]$urls.Add($Url)
    if ($Url -like "https://github.com/*") {
        [void]$urls.Add("https://ghproxy.net/$Url")
        [void]$urls.Add("https://mirror.ghproxy.com/$Url")
    }
    return $urls
}

function Invoke-FileDownload {
    param(
        [Parameter(Mandatory = $true)][string]$Url,
        [Parameter(Mandatory = $true)][string]$OutFile
    )

    $dir = Split-Path -Parent $OutFile
    if ($dir -and -not (Test-Path $dir)) {
        New-Item -ItemType Directory -Path $dir -Force | Out-Null
    }

    $curl = Get-CurlExe
    $lastError = $null
    foreach ($candidate in (Get-DownloadUrls $Url)) {
        try {
            if (Test-Path $OutFile) { Remove-Item $OutFile -Force -ErrorAction SilentlyContinue }
            if ($curl) {
                Write-Info "下载 $candidate"
                & $curl -fsSL --tlsv1.2 --connect-timeout 10 --max-time 60 --retry 1 -o $OutFile $candidate
                if ($LASTEXITCODE -ne 0) { throw "curl 退出码 $LASTEXITCODE" }
            } else {
                Write-Info "下载 $candidate (Invoke-WebRequest)"
                Invoke-WebRequest -Uri $candidate -OutFile $OutFile -UseBasicParsing
            }
            if ((Test-Path $OutFile) -and (Get-Item $OutFile).Length -gt 0) {
                return
            }
            throw "下载文件为空"
        } catch {
            $lastError = $_
            Write-Warn "失败: $($_.Exception.Message)"
        }
    }
    Write-Err "下载失败: $Url`n$lastError"
}

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
    $tmp = Join-Path $env:TEMP "uupt-cli-latest.txt"
    try {
        Invoke-FileDownload -Url $LatestVersionUrl -OutFile $tmp
        $version = (Get-Content $tmp -Raw -Encoding UTF8).Trim()
        if ([string]::IsNullOrWhiteSpace($version)) {
            Write-Err "无法获取最新版本号"
        }
        Write-Info "最新版本: v$version"
        return $version
    } finally {
        if (Test-Path $tmp) { Remove-Item $tmp -Force -ErrorAction SilentlyContinue }
    }
}

# 架构检测
function Get-TargetArch {
    $arch = [System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture
    if ($arch -eq 'Arm64') {
        # 当前 goreleaser 不发布 windows/arm64，提示用户
        Write-Err "Windows ARM64 暂不支持，请使用 amd64 版本"
    }
    return "amd64"
}

# 下载
function Download-Release {
    param([string]$Version, [string]$Arch)

    $filename = "uupt-open-cli-$Version-windows-$Arch.zip"
    $url = "$BaseUrl/v$Version/$filename"
    $dest = "$DownloadDir\$filename"

    if (-not (Test-Path $DownloadDir)) {
        New-Item -ItemType Directory -Path $DownloadDir -Force | Out-Null
    }

    Write-Info "下载 $filename..."
    Invoke-FileDownload -Url $url -OutFile $dest

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
    $arch = Get-TargetArch
    Write-Info "平台: windows/$arch"

    $archive = Download-Release -Version $version -Arch $arch
    Install-Release -ArchivePath $archive -Version $version
    Configure-Path
    Verify-Install

    Write-Host ""
    Write-Info "安装完成！运行 'uupt-open-cli --help' 开始使用"
    Write-Host ""
}

# 执行
Main -RequestedVersion $args[0]
