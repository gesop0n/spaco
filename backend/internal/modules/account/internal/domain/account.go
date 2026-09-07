package domain

import (
	"errors"
	"fmt"
	"strings"
	"time"
	_ "time/tzdata"

	"github.com/gesop0n/spaco/backend/internal/shared/identifier"
)

var (
	ErrInvalidAccount   = errors.New("invalid account")
	ErrInvalidAtCoderID = errors.New("invalid atcoder id")
	ErrInvalidTimeZone  = errors.New("invalid time zone")
)

const defaultTimeZone = "Asia/Tokyo"

// Accountは、アプリ内ユーザーと復習に必要なprofileを保持する集約である。
type Account struct {
	id        identifier.UserID
	atCoderID *string
	timeZone  string
}

// NewAccountは、外部identityを初めて確認したユーザーの未設定Accountを生成する。
func NewAccount(id identifier.UserID) (Account, error) {
	if id.IsZero() {
		return Account{}, fmt.Errorf("%w: user id is required", ErrInvalidAccount)
	}

	return Account{id: id, timeZone: defaultTimeZone}, nil
}

// RehydrateAccountは、repositoryから取得した値をAccountへ復元する。
func RehydrateAccount(
	id identifier.UserID,
	atCoderID *string,
	timeZone string,
) (Account, error) {
	account, err := NewAccount(id)
	if err != nil {
		return Account{}, err
	}

	parsedTimeZone, err := NewTimeZone(timeZone)
	if err != nil {
		return Account{}, err
	}
	account.timeZone = parsedTimeZone.String()

	if atCoderID != nil {
		parsedAtCoderID, parseErr := NewAtCoderID(*atCoderID)
		if parseErr != nil {
			return Account{}, parseErr
		}
		value := parsedAtCoderID.String()
		account.atCoderID = &value
	}

	return account, nil
}

func (a Account) ID() identifier.UserID { return a.id }

func (a Account) AtCoderID() (string, bool) {
	if a.atCoderID == nil {
		return "", false
	}
	return *a.atCoderID, true
}

func (a Account) TimeZone() string { return a.timeZone }

// SetupCompletedは、利用開始に必要なprofileが設定済みかを返す。
func (a Account) SetupCompleted() bool {
	return a.atCoderID != nil && a.timeZone != ""
}

// Profileは、ユーザーが変更できるAccount属性をまとめた値である。
type Profile struct {
	atCoderID AtCoderID
	timeZone  TimeZone
}

func NewProfile(atCoderID, timeZone string) (Profile, error) {
	parsedAtCoderID, err := NewAtCoderID(atCoderID)
	if err != nil {
		return Profile{}, err
	}
	parsedTimeZone, err := NewTimeZone(timeZone)
	if err != nil {
		return Profile{}, err
	}

	return Profile{atCoderID: parsedAtCoderID, timeZone: parsedTimeZone}, nil
}

func (p Profile) AtCoderID() string { return p.atCoderID.String() }
func (p Profile) TimeZone() string  { return p.timeZone.String() }

type AtCoderID struct{ value string }

func NewAtCoderID(value string) (AtCoderID, error) {
	value = strings.TrimSpace(value)
	if value == "" || len(value) > 64 || strings.ContainsAny(value, " \t\r\n") {
		return AtCoderID{}, fmt.Errorf("%w: must be 1-64 characters without spaces", ErrInvalidAtCoderID)
	}

	return AtCoderID{value: value}, nil
}

func (id AtCoderID) String() string { return id.value }

type TimeZone struct{ value string }

func NewTimeZone(value string) (TimeZone, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return TimeZone{}, fmt.Errorf("%w: value is required", ErrInvalidTimeZone)
	}
	if _, err := time.LoadLocation(value); err != nil {
		return TimeZone{}, fmt.Errorf("%w: %q: %v", ErrInvalidTimeZone, value, err)
	}

	return TimeZone{value: value}, nil
}

func (zone TimeZone) String() string { return zone.value }
