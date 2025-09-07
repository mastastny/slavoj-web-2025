package dtos

import "time"

type Event struct {
	Title string    `json:"title"`
	Start time.Time `json:"start"`
	End   time.Time `json:"end"`
}
