# uupt-open-cli

UU跑腿开放平台 CLI 工具 —— 为 AI 智能体提供同城即时配送能力。

## 功能

- **register** — 手机号注册/获取授权
- **price** — 订单询价
- **create** — 创建订单
- **detail** — 查询订单详情
- **cancel** — 取消订单
- **track** — 跑男实时追踪
- **update** — 检查并更新到最新版本
- **uninstall** — 卸载 uupt-open-cli

## 安装

### 方式一：一键安装脚本（推荐）

#### macOS / Linux

```bash
curl -fsSL https://raw.githubusercontent.com/uupt-mcp/uupt-open-cli/main/scripts/install.sh | bash
```

#### Windows

```powershell
irm https://raw.githubusercontent.com/uupt-mcp/uupt-open-cli/main/scripts/install.ps1 | iex
```

### 方式二：手动下载安装

从 [Releases 页面](https://github.com/uupt-mcp/uupt-open-cli/releases) 下载对应平台的压缩包：

#### macOS

```bash
# Apple Silicon (M1/M2/M3)
curl -LO https://github.com/uupt-mcp/uupt-open-cli/releases/latest/download/uupt-open-cli-1.0.0-macos-arm64.tar.gz
tar -xzf uupt-open-cli-1.0.0-macos-arm64.tar.gz

# Intel
curl -LO https://github.com/uupt-mcp/uupt-open-cli/releases/latest/download/uupt-open-cli-1.0.0-macos-amd64.tar.gz
tar -xzf uupt-open-cli-1.0.0-macos-amd64.tar.gz
```

#### Linux

```bash
# ARM64
curl -LO https://github.com/uupt-mcp/uupt-open-cli/releases/latest/download/uupt-open-cli-1.0.0-linux-arm64.tar.gz
tar -xzf uupt-open-cli-1.0.0-linux-arm64.tar.gz

# AMD64
curl -LO https://github.com/uupt-mcp/uupt-open-cli/releases/latest/download/uupt-open-cli-1.0.0-linux-amd64.tar.gz
tar -xzf uupt-open-cli-1.0.0-linux-amd64.tar.gz
```

#### Windows

下载 [uupt-open-cli-1.0.0-windows-amd64.zip](https://github.com/uupt-mcp/uupt-open-cli/releases/latest/download/uupt-open-cli-1.0.0-windows-amd64.zip) 并解压。

> **注意：** 请将上述链接中的 `1.0.0` 替换为 [Releases 页面](https://github.com/uupt-mcp/uupt-open-cli/releases) 中的最新版本号。

解压后，将二进制文件移动到安装目录并配置 PATH：

**macOS / Linux：**

```bash
mkdir -p ~/.uupt-open-cli/configs ~/.uupt-open-cli/logs
mv uupt-open-cli ~/.uupt-open-cli/
chmod +x ~/.uupt-open-cli/uupt-open-cli

# 添加到 PATH（选择对应 shell 配置文件）
echo 'export PATH="$HOME/.uupt-open-cli:$PATH"' >> ~/.bashrc   # bash
echo 'export PATH="$HOME/.uupt-open-cli:$PATH"' >> ~/.zshrc   # zsh
source ~/.bashrc  # 或 source ~/.zshrc
```

**Windows（PowerShell）：**

```powershell
mkdir "$env:USERPROFILE\.uupt-open-cli\configs" -Force
mkdir "$env:USERPROFILE\.uupt-open-cli\logs" -Force
Move-Item uupt-open-cli.exe "$env:USERPROFILE\.uupt-open-cli\"
$installDir = "$env:USERPROFILE\.uupt-open-cli"
[Environment]::SetEnvironmentVariable("PATH", "$installDir;" + [Environment]::GetEnvironmentVariable("PATH", "User"), "User")
```

## 使用

```bash
# 查看版本
uupt-open-cli --version

# 列出所有命令
uupt-open-cli --list

# 注册（发送验证码）
uupt-open-cli register --mobile="13800138000"

# 注册（提交验证码）
uupt-open-cli register --mobile="13800138000" --sms-code="123456"

# 询价
uupt-open-cli price --from-address="郑州市金水区花园路" --to-address="郑州市二七区大学路"

# 创建订单
uupt-open-cli create --price-token="xxx" --receiver-phone="13800138000"

# 查询订单
uupt-open-cli detail --order-code="UU123456789"

# 取消订单
uupt-open-cli cancel --order-code="UU123456789" --reason="不需要了"

# 跑男追踪
uupt-open-cli track --order-code="UU123456789"
```

## 更新

### 方式一：CLI 自更新（推荐）

```bash
uupt-open-cli update
```

自动检测最新版本、下载并替换二进制，配置和日志不受影响。

### 方式二：重新运行安装脚本

macOS / Linux：

```bash
curl -fsSL https://raw.githubusercontent.com/uupt-mcp/uupt-open-cli/main/scripts/install.sh | bash
```

Windows：

```powershell
irm https://raw.githubusercontent.com/uupt-mcp/uupt-open-cli/main/scripts/install.ps1 | iex
```

安装脚本会自动检测已有安装，仅替换二进制文件，保留 `configs/config.json` 和 `logs/`。

## 卸载

```bash
uupt-open-cli uninstall
```

将删除 `~/.uupt-open-cli/` 目录（含二进制、配置、日志）并移除 PATH 配置。使用 `-y` 跳过确认：

```bash
uupt-open-cli uninstall -y
```

## 配置

配置文件位于 `$HOME/.uupt-open-cli/configs/`：
- `defaults.json` — 内置凭证（不可修改）
- `config.json` — 用户配置（注册后自动生成）

环境变量（优先级最高）：
- `UUPT_APP_ID`
- `UUPT_APP_SECRET`
- `UUPT_OPEN_ID`
- `UUPT_API_URL`

## 构建

```bash
./scripts/build.sh
```

## License

Proprietary
