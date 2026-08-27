---
name: uupt-open-cli
description: 当 AI 智能体需要调用 UU跑腿同城配送服务（注册、询价、下单、查询、取消、跑男追踪、领取优惠券）时使用本 skill；用户说「领券」「领优惠券」「有优惠券吗」「有什么优惠」「有什么活动」「参加活动」时也应触发
---

# UU跑腿开放平台 CLI

## 适用智能体

本 skill 面向通用 AI 智能体应用使用，包括但不限于 Codex CLI、Claude Code、Openclaw 以及其他支持读取 Markdown 指令和执行本地命令的 Agent。使用时不要依赖特定厂商能力；优先按本文件中的固定二进制路径、配置位置、命令格式和参考文档执行。

## 核心流程

1. 接收用户配送需求
2. 检查注册状态（无 openId 时引导注册）
3. 调用对应 CLI 命令
4. 解析输出（识别标记和 JSON 数据）
5. 以自然语言回复用户

发起接口调用前，先阅读 [CLI 使用说明](references/cli-usage.md)。构造参数时，阅读 [接口参考](references/api-reference.md)。

## 已支持命令

- `register` — 手机号注册/获取授权
- `auth login` — 浏览器/本地页授权登录（WorkBuddy 连接器用）
- `auth status` — 检查授权状态
- `auth logout` — 取消授权
- `price` — 订单询价（支持跑腿配送和帮忙服务）
- `create` — 创建订单（支持跑腿配送和帮忙服务）
- `detail` — 查询订单详情
- `cancel` — 取消订单
- `track` — 跑男实时追踪
- `coupon` — 领取优惠券（每日可领，符合条件时附主题活动入口）
- `skill install` — 安装 Agent Skill 到智能体目录
- `skill uninstall` — 卸载 Agent Skill
- `update` — 检查并更新到最新版本
- `uninstall` — 卸载 uupt-open-cli

## 调用方式

```bash
$HOME/.uupt-open-cli/uupt-open-cli <command> [flags]
```

命令表：

| 命令 | 示例 |
|------|------|
| register | `$HOME/.uupt-open-cli/uupt-open-cli register --mobile="13800138000"` |
| register(带验证码) | `$HOME/.uupt-open-cli/uupt-open-cli register --mobile="13800138000" --sms-code="123456"` |
| register(带图片验证码) | `$HOME/.uupt-open-cli/uupt-open-cli register --mobile="13800138000" --image-code="1234"` |
| auth login | `$HOME/.uupt-open-cli/uupt-open-cli auth login` |
| auth status | `$HOME/.uupt-open-cli/uupt-open-cli auth status` |
| auth logout | `$HOME/.uupt-open-cli/uupt-open-cli auth logout` |
| price(跑腿配送) | `$HOME/.uupt-open-cli/uupt-open-cli price --from-address="郑州市金水区" --to-address="郑州市二七区"` |
| price(帮忙服务) | `$HOME/.uupt-open-cli/uupt-open-cli price --from-address="郑州市金水区" --order-type="help"` |
| create(跑腿配送) | `$HOME/.uupt-open-cli/uupt-open-cli create --price-token="xxx" --receiver-phone="13800138000"` |
| create(帮忙服务) | `$HOME/.uupt-open-cli/uupt-open-cli create --price-token="xxx" --receiver-phone="13800138000" --note="帮忙内容"` |
| create(微信) | `$HOME/.uupt-open-cli/uupt-open-cli create --price-token="xxx" --receiver-phone="13800138000" --channel="wechat"` |
| detail | `$HOME/.uupt-open-cli/uupt-open-cli detail --order-code="UU123456789"` |
| cancel | `$HOME/.uupt-open-cli/uupt-open-cli cancel --order-code="UU123456789" --reason="不需要了"` |
| track | `$HOME/.uupt-open-cli/uupt-open-cli track --order-code="UU123456789"` |
| coupon | `$HOME/.uupt-open-cli/uupt-open-cli coupon` |
| update | `$HOME/.uupt-open-cli/uupt-open-cli update` |
| uninstall | `$HOME/.uupt-open-cli/uupt-open-cli uninstall` |
| uninstall(跳过确认) | `$HOME/.uupt-open-cli/uupt-open-cli uninstall --force` |

## 输出标记处理

| 标记 | 含义 | 处理方式 |
|------|------|---------|
| `[REGISTRATION_REQUIRED]` | 需要注册 | 引导用户提供手机号 |
| `[SMS_SENT]` | 验证码已发送 | 提示用户查看短信输入验证码 |
| `[IMAGE_CAPTCHA_REQUIRED]` | 需要图片验证码 | 从 IMAGE_DATA= 提取 base64 图片展示给用户 |
| `[REGISTRATION_SUCCESS]` | 注册成功 | 告知成功，继续执行原始请求 |
| `[REGISTRATION_FAILED]` | 注册失败 | 从发送验证码步骤重试（最多3次） |
| `[PAYMENT_REQUIRED]` | 余额不足需支付 | 提取 ORDER_CODE=、PAYMENT_URL=、QRCODE_FILE=，按渠道展示支付信息（微信发二维码图片，其他渠道发链接），并引导支付确认流程 |
| `[COUPON_RESULT]` | 领券结果结构化标记 | 提取 NEWLY_CLAIMED=、COUPON_COUNT=，符合条件时还有 THURSDAY_JOIN_ABLE=true、THURSDAY_QRCODE_URL=、THURSDAY_QRCODE_FILE=，按场景六话术模板回复 |
| `[UPDATE_AVAILABLE]` | 有新版本可更新 | 提示用户运行 update 命令或重新运行安装脚本 |
| `[INFO]` | 信息提示 | 展示给用户参考 |
| `[OK]` | 操作成功 | 告知用户操作完成 |
| `[ERROR]` | 操作失败 | 展示错误信息 |
| `[WARN]` | 警告信息 | 展示警告，操作可能未完全成功 |

## 场景流程

### 场景零：注册引导

- **触发**：执行任何命令时输出包含 `[REGISTRATION_REQUIRED]`
- **流程**：询问手机号 → register → [SMS_SENT] → 等待验证码 → register(带sms-code) → [REGISTRATION_SUCCESS]
- **图片验证码**：如果出现 [IMAGE_CAPTCHA_REQUIRED]，从 IMAGE_DATA 提取 base64 展示，等用户识别后 register(带image-code)

### 场景一：询价

- 判断订单类型（配送 vs 帮忙）
- 配送订单：提供起点和终点地址
- 帮忙订单：只提供帮忙地点，`--order-type="help"`，`--to-address` 自动使用 `--from-address` 的值
- 执行 price 命令
- 返回价格（分转元）和距离

### 场景二：下单

- 需要先询价获得 priceToken
- 帮忙订单必须传递 `--note` 参数描述具体帮忙内容
- 执行 create 命令
- 成功返回 orderCode
- 如果 [PAYMENT_REQUIRED]，按渠道展示支付信息

#### 余额不足时的支付处理

当 create 命令输出包含 `[PAYMENT_REQUIRED]` 时，表示账户余额不足，需要用户完成支付。脚本会输出：
- `ORDER_CODE={order_code}` — 订单编号
- `PAYMENT_URL={payment_url}` — 支付链接（用户点击后可选择微信或支付宝支付）
- `QRCODE_FILE={qrcode_path}` — 支付二维码图片本地路径（**仅 `--channel="wechat"` 时才有此输出**）

##### 根据渠道调用 create 命令

- **微信渠道**：必须传递 `--channel="wechat"` 参数以生成二维码图片
  ```bash
  $HOME/.uupt-open-cli/uupt-open-cli create --price-token="xxx" --receiver-phone="13800138000" --channel="wechat"
  ```
- **其他渠道**（飞书/钉钉/企业微信/QQ/Telegram/其他）：无需传递 `--channel` 参数
  ```bash
  $HOME/.uupt-open-cli/uupt-open-cli create --price-token="xxx" --receiver-phone="13800138000"
  ```

##### 根据渠道展示支付信息

⚠️ **微信渠道特殊处理**：微信中链接无法直接打开，必须发送二维码图片附件！

**微信渠道：**
```
message(action=send, channel="wechat", path="{QRCODE_FILE}", message="请扫码支付 {price/100} 元")
```

回复模板：
```
账户余额不足，需要完成支付

订单编号：{order_code}
配送费用：{price/100} 元

请扫码支付，支付完成后告诉我。

（附件：支付二维码）
```

**其他渠道**（飞书/钉钉/企业微信/QQ/Telegram/其他）：直接发送支付链接

回复模板：
```
账户余额不足，需要完成支付

订单编号：{order_code}
配送费用：{price/100} 元

💳 请点击以下链接完成支付（支持微信/支付宝）：
{PAYMENT_URL}

支付完成后请告诉我。
```

##### 支付确认流程

1. 展示支付信息后，等待用户返回
2. 询问用户是否已完成支付：
   ```
   您好，请问是否已完成支付？
   - 是，已支付完成
   - 否，还未支付
   ```
3. 用户确认支付完成后，立即调用 detail 命令查询订单状态：
   ```bash
   $HOME/.uupt-open-cli/uupt-open-cli detail --order-code="{order_code}"
   ```
4. 展示订单详情：
   ```
   支付成功！订单详情如下：

   订单编号：{order_code}
   订单状态：{status}
   起点：{from_address}
   终点：{to_address}
   配送费：{price/100} 元

   骑手正在接单中，请保持电话畅通。
   ```

### 场景三：查询订单

- 执行 detail 命令
- 展示订单状态、骑手信息

### 场景四：取消订单

- 执行 cancel 命令
- 确认取消结果

### 场景五：骑手追踪

- 执行 track 命令
- 展示骑手位置和预计到达信息

### 场景六：领取优惠券

- **触发**：用户要领券或询问优惠（「领券」「领优惠券」「有优惠券吗」「有什么优惠」「有什么活动」「参加活动」），直接执行，无需额外信息；输出 `[REGISTRATION_REQUIRED]` 时先走场景零注册后重试
- **命令**：`$HOME/.uupt-open-cli/uupt-open-cli coupon`（`--source` 领取来源，默认1，一般无需传入）
- **返回字段**（JSON `body`）：

| 字段 | 说明 |
|------|------|
| `newlyClaimed` | 是否本次新领取：true-本次新领取；false-今天已领过（返回当日记录） |
| `couponList` | 领取的优惠券列表，每项含 `packageName`（券包名称，可为空）、`couponDetail`（优惠券信息）、`expireDate`（过期时间 yyyy-MM-dd） |
| `thursdayJoinAble` | 是否可参与淡定星期四活动 |

- **回复优先级**：`couponList` 为空或 null → 场景 C；`newlyClaimed=false` → 场景 B；其他 → 场景 A。任意场景下 `thursdayJoinAble=true` 时追加场景 D。回复严格按以下话术模板输出，不得改动任何标点、空行、换行位置，不得输出触发条件或任何 JSON 字段名

#### 场景 A：领券成功（`newlyClaimed=true` 且 `couponList` 非空）

```
🎉 一键领券完成！本次共领取 N 张优惠券

| 券名称 | 优惠券信息 | 过期时间 |
|--------|---------|--------|
| [packageName] | [couponDetail] | [expireDate] |

可以在UU跑腿App优惠券列表查看所有券详情。
```

> N = couponList 条数；表格按 couponList 逐行输出；packageName 为空时填「优惠券」。

#### 场景 B：当日已领过券（`newlyClaimed=false`）

```
您今天已经领过UU跑腿的优惠券啦，这是今日领取的优惠详情：

| 券名称 | 优惠券信息 | 过期时间 |
|--------|---------|--------|
| [packageName] | [couponDetail] | [expireDate] |

有新的优惠我第一时间通知你 🔔
```

#### 场景 C：无可领券（`couponList` 为空或 null）

```
当前UU跑腿暂无优惠券，有新券上线我第一时间通知你 🔔
```

#### 场景 D：淡定星期四活动（`thursdayJoinAble=true`，附加在上述任意场景回复之后）

用户可参与淡定星期四（下单1元起活动）。命令会输出 `THURSDAY_QRCODE_URL`（远程图片链接）和 `THURSDAY_QRCODE_FILE`（本地图片路径，随 CLI 内嵌，始终可用）。在上述场景话术之后追加以下内容，**图片必须真实展示，不能只输出 URL/路径文本**：

```
另外你还可以参与「淡定星期四」活动，下单1元起！用微信扫描下方二维码即可参与 👇
```

图片展示方式按当前平台能力选择：支持远程 Markdown 图片渲染的平台用 `![淡定星期四活动](THURSDAY_QRCODE_URL)`；不渲染远程链接但有本地图片发送机制的平台（如 `present_files`/`send_file`）直接发送 `THURSDAY_QRCODE_FILE`；两者都不可用时输出活动说明文字 + 可点击的 `THURSDAY_QRCODE_URL` 链接，引导用户微信自行打开。

### 场景七：卸载

- 执行 uninstall 命令
- 确认卸载（除非使用 --force 跳过确认）
- 自动移除 PATH 配置、删除安装目录
- Windows 下如果从安装目录运行，优先使用 PowerShell 等待进程退出后自动删除；若 PowerShell 不可用则降级使用 cmd.exe 批处理延迟删除

## 项目约定

- 二进制路径固定为 `$HOME/.uupt-open-cli/uupt-open-cli`
- 所有命令通过 CLI 执行，不直接调用 API
- 价格单位为分，展示时除以100转为元
- 地址越详细越准确（精确到门牌号最佳）
- 不存储用户敏感信息（仅 openId 保存在本地配置）
- 余额不足时返回 `[PAYMENT_REQUIRED]`，微信渠道必须用 `message(action=send, channel="wechat", path="{QRCODE_FILE}")` 发送二维码图片附件；其他渠道直接发送 `{PAYMENT_URL}` 支付链接
- 微信渠道创建订单时必须传递 `--channel="wechat"` 参数以生成支付二维码图片
- 帮忙订单询价时使用 `--order-type="help"`，起始地址和终点地址自动保持一致
- 帮忙订单创建时必须传递 `--note` 参数描述具体帮忙内容
- 领取优惠券需先完成注册；同一用户同一来源当天只能新领一次，重复领取返回当日记录（`newlyClaimed=false`）；回复严格按场景六话术模板输出；`thursdayJoinAble=true` 时必须真实展示活动二维码图片
