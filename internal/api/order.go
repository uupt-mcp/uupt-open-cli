package api

import (
	"uupt-open-cli/internal/config"
)

// OrderPrice 订单询价
func OrderPrice(cfg *config.Config, fromAddress string, toAddress string, city string, orderType string) (map[string]interface{}, error) {
	isHelp := orderType == "help"

	sendType := "SEND"
	if isHelp {
		sendType = "HELP"
	}

	bizParams := map[string]interface{}{
		"fromAddress":    fromAddress,
		"toAddress":      toAddress,
		"sendType":       sendType,
		"cityName":       city,
		"specialChannel": 5,
	}

	if isHelp {
		bizParams["goodsType"] = "ALLHELP"
	}

	return AuthorizedRequest(cfg, "order/orderPrice", bizParams)
}

// CreateOrder 创建订单
func CreateOrder(cfg *config.Config, priceToken string, receiverPhone string, channel string, note string) (map[string]interface{}, error) {
	specialChannel := 5
	if channel == "wechat" {
		specialChannel = 6
	}
	bizParams := map[string]interface{}{
		"priceToken":     priceToken,
		"receiver_phone": receiverPhone,
		"pushType":       "OPEN_ORDER",
		"payType":        "BALANCE_PAY",
		"specialChannel": specialChannel,
		"specialType":    "NOT_NEED_WARM",
	}

	if note != "" {
		bizParams["note"] = note
	}

	return AuthorizedRequest(cfg, "order/addOrder", bizParams)
}

// OrderDetail 查询订单详情
func OrderDetail(cfg *config.Config, orderCode string) (map[string]interface{}, error) {
	bizParams := map[string]interface{}{
		"order_code": orderCode,
	}

	return AuthorizedRequest(cfg, "order/orderDetail", bizParams)
}

// CancelOrder 取消订单
func CancelOrder(cfg *config.Config, orderCode string, reason string) (map[string]interface{}, error) {
	bizParams := map[string]interface{}{
		"order_code": orderCode,
	}
	if reason != "" {
		bizParams["reason"] = reason
	}

	return AuthorizedRequest(cfg, "order/cancelOrder", bizParams)
}
