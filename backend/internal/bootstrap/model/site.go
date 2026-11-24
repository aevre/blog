// Package bootstrapmodel defines site setup domain data.
package bootstrapmodel

import "time"

type Site struct {
	Name        string
	Logo        *string
	Greeting    *string
	Description *string
	CreatedAt   *time.Time
}
