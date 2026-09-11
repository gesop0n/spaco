package domain

import (
	"fmt"

	"github.com/google/uuid"
)

// ReviewItemIDは、復習対象の識別子である。
type ReviewItemID struct{ value uuid.UUID }

func NewReviewItemID() ReviewItemID { return ReviewItemID{value: uuid.Must(uuid.NewV7())} }

func ParseReviewItemID(value string) (ReviewItemID, error) {
	parsed, err := parseID("review item id", value)
	return ReviewItemID{value: parsed}, err
}

func (id ReviewItemID) String() string  { return id.value.String() }
func (id ReviewItemID) UUID() uuid.UUID { return id.value }
func (id ReviewItemID) IsZero() bool    { return id.value == uuid.Nil }

// ReviewLogIDは、再挑戦の記録の識別子である。
type ReviewLogID struct{ value uuid.UUID }

func NewReviewLogID() ReviewLogID { return ReviewLogID{value: uuid.Must(uuid.NewV7())} }

func ParseReviewLogID(value string) (ReviewLogID, error) {
	parsed, err := parseID("review log id", value)
	return ReviewLogID{value: parsed}, err
}

func (id ReviewLogID) String() string  { return id.value.String() }
func (id ReviewLogID) UUID() uuid.UUID { return id.value }
func (id ReviewLogID) IsZero() bool    { return id.value == uuid.Nil }

// RequestIDは、保存ボタンの二重送信を判定するため、入力画面ごとにクライアントが生成するIDである。
type RequestID struct{ value uuid.UUID }

func ParseRequestID(value string) (RequestID, error) {
	parsed, err := parseID("request id", value)
	return RequestID{value: parsed}, err
}

func (id RequestID) String() string  { return id.value.String() }
func (id RequestID) UUID() uuid.UUID { return id.value }
func (id RequestID) IsZero() bool    { return id.value == uuid.Nil }

func parseID(name, value string) (uuid.UUID, error) {
	parsed, err := uuid.Parse(value)
	if err != nil || parsed == uuid.Nil {
		return uuid.Nil, fmt.Errorf("%w: %s must be a non-nil UUID", ErrInvalidID, name)
	}
	return parsed, nil
}
