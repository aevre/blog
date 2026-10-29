package contentmodel

import "time"

type PostTag struct {
	ID uint

	Name string
	Slug string

	CreatedAt time.Time
	UpdatedAt time.Time
}
