package dto

import "errors"

var (
	ErrInvalidUserID    = errors.New("invalid user_id: must be a valid UUID")
	ErrInvalidStartDate = errors.New("invalid start_date: must be in format MM-YYYY")
	ErrInvalidEndDate   = errors.New("invalid end_date: must be in format MM-YYYY")
)
