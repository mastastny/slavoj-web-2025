package apperrors

import "errors"

var ErrEventOverlap = errors.New("event overlaps with existing one")
var ErrEventInThePast = errors.New("event start time is in the past")

// ValidationError carries a user-facing message about invalid input.
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string { return e.Message }
