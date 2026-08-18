# UU跑腿 CLI Windows 安装脚本
$ErrorActionPreference = "Stop"

# 旧版 Windows / PowerShell 5 默认 TLS 1.0，访问 GitHub 会报
# “未能创建 SSL/TLS 安全通道”。必须先启用 TLS 1.2，下载优先走 curl.exe。
try {
    [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12
} catch {}

# 核心变量（二进制优先走 UU OSS，避免 GitHub 超时）
$OssBaseUrl = "https://otherfiles.uupt.com/open-cli"
$GithubReleaseBase = "https://github.com/uupt-mcp/uupt-open-cli/releases/download"
$BaseUrl = $OssBaseUrl
$LatestVersionUrl = "$OssBaseUrl/latest"
$InstallDir = "$env:USERPROFILE\.uupt-open-cli"
# 下载先落到临时目录。不要在安装成功前创建 ~/.uupt-open-cli，
# 否则空目录会被当成“已安装”，WorkBuddy 会跳过 init 直接 auth。
$DownloadDir = Join-Path $env:TEMP "uupt-open-cli-downloads"
$BinaryName = "uupt-open-cli.exe"
$MinReleaseZipBytes = 5000000

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

    if ($Url.StartsWith($OssBaseUrl)) {
        [void]$urls.Add($Url)
        $rest = $Url.Substring($OssBaseUrl.Length).TrimStart('/')
        if ($rest -eq "latest") {
            [void]$urls.Add("https://cdn.jsdelivr.net/gh/uupt-mcp/uupt-open-cli@main/latest")
            [void]$urls.Add("https://ghproxy.net/https://raw.githubusercontent.com/uupt-mcp/uupt-open-cli/refs/heads/main/latest")
            return $urls
        }
        if ($rest -match '^v(\d+\.\d+\.\d+)/(.+)$') {
            $gh = "$GithubReleaseBase/v$($Matches[1])/$($Matches[2])"
            [void]$urls.Add("https://ghfast.top/$gh")
            [void]$urls.Add($gh)
            [void]$urls.Add("https://ghproxy.net/$gh")
            return $urls
        }
        if ($rest -match '^install\.(ps1|sh)$') {
            [void]$urls.Add("https://cdn.jsdelivr.net/gh/uupt-mcp/uupt-open-cli@main/scripts/$rest")
            return $urls
        }
        return $urls
    }

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

    if ($Url -like "https://github.com/*") {
        # 国内实测 ghfast.top 明显快于 ghproxy；慢镜像放到后面并用低速中止。
        [void]$urls.Add("https://ghfast.top/$Url")
        [void]$urls.Add($Url)
        [void]$urls.Add("https://ghproxy.net/$Url")
        [void]$urls.Add("https://mirror.ghproxy.com/$Url")
        return $urls
    }

    [void]$urls.Add($Url)
    return $urls
}

function Invoke-FileDownload {
    param(
        [Parameter(Mandatory = $true)][string]$Url,
        [Parameter(Mandatory = $true)][string]$OutFile,
        [int]$MaxTimeSeconds = 45,
        [int]$MinBytes = 1
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
                # --speed-limit/--speed-time：镜像过慢时尽快换源，避免占满 WorkBuddy 的 init 超时。
                & $curl -fsSL --tlsv1.2 --connect-timeout 10 --max-time $MaxTimeSeconds `
                    --retry 1 --speed-limit 20480 --speed-time 20 `
                    -o $OutFile $candidate
                if ($LASTEXITCODE -ne 0) { throw "curl 退出码 $LASTEXITCODE" }
            } else {
                Write-Info "下载 $candidate (Invoke-WebRequest)"
                Invoke-WebRequest -Uri $candidate -OutFile $OutFile -UseBasicParsing
            }
            if ((Test-Path $OutFile) -and (Get-Item $OutFile).Length -ge $MinBytes) {
                return
            }
            throw "下载文件过小或不存在"
        } catch {
            $lastError = $_
            Write-Warn "失败: $($_.Exception.Message)"
        }
    }
    Write-Err "下载失败: $Url`n$lastError"
}

function Test-ZipArchive {
    param([string]$Path)
    if (-not (Test-Path $Path)) { return $false }
    if ((Get-Item $Path).Length -lt $MinReleaseZipBytes) { return $false }
    try {
        Add-Type -AssemblyName System.IO.Compression.FileSystem
        $zip = [System.IO.Compression.ZipFile]::OpenRead((Resolve-Path $Path))
        try {
            $hasExe = $false
            foreach ($entry in $zip.Entries) {
                if ($entry.Name -like "uupt-open-cli*.exe") { $hasExe = $true; break }
            }
            return $hasExe
        } finally {
            $zip.Dispose()
        }
    } catch {
        return $false
    }
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

function Get-InstalledVersion {
    $exe = Join-Path $InstallDir $BinaryName
    if (-not (Test-Path $exe)) { return $null }
    try {
        $out = & $exe --version 2>&1 | Out-String
        if ($out -match '(\d+\.\d+\.\d+)') { return $Matches[1] }
    } catch {}
    return $null
}

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

    if (Test-ZipArchive $dest) {
        Write-Info "使用已下载文件: $dest"
        return $dest
    }
    if (Test-Path $dest) {
        Write-Warn "已下载文件不完整或损坏，重新下载"
        Remove-Item $dest -Force -ErrorAction SilentlyContinue
    }

    Write-Info "下载 $filename..."
    Invoke-FileDownload -Url $url -OutFile $dest -MaxTimeSeconds 180 -MinBytes $MinReleaseZipBytes
    if (-not (Test-ZipArchive $dest)) {
        Write-Err "下载的安装包无效: $dest"
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
    $exe = Join-Path $InstallDir $BinaryName
    if (-not (Test-Path $exe)) {
        Write-Err "未找到 CLI 可执行文件: $exe"
    }
    try {
        $ver = & $exe --version 2>&1 | Out-String
        if ($ver -notmatch '\d+\.\d+\.\d+') {
            Write-Err "无法运行 $exe --version：$ver"
        }
        Write-Info "安装验证成功: $($ver.Trim())"
    } catch {
        Write-Err "无法运行 $exe：$($_.Exception.Message)"
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

    $installed = Get-InstalledVersion
    if ($installed) {
        Write-Info "检测到已安装: v$installed"
        try {
            if ([version]$installed -ge [version]$version) {
                Write-Info "已是目标版本，跳过下载"
                Configure-Path
                Verify-Install
                Write-Host ""
                Write-Info "安装完成！运行 'uupt-open-cli --help' 开始使用"
                Write-Host ""
                return
            }
        } catch {}
    }

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
