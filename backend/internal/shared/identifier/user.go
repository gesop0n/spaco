package identifier

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
)

var ErrInvalidUserID = errors.New("invalid user id")

type UserID struct {
	value uuid.UUID
}

func NewUserID() UserID {
	return UserID{value: uuid.Must(uuid.NewV7())}
}

func ParseUserID(text string) (UserID, error) {
	value, err := uuid.Parse(text)
	if err != nil {
		return UserID{}, fmt.Errorf("parse user id %q: %w", text, err)
	}

	return UserID{value: value}, nil
}

func (id UserID) String() string  { return id.value.String() }
func (id UserID) UUID() uuid.UUID { return id.value }
func (id UserID) IsZero() bool    { return id.value == uuid.Nil }
