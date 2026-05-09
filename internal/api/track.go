package api

import (
	"uupt-open-cli/internal/config"
)

// DriverTrack 跑男实时追踪
func DriverTrack(cfg *config.Config, orderCode string) (map[string]interface{}, error) {
	bizParams := map[string]interface{}{
		"order_code": orderCode,
	}

	return AuthorizedRequest(cfg, "order/driverTrack", bizParams)
}
