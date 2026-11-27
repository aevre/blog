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

type LinkReviewStatus string

const (
	LinkReviewPending  LinkReviewStatus = "pending"
	LinkReviewApproved LinkReviewStatus = "approved"
	LinkReviewRejected LinkReviewStatus = "rejected"
)

func (LinkReviewStatus) Values() []string {
	return []string{
		string(LinkReviewPending),
		string(LinkReviewApproved),
		string(LinkReviewRejected),
	}
}

type Link struct {
	ID          uint
	SubmitterID uint

	Name        string
	URL         string
	Description *string
	Avatar      *string

	Status          LinkStatus
	ReviewStatus    LinkReviewStatus
	ReviewedAt      *time.Time
	RejectionReason *string
	Version         int64

	SortOrder int

	CategoryID *uint

	CreatedAt time.Time
	UpdatedAt time.Time
}
