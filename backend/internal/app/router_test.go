package app

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRouterMountsConnectHandlersWithoutChangingPath(t *testing.T) {
	t.Parallel()

	const (
		accountPath = "/spaco.account.v1.AccountService/"
		reviewPath  = "/spaco.review.v1.ReviewService/"
	)
	handlerFor := func(procedurePath string, status int) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			if request.URL.Path != procedurePath {
				t.Fatalf("request path = %q, want %q", request.URL.Path, procedurePath)
			}
			writer.WriteHeader(status)
		})
	}
	router := newRouter([]connectHandler{
		{path: accountPath, handler: handlerFor(accountPath+"GetCurrentAccount", http.StatusAccepted)},
		{path: reviewPath, handler: handlerFor(reviewPath+"GetTodayReviews", http.StatusCreated)},
	}, []string{"http://localhost:5173"})

	for procedurePath, want := range map[string]int{
		accountPath + "GetCurrentAccount": http.StatusAccepted,
		reviewPath + "GetTodayReviews":    http.StatusCreated,
	} {
		response := httptest.NewRecorder()
		router.ServeHTTP(response, httptest.NewRequest(http.MethodPost, procedurePath, nil))
		if response.Code != want {
			t.Fatalf("%s status = %d, want %d", procedurePath, response.Code, want)
		}
	}
}

func TestRouterServesHealthCheck(t *testing.T) {
	t.Parallel()

	router := newRouter([]connectHandler{{path: "/example.Service/", handler: http.NotFoundHandler()}}, nil)
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
