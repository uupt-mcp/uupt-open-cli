# UU跑腿 CLI Skill 安装脚本 (Windows)
# Install UU跑腿 agent skills from GitHub Releases into agent skill directories.
# Usage:
#   irm https://raw.githubusercontent.com/uupt-mcp/uupt-open-cli/main/scripts/install-skills.ps1 | iex
#
# Downloads uupt-skills.zip from GitHub Releases and copies it under each target
# agent skill directory. Set UUPT_SKILLS_ROOT environment variable to change the
# base path (default: current directory). Set UUPT_SKILLS_ROOT=$HOME to match
# home-directory layout.
#
# Environment variables (optional):
#   UUPT_SKILLS_VERSION  — release tag (default: latest)
#   UUPT_SKILLS_ROOT     — base path for agent dirs (default: $PWD)

$ErrorActionPreference = "Stop"
try {
    [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12
} catch {}

$Repo = "uupt-mcp/uupt-open-cli"
$Version = if ($env:UUPT_SKILLS_VERSION) { $env:UUPT_SKILLS_VERSION } else { "latest" }
$SkillName = "uupt"
$Root = if ($env:UUPT_SKILLS_ROOT) { $env:UUPT_SKILLS_ROOT } else { $PWD.Path }

# ── Helpers ──────────────────────────────────────────────────────────────────

function Write-Info { param([string]$Message) Write-Host "  $Message" -ForegroundColor Green }
function Write-Warn { param([string]$Message) Write-Host "  $Message" -ForegroundColor Yellow }
function Write-Err  { param([string]$Message) Write-Host "  ❌ $Message" -ForegroundColor Red; exit 1 }

function Resolve-SkillVersion {
    $resolvedVersion = $Version

    try {
        $response = Invoke-WebRequest -Uri "https://github.com/$Repo/releases/latest" -Method Head -MaximumRedirection 0 -ErrorAction SilentlyContinue -UseBasicParsing
        $location = $response.Headers.Location
        if (-not $location) {
            $location = $response.Headers['Location']
        }
        if ($location) {
            $tag = $location -replace '.*\/tag\/', ''
            if ($tag) { $resolvedVersion = $tag }
        }
    } catch {
        # 尝试从异常中提取 Location 头
        if ($_.Exception.Response -and $_.Exception.Response.Headers) {
            $location = $_.Exception.Response.Headers['Location']
            if ($location) {
                $tag = $location -replace '.*\/tag\/', ''
                if ($tag) { $resolvedVersion = $tag }
            }
        }
    }

    # 回退到 latest 文件
    if ($resolvedVersion -eq "latest") {
        try {
            $ver = (Invoke-WebRequest -Uri "https://cdn.jsdelivr.net/gh/uupt-mcp/uupt-open-cli@main/latest" -UseBasicParsing).Content.Trim()
            if (-not [string]::IsNullOrWhiteSpace($ver)) { $resolvedVersion = $ver }
        } catch {}
    }

    if ($resolvedVersion -eq "latest") {
        Write-Err "无法确定最新版本，请设置 UUPT_SKILLS_VERSION 环境变量"
    }

    # 确保版本号以 'v' 开头
    if (-not $resolvedVersion.StartsWith('v')) {
        $resolvedVersion = "v$resolvedVersion"
    }

    return $resolvedVersion
}

function Copy-SkillFiles {
    param(
        [string]$Source,
        [string]$Destination,
        [string]$Label,
        [bool]$Detailed = $false
    )

    if (Test-Path $Destination) {
        Remove-Item -Path $Destination -Recurse -Force
    }

    New-Item -ItemType Directory -Path $Destination -Force | Out-Null
    Copy-Item -Path "$Source\*" -Destination $Destination -Recurse -Force

    $fileCount = (Get-ChildItem -Path $Destination -Recurse -File).Count
    Write-Info "✅ Skills → $Label ($fileCount 个文件)"

    if ($Detailed) {
        Get-ChildItem -Path $Destination | ForEach-Object {
            if ($_.PSIsContainer) {
                $subCount = (Get-ChildItem -Path $_.FullName -Recurse -File).Count
                Write-Host "     📁 $($_.Name)/ ($subCount 个文件)"
            } else {
                Write-Host "     📄 $($_.Name)"
            }
        }
    }
}

# Install skills into all detected agent directories under a given root.
function Install-SkillsToRoot {
    param(
        [string]$SkillSource,
        [string]$RootPath
    )

    $agentDirs = @(
        ".agents\skills",
        ".claude\skills",
        ".cursor\skills",
        ".gemini\skills",
        ".codex\skills",
        ".github\skills",
        ".windsurf\skills",
        ".augment\skills",
        ".cline\skills",
        ".amp\skills",
        ".kiro\skills",
        ".trae\skills",
        ".openclaw\skills",
        ".hermes\skills",
        ".qoder\skills"
    )

    $installed = 0
    $idx = 0

    foreach ($agentDir in $agentDirs) {
        $baseDir = Join-Path $RootPath $agentDir
        $parentGate = Split-Path $baseDir -Parent

        if ($idx -gt 0 -and -not (Test-Path $parentGate)) {
            $idx++
            continue
        }

        $dest = Join-Path $baseDir $SkillName
        $label = if ($RootPath -eq $HOME) { "~\$agentDir\$SkillName" } else { "$baseDir\$SkillName" }

        if ($installed -eq 0) {
            Copy-SkillFiles -Source $SkillSource -Destination $dest -Label $label -Detailed $true
        } else {
            Copy-SkillFiles -Source $SkillSource -Destination $dest -Label $label -Detailed $false
        }

        $installed++
        $idx++
    }

    if ($installed -eq 0) {
        $fallbackDest = Join-Path $RootPath ".agents\skills\$SkillName"
        $fallbackLabel = if ($RootPath -eq $HOME) { "~\.agents\skills\$SkillName" } else { $fallbackDest }
        Copy-SkillFiles -Source $SkillSource -Destination $fallbackDest -Label $fallbackLabel -Detailed $true
    }
}

# ── Main ─────────────────────────────────────────────────────────────────────

Write-Host ""
Write-Host "  ┌──────────────────────────────────────┐"
Write-Host "  │     UU跑腿 Skill 安装器              │"
Write-Host "  │     UU跑腿开放平台 CLI               │"
Write-Host "  └──────────────────────────────────────┘"
Write-Host ""

$version = Resolve-SkillVersion

$tmpDir = Join-Path $env:TEMP "uupt-skills-install-$(Get-Random)"
New-Item -ItemType Directory -Path $tmpDir -Force | Out-Null

try {
    $assetUrl = "https://github.com/$Repo/releases/download/$version/uupt-skills.zip"
    $zipPath = Join-Path $tmpDir "uupt-skills.zip"

    Write-Info "⬇  从 GitHub Releases 下载 skills: $Repo ($version)"
    $candidates = @(
        "https://ghfast.top/$assetUrl",
        $assetUrl,
        "https://ghproxy.net/$assetUrl",
        "https://mirror.ghproxy.com/$assetUrl"
    )
    $downloaded = $false
    foreach ($candidate in $candidates) {
        try {
            Write-Info "尝试 $candidate"
            Invoke-WebRequest -Uri $candidate -OutFile $zipPath -UseBasicParsing
            if ((Test-Path $zipPath) -and (Get-Item $zipPath).Length -gt 0) {
                $downloaded = $true
                break
            }
        } catch {
            Write-Warn "失败: $($_.Exception.Message)"
        }
    }
    if (-not $downloaded) {
        Write-Err "下载 skills 失败"
    }

    $extractDir = Join-Path $tmpDir "extracted"
    Expand-Archive -Path $zipPath -DestinationPath $extractDir -Force

    $skillSrc = $extractDir
    $nestedSkill = Join-Path $extractDir "$SkillName\SKILL.md"
    if (Test-Path $nestedSkill) {
        $skillSrc = Join-Path $extractDir $SkillName
    }

    $skillMd = Join-Path $skillSrc "SKILL.md"
    if (-not (Test-Path $skillMd)) {
        Write-Err "Release 资产中未找到 skill 文件"
    }

    Write-Host ""
    Write-Info "安装到根目录: $Root"
    Install-SkillsToRoot -SkillSource $skillSrc -RootPath $Root

    Write-Host ""
    Write-Host "  📖 Skill 包含:"
    Write-Host "     • SKILL.md — 主 skill 文件（产品概览与意图路由）"
    Write-Host "     • references/ — 详细的产品命令参考文档"
    Write-Host ""
    Write-Host "  ⚡ 前提条件: uupt-open-cli 已安装并在 PATH 中"
    Write-Host "     安装方式: irm https://raw.githubusercontent.com/uupt-mcp/uupt-open-cli/main/scripts/install.ps1 | iex"
    Write-Host ""
} finally {
    Remove-Item -Path $tmpDir -Recurse -Force -ErrorAction SilentlyContinue
}
