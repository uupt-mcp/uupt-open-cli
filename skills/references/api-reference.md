# UU跑腿开放平台 API 参考

## 签名与协议

### 请求方式
- HTTP POST
- Content-Type: application/json
- Header: X-App-Id

### 签名算法
1. 将业务参数序列化为紧凑JSON（中文原字符，无多余空格）
2. 拼接签名字符串：`bizJson + appSecret + timestamp`
3. 计算MD5并转为大写：`MD5(signStr).toUpperCase()`

### 请求体结构

**需要授权的接口**（业务接口）：
```json
{
  "openId": "用户授权ID",
  "timestamp": 1700000000,
  "biz": "{\"fromAddress\":\"...\",\"toAddress\":\"...\"}",
  "sign": "MD5签名大写"
}
```

**不需要授权的接口**（注册接口）：
```json
{
  "timestamp": 1700000000,
  "biz": "{\"userMobile\":\"...\",\"userIp\":\"...\"}",
  "sign": "MD5签名大写"
}
```

### API基础URL
- 生产：`https://api-open.uupt.com/openapi/v3/`
- 测试：`http://api-open.test.uupt.com/openapi/v3/`

---

## register 注册/授权

### 步骤一：发送短信验证码

**路径**：`user/unauthorized/sendSmsCode`

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| userMobile | string | 是 | 手机号 |
| userIp | string | 是 | 用户公网IP（CLI自动检测） |
| imageCode | string | 否 | 图片验证码（触发后需要） |

**响应**：
- 成功：code=0，输出 [SMS_SENT]
- 需图片验证码：code=88100106，输出 [IMAGE_CAPTCHA_REQUIRED] + IMAGE_DATA

### 步骤二：完成授权

**路径**：`user/unauthorized/auth`

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| userMobile | string | 是 | 手机号 |
| userIp | string | 是 | 用户公网IP |
| smsCode | string | 是 | 短信验证码 |
| cityName | string | 是 | 固定"郑州市" |
| countyName | string | 是 | 固定空字符串 |

**响应**：
- 成功：body.openId（自动保存到config.json）
- 失败：错误码和错误信息

---

## orderPrice 订单询价

**路径**：`order/orderPrice`

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| fromAddress | string | 是 | 起点地址（越详细越准确） |
| toAddress | string | 是 | 终点地址 |
| sendType | string | 是 | 固定"SEND" |
| cityName | string | 是 | 城市名（默认"郑州市"） |
| specialChannel | int | 是 | 固定5 |

**响应关键字段**：

| 字段 | 说明 |
|------|------|
| body.needPayMoney | 需支付金额（单位：分） |
| body.totalMoney | 总金额（单位：分） |
| body.distance | 距离（单位：米） |
| body.priceToken | 价格令牌（创建订单时使用） |

---

## addOrder 创建订单

**路径**：`order/addOrder`

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| priceToken | string | 是 | 询价返回的priceToken |
| receiver_phone | string | 是 | 收件人手机号 |
| pushType | string | 是 | 固定"OPEN_ORDER" |
| payType | string | 是 | 固定"BALANCE_PAY" |
| specialChannel | int | 是 | 5(默认) 或 6(微信) |
| specialType | string | 是 | 固定"NOT_NEED_WARM" |

**specialChannel 映射**：
| 渠道 | specialChannel值 |
|------|-----------------|
| 默认(余额支付) | 5 |
| wechat(微信支付) | 6 |

**响应关键字段**：

| 字段 | 说明 |
|------|------|
| body.orderCode | 订单号 |
| body.orderUrl | 支付链接（余额不足时） |

---

## orderDetail 查询订单详情

**路径**：`order/orderDetail`

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| order_code | string | 是 | 订单号 |

**响应关键字段**：

| 字段 | 说明 |
|------|------|
| body.orderCode | 订单号 |
| body.state | 订单状态码 |
| body.orderPrice | 订单价格（分） |
| body.fromAddress | 起点地址 |
| body.toAddress | 终点地址 |
| body.driverName | 骑手姓名 |
| body.driverMobile | 骑手电话 |

**订单状态映射**：

| 状态码 | 描述 |
|--------|------|
| 1 | 下单成功 |
| 3 | 骑手已接单 |
| 4 | 骑手已到达 |
| 5 | 骑手已取件 |
| 6 | 骑手送达中 |
| 10 | 已完成 |
| 11 | 已取消 |
| 20 | 异常订单 |

---

## cancelOrder 取消订单

**路径**：`order/cancelOrder`

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| order_code | string | 是 | 订单号 |
| reason | string | 否 | 取消原因 |

**响应**：code=0 表示取消成功

---

## driverTrack 跑男追踪

**路径**：`order/driverTrack`

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| order_code | string | 是 | 订单号 |

**响应关键字段**：

| 字段 | 说明 |
|------|------|
| body.driver_name | 骑手姓名 |
| body.driver_phone | 骑手电话 |
| body.longitude | 经度 |
| body.latitude | 纬度 |
| body.distance | 到终点距离（米） |

---

## 常用枚举

### specialChannel 值

| 值 | 含义 |
|----|------|
| 5 | 默认（余额支付） |
| 6 | 微信支付 |

### 退出码

| 退出码 | 含义 |
|--------|------|
| 0 | 成功 |
| 1 | 一般错误 |
| 2 | 需要图片验证码 |

### 输出标记

| 标记 | 含义 |
|------|------|
| [REGISTRATION_REQUIRED] | 需要先注册获取OpenId |
| [SMS_SENT] | 短信验证码发送成功 |
| [IMAGE_CAPTCHA_REQUIRED] | 需要图片验证码（附带 IMAGE_DATA=base64） |
| [REGISTRATION_SUCCESS] | 注册成功（附带 openId） |
| [REGISTRATION_FAILED] | 注册失败 |
| [PAYMENT_REQUIRED] | 余额不足需支付（附带 ORDER_CODE=、PAYMENT_URL=、QRCODE_FILE=） |
| [UPDATE_AVAILABLE] | 有新版本可更新（附带 JSON: {"current":"...","latest":"..."}） |
| [FATAL] | 致命配置错误 |
| [INFO] | 一般信息提示 |
| [OK] | 操作成功 |
| [ERROR] | 操作失败 |
| [WARN] | 警告信息 |
