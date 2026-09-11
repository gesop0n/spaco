package usecase

import "errors"

var (
	ErrContestNotFound = errors.New("contest not found")
	ErrProblemNotFound = errors.New("problem not found")
	ErrEmptySnapshot   = errors.New("atcoder problems returned no contests or problems")
)
