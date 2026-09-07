package app

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
)

func TestCORSMiddlewareAllowsConfiguredPreflight(t *testing.T) {
	t.Parallel()

	router := echo.New()
	router.Use(corsMiddleware([]string{"http://localhost:5173"}))
	router.Any("/rpc", func(*echo.Context) error {
		t.Fatal("preflightでnext handlerが呼ばれた")
		return nil
	})
	request := httptest.NewRequest(http.MethodOptions, "/rpc", nil)
	request.Header.Set("Origin", "http://localhost:5173")
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusNoContent)
	}
	if got := response.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:5173" {
		t.Fatalf("Access-Control-Allow-Origin = %q", got)
	}
	if got := response.Header().Get("Access-Control-Allow-Headers"); got == "" {
		t.Fatal("Access-Control-Allow-Headers is empty")
	}
}

func TestCORSMiddlewareRejectsUnknownOrigin(t *testing.T) {
	t.Parallel()

	router := echo.New()
	router.Use(corsMiddleware([]string{"http://localhost:5173"}))
	router.Any("/rpc", func(*echo.Context) error {
		t.Fatal("未許可Originでnext handlerが呼ばれた")
		return nil
	})
	request := httptest.NewRequest(http.MethodPost, "/rpc", nil)
	request.Header.Set("Origin", "https://malicious.example")
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusForbidden)
	}
}

func TestCORSMiddlewareAllowsRequestWithoutOrigin(t *testing.T) {
	t.Parallel()

	router := echo.New()
	router.Use(corsMiddleware(nil))
	router.Any("/rpc", func(context *echo.Context) error {
		return context.NoContent(http.StatusCreated)
	})
	request := httptest.NewRequest(http.MethodPost, "/rpc", nil)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusCreated)
	}
}
