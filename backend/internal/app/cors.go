package app

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

var corsAllowedHeaders = []string{
	"Authorization",
	"Connect-Protocol-Version",
	"Connect-Timeout-Ms",
	"Content-Type",
	"Grpc-Timeout",
	"X-Grpc-Web",
	"X-User-Agent",
}

var corsExposedHeaders = []string{
	"Grpc-Status",
	"Grpc-Message",
	"Grpc-Status-Details-Bin",
}

// corsMiddlewareは、設定したfrontend originからのConnectRPC requestだけを許可する。
// Bearer token方式なのでcookie用のcredentials headerは付与しない。
func corsMiddleware(allowedOrigins []string) echo.MiddlewareFunc {
	allowed := make(map[string]struct{}, len(allowedOrigins))
	for _, origin := range allowedOrigins {
		allowed[strings.TrimSpace(origin)] = struct{}{}
	}

	return middleware.CORSWithConfig(middleware.CORSConfig{
		// Echo標準では未許可Originの通常requestを処理するため、既存の
		// allowlist境界を維持するcustom validatorを設定する。
		UnsafeAllowOriginFunc: func(
			_ *echo.Context,
			origin string,
		) (string, bool, error) {
			if _, ok := allowed[origin]; ok {
				return origin, true, nil
			}
			return "", false, echo.NewHTTPError(
				http.StatusForbidden,
				"origin is not allowed",
			)
		},
		AllowMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodOptions,
		},
		AllowHeaders:  corsAllowedHeaders,
		ExposeHeaders: corsExposedHeaders,
		MaxAge:        7200,
	})
}
