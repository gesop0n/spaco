package app

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

// newRouterは、ConnectRPC Handlerと通常のHTTP endpointをEchoへ登録する。
// Echo固有のroutingやmiddleware設定は、この組み立て層から外へ漏らさない。
func newRouter(
	connectPath string,
	connectHandler http.Handler,
	allowedOrigins []string,
) *echo.Echo {
	router := echo.New()
	router.Use(middleware.Recover())
	router.Use(corsMiddleware(allowedOrigins))

	// Cloud Runでは一部のz終端pathが予約されるため、/healthを使用する。
	router.GET("/health", func(context *echo.Context) error {
		return context.String(http.StatusOK, "ok\n")
	})

	// Connectが返すpathは末尾が`/`なので、Echoのwildcardと結合して
	// `/ServiceName/MethodName`配下を生成済みHandlerへそのまま渡す。
	connectRoute := strings.TrimSuffix(connectPath, "/") + "/*"
	router.Any(connectRoute, echo.WrapHandler(connectHandler))

	return router
}
