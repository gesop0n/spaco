package app

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRouterMountsConnectHandlerWithoutChangingPath(t *testing.T) {
	t.Parallel()

	const (
		servicePath   = "/spaco.account.v1.AccountService/"
		procedurePath = servicePath + "GetCurrentAccount"
	)
	connectHandler := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != procedurePath {
			t.Fatalf("request path = %q, want %q", request.URL.Path, procedurePath)
		}
		writer.WriteHeader(http.StatusAccepted)
	})
	router := newRouter(servicePath, connectHandler, []string{"http://localhost:5173"})
	request := httptest.NewRequest(http.MethodPost, procedurePath, nil)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusAccepted {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusAccepted)
	}
}

func TestRouterServesHealthCheck(t *testing.T) {
	t.Parallel()

	router := newRouter("/example.Service/", http.NotFoundHandler(), nil)
	request := httptest.NewRequest(http.MethodGet, "/health", nil)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	if response.Body.String() != "ok\n" {
		t.Fatalf("body = %q, want %q", response.Body.String(), "ok\\n")
	}
}
