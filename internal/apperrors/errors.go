package apperrors

import "errors"

var ErrEventOverlap = errors.New("event overlaps with existing one")
var ErrEventInThePast = errors.New("event start time is in the past")
