// Package domainは、review moduleの復習対象・再挑戦の記録に関するルールを定義する。
package domain

import "errors"

var (
	ErrInvalidReviewItem  = errors.New("invalid review item")
	ErrInvalidReviewLog   = errors.New("invalid review log")
	ErrInvalidID          = errors.New("invalid review id")
	ErrInvalidNote        = errors.New("invalid note")
	ErrInvalidOutcome     = errors.New("invalid review outcome")
	ErrInvalidPerformedAt = errors.New("invalid performed at")
	ErrReviewItemPaused   = errors.New("review item is paused")
)
