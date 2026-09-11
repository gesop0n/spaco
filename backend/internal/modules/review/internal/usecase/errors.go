package usecase

import "errors"

var (
	ErrReviewItemNotFound = errors.New("review item not found")
	ErrProblemNotFound    = errors.New("problem not found")
	ErrNoProblems         = errors.New("at least one problem is required")
	ErrTooManyProblems    = errors.New("too many problems")
	ErrRequestIDConflict  = errors.New("request id is already used for another review item")
	// ErrProblemInfoMissingは、登録済みの復習対象に対応する問題情報が見つからない不整合を表す。
	ErrProblemInfoMissing = errors.New("problem information is missing")
)
