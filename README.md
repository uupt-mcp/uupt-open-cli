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

## 安装

### macOS / Linux

```bash
curl -fsSL https://raw.githubusercontent.com/uupt-mcp/uupt-open-cli/main/scripts/install.sh | bash
```

### Windows

```powershell
irm https://raw.githubusercontent.com/uupt-mcp/uupt-open-cli/main/scripts/install.ps1 | iex
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
