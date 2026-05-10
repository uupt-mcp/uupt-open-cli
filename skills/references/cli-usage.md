# UU跑腿 CLI 使用说明

## 智能体执行原则

- 始终使用完整的二进制路径调用命令
- 解析标准输出中的标记（如 [SMS_SENT]、[PAYMENT_REQUIRED] 等）
- JSON 输出即为 API 原始响应，直接解析使用
- 出错时检查退出码：0=成功, 1=一般错误, 2=需要图片验证码
- 不要猜测参数值，始终使用用户提供或上一步返回的数据

## 安装路径与发现

- 二进制文件：`$HOME/.uupt-open-cli/uupt-open-cli`
- Windows：`%USERPROFILE%\.uupt-open-cli\uupt-open-cli.exe`
- 配置目录：`$HOME/.uupt-open-cli/configs/`
- 日志目录：`$HOME/.uupt-open-cli/logs/`
- 版本检查：`$HOME/.uupt-open-cli/uupt-open-cli --version`

## 配置

配置文件位于 `$HOME/.uupt-open-cli/configs/`：

| 文件 | 用途 | 修改权限 |
|------|------|---------|
| defaults.json | 内置凭证 | 只读 |
| config.json | 用户配置 | 可写（注册后自动生成） |

环境变量（优先级最高）：
- `UUPT_APP_ID` — 应用ID
- `UUPT_APP_SECRET` — 应用密钥
- `UUPT_OPEN_ID` — 用户授权ID
- `UUPT_API_URL` — API基础URL
- `UUPT_LOG_LEVEL` — 日志级别（DEBUG/INFO/WARN/ERROR，默认INFO）

优先级：环境变量 > config.json > defaults.json

## 命令格式

```bash
$HOME/.uupt-open-cli/uupt-open-cli <command> [flags]
```

全局选项：
- `--version` — 显示版本号
- `--list` — 列出所有支持的命令
- `--help` — 显示帮助

## 常用调用

### 注册（发送验证码）
```bash
$HOME/.uupt-open-cli/uupt-open-cli register --mobile="13800138000"
```

### 注册（指定IP地址）
```bash
$HOME/.uupt-open-cli/uupt-open-cli register --mobile="13800138000" --ip="1.2.3.4"
```

### 注册（提交验证码）
```bash
$HOME/.uupt-open-cli/uupt-open-cli register --mobile="13800138000" --sms-code="123456"
```

### 注册（带图片验证码）
```bash
$HOME/.uupt-open-cli/uupt-open-cli register --mobile="13800138000" --image-code="5678"
```

### 订单询价
```bash
$HOME/.uupt-open-cli/uupt-open-cli price --from-address="郑州市金水区花园路1号" --to-address="郑州市二七区大学路100号" --city="郑州市"
```

### 创建订单
```bash
$HOME/.uupt-open-cli/uupt-open-cli create --price-token="TOKEN_FROM_PRICE" --receiver-phone="13800138000"
```

### 创建订单（微信支付渠道）
```bash
$HOME/.uupt-open-cli/uupt-open-cli create --price-token="TOKEN_FROM_PRICE" --receiver-phone="13800138000" --channel="wechat"
```

### 查询订单
```bash
$HOME/.uupt-open-cli/uupt-open-cli detail --order-code="UU123456789"
```

### 取消订单
```bash
$HOME/.uupt-open-cli/uupt-open-cli cancel --order-code="UU123456789" --reason="不需要了"
```

### 跑男追踪
```bash
$HOME/.uupt-open-cli/uupt-open-cli track --order-code="UU123456789"
```

### 更新版本
```bash
$HOME/.uupt-open-cli/uupt-open-cli update
```

### 卸载
```bash
$HOME/.uupt-open-cli/uupt-open-cli uninstall
```

### 卸载（跳过确认）
```bash
$HOME/.uupt-open-cli/uupt-open-cli uninstall --force
```

## 输出标记参考

| 标记 | 退出码 | 后续数据 | 说明 |
|------|--------|---------|------|
| [REGISTRATION_REQUIRED] | 1 | 引导信息 | 需要先注册 |
| [SMS_SENT] | 0 | 无 | 验证码发送成功 |
| [IMAGE_CAPTCHA_REQUIRED] | 2 | IMAGE_DATA=base64... | 需要识别图片验证码 |
| [REGISTRATION_SUCCESS] | 0 | openId值 | 注册成功 |
| [REGISTRATION_FAILED] | 1 | 错误信息 | 注册失败 |
| [PAYMENT_REQUIRED] | 0 | ORDER_CODE=, PAYMENT_URL=, QRCODE_FILE= | 余额不足需支付 |
| [UPDATE_AVAILABLE] | 0 | JSON(current,latest) | 有新版本可更新 |
| [FATAL] | 1 | 错误信息 | 致命配置错误 |
| [INFO] | 0 | 信息文本 | 一般信息提示 |
| [OK] | 0 | 成功文本 | 操作成功 |
| [ERROR] | 1 | 错误信息 | 操作失败 |
| [WARN] | 0 | 警告信息 | 警告，操作可能未完全成功 |

## 各命令参数参考

### register
| 参数 | 必填 | 说明 |
|------|------|------|
| --mobile | 是 | 手机号 |
| --sms-code | 否 | 短信验证码（提交授权时使用） |
| --image-code | 否 | 图片验证码（触发后需要） |
| --ip | 否 | 公网IP（默认自动检测） |

### price
| 参数 | 必填 | 说明 |
|------|------|------|
| --from-address | 是 | 发货地址 |
| --to-address | 是 | 收货地址 |
| --city | 否 | 城市名称（默认"郑州市"） |

### create
| 参数 | 必填 | 说明 |
|------|------|------|
| --price-token | 是 | 询价token |
| --receiver-phone | 是 | 收件人手机号 |
| --channel | 否 | 支付渠道（wechat表示微信） |

### detail
| 参数 | 必填 | 说明 |
|------|------|------|
| --order-code | 是 | 订单编号 |

### cancel
| 参数 | 必填 | 说明 |
|------|------|------|
| --order-code | 是 | 订单编号 |
| --reason | 否 | 取消原因 |

### track
| 参数 | 必填 | 说明 |
|------|------|------|
| --order-code | 是 | 订单编号 |

### uninstall
| 参数 | 必填 | 说明 |
|------|------|------|
| --force / -y | 否 | 跳过确认提示 |

## 排查要点

- 确认二进制存在：`ls $HOME/.uupt-open-cli/uupt-open-cli`
- 确认版本：`$HOME/.uupt-open-cli/uupt-open-cli --version`
- 检查配置：`cat $HOME/.uupt-open-cli/configs/config.json`
- 查看日志：`cat $HOME/.uupt-open-cli/logs/uupt-open-cli.$(date +%Y-%m-%d).log`
- 网络问题：确认能访问 `https://api-open.uupt.com`
- 注册问题：删除 config.json 中的 openId 可重新注册
