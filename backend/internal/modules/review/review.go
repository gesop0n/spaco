// Package reviewは、復習対象の登録、今日の復習、再挑戦の記録を扱うmoduleである。
package review

import (
	"errors"
	"net/http"

	"connectrpc.com/connect"

	"github.com/gesop0n/spaco/backend/generated/spaco/review/v1/reviewv1connect"
)

// Moduleは、組み立て済みreview moduleの公開窓口である。
// use caseやrepositoryの具象型はmodule外へ公開しない。
type Module struct {
	handler reviewv1connect.ReviewServiceHandler
}

func NewModule(handler reviewv1connect.ReviewServiceHandler) (*Module, error) {
	if handler == nil {
		return nil, errors.New("create review module: handler is required")
	}
	return &Module{handler: handler}, nil
}

// ConnectHandlerは、ReviewServiceをmountするpathとHTTP handlerを返す。
func (m *Module) ConnectHandler(options ...connect.HandlerOption) (string, http.Handler) {
	return reviewv1connect.NewReviewServiceHandler(m.handler, options...)
}
