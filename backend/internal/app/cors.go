package app

import (
	"net/http"
	"strings"
)

var corsAllowedHeaders = strings.Join([]string{
	"Authorization",
	"Connect-Protocol-Version",
	"Connect-Timeout-Ms",
	"Content-Type",
	"Grpc-Timeout",
	"X-Grpc-Web",
	"X-User-Agent",
}, ", ")

// corsMiddlewareは、設定したfrontend originからのConnectRPC requestだけを許可する。
// Bearer token方式なのでcookie用のcredentials headerは付与しない。
func corsMiddleware(allowedOrigins []string, next http.Handler) http.Handler {
	allowed := make(map[string]struct{}, len(allowedOrigins))
	for _, origin := range allowedOrigins {
		allowed[strings.TrimSpace(origin)] = struct{}{}
	}

	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		origin := request.Header.Get("Origin")
		if origin == "" {
			next.ServeHTTP(writer, request)
			return
		}
		if _, ok := allowed[origin]; !ok {
			http.Error(writer, "origin is not allowed", http.StatusForbidden)
			return
		}

		header := writer.Header()
		header.Set("Access-Control-Allow-Origin", origin)
		header.Add("Vary", "Origin")
		header.Set("Access-Control-Expose-Headers", "Grpc-Status, Grpc-Message, Grpc-Status-Details-Bin")

		if request.Method == http.MethodOptions {
			header.Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			header.Set("Access-Control-Allow-Headers", corsAllowedHeaders)
			header.Set("Access-Control-Max-Age", "7200")
			writer.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(writer, request)
	})
}
