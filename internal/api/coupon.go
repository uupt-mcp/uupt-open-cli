package api

import (
	"uupt-open-cli/internal/config"
)

// ReceiveCouponPackages 领取优惠券包
// source: 领取来源（决定可领哪些券包）
func ReceiveCouponPackages(cfg *config.Config, source int) (map[string]interface{}, error) {
	bizParams := map[string]interface{}{
		"source": source,
	}

	return AuthorizedRequest(cfg, "/openapiext/v3/aiagentcoupon/receiveCouponPackages", bizParams)
}
