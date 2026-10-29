// Package linkmodel defines link domain data.
package linkmodel

import "time"

type LinkStatus string

const (
	LinkStatusNormal   LinkStatus = "normal"
	LinkStatusAbnormal LinkStatus = "abnormal"
)

func (LinkStatus) Values() []string {
	return []string{
		string(LinkStatusNormal),
		string(LinkStatusAbnormal),
	}
}

type Link struct {
	ID uint

	Name        string
	URL         string
	Description *string
	Avatar      *string

	Enabled   bool
	Status    LinkStatus
	SortOrder int

	CategoryID *uint

	CreatedAt time.Time
	UpdatedAt time.Time
}
