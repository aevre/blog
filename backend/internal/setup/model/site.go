// Package setupmodel defines site setup domain data.
package setupmodel

import "time"

type Site struct {
	Name        string
	Logo        *string
	Greeting    *string
	Description *string
	CreatedAt   *time.Time
}
