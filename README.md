# uupt-open-cli

UU跑腿开放平台 CLI 工具 —— 为 AI 智能体提供同城即时配送能力。

## 功能

- **register** — 手机号注册/获取授权
- **auth** — 授权管理（login / status / logout，供 WorkBuddy 连接器调用）
- **price** — 订单询价
- **create** — 创建订单
- **detail** — 查询订单详情
- **cancel** — 取消订单
- **track** — 跑男实时追踪
- **skill** — 管理 Agent Skill（安装/卸载）
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

# 授权状态（WorkBuddy 连接器用）
uupt-open-cli auth status
uupt-open-cli auth login
uupt-open-cli auth logout

# 跑腿配送询价
uupt-open-cli price --from-address="郑州市金水区花园路" --to-address="郑州市二七区大学路"

# 帮忙服务询价
uupt-open-cli price --from-address="郑州市金水区花园路" --order-type="help"

# 创建跑腿配送订单
uupt-open-cli create --price-token="xxx" --receiver-phone="13800138000"

# 创建帮忙服务订单
uupt-open-cli create --price-token="xxx" --receiver-phone="13800138000" --note="帮我搬一箱矿泉水到3楼"

# 查询订单
uupt-open-cli detail --order-code="UU123456789"

# 取消订单
uupt-open-cli cancel --order-code="UU123456789" --reason="不需要了"

# 跑男追踪
uupt-open-cli track --order-code="UU123456789"
```

### 订单类型

支持两种订单类型：

| 类型 | 参数值 | 说明 |
|------|--------|------|
| 跑腿配送 | `send`（默认） | 物品从 A 地送到 B 地 |
| 帮忙服务 | `help` | 跑男在指定地点提供现场协助（起始地址和终点地址相同） |

帮忙订单需要额外传递 `--note` 参数描述具体的帮忙内容。

## AI 智能体集成

uupt-open-cli 为 AI 智能体提供完整的 Agent Skill，安装后 Claude Code / Cursor / Windsurf / Codex 等智能体可通过自然语言直接操作 UU跑腿服务。

### 安装 Skill

**方式一：CLI 命令安装（推荐）**

```bash
# 安装到当前项目目录
uupt-open-cli skill install

# 安装到 HOME 目录（所有智能体全局可用）
uupt-open-cli skill install --root ~

# 卸载
uupt-open-cli skill uninstall
```

**方式二：独立脚本安装**

**macOS / Linux：**

```bash
curl -fsSL https://raw.githubusercontent.com/uupt-mcp/uupt-open-cli/main/scripts/install-skills.sh | sh
```

**Windows (PowerShell)：**

```powershell
irm https://raw.githubusercontent.com/uupt-mcp/uupt-open-cli/main/scripts/install-skills.ps1 | iex
```

> `install.sh` 安装 CLI 到 `$HOME/.uupt-open-cli/`（全局）；`install-skills.sh` 安装 Skill 到 `./.agents/skills/uupt`（当前项目），自动检测已安装的智能体目录。

**指定安装目录：**

```bash
# 安装到 HOME 目录（所有智能体全局可用）
UUPT_SKILLS_ROOT=$HOME curl -fsSL https://raw.githubusercontent.com/uupt-mcp/uupt-open-cli/main/scripts/install-skills.sh | sh

# 指定版本
UUPT_SKILLS_VERSION=v1.0.0 curl -fsSL https://raw.githubusercontent.com/uupt-mcp/uupt-open-cli/main/scripts/install-skills.sh | sh
```

### Skill 内容

| 组件 | 路径 | 说明 |
|------|------|------|
| 主 Skill | `SKILL.md` | 意图路由、决策树、安全规则、输出标记处理 |
| CLI 使用说明 | `references/cli-usage.md` | 命令格式、参数规范、交互流程 |
| 接口参考 | `references/api-reference.md` | 各命令参数详细说明与示例 |

### 支持的智能体

Skill 自动安装到以下已检测到的智能体目录：

`.agents` · `.claude` · `.cursor` · `.gemini` · `.codex` · `.github` · `.windsurf` · `.augment` · `.cline` · `.amp` · `.kiro` · `.trae` · `.openclaw` · `.hermes` · `.qoder`

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
