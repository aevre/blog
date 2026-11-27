package handler

import (
	"time"

	linkmodel "blog-server/internal/link/model"
)

// LinkRes is the response body for link list endpoints.
// Includes all fields the frontend needs to display and manage links.
type LinkRes struct {
	ID          uint                 `json:"id"`
	Name        string               `json:"name"`
	URL         string               `json:"url"`
	Description string               `json:"description"`
	Avatar      string               `json:"avatar"`
	Status      linkmodel.LinkStatus `json:"status"`
	SortOrder   int                  `json:"sortOrder"`
}

type ManagedLinkRes struct {
	ID              uint                       `json:"id"`
	SubmitterID     uint                       `json:"submitterID"`
	Name            string                     `json:"name"`
	URL             string                     `json:"url"`
	Description     *string                    `json:"description"`
	Avatar          *string                    `json:"avatar"`
	Status          linkmodel.LinkStatus       `json:"status"`
	ReviewStatus    linkmodel.LinkReviewStatus `json:"reviewStatus"`
	ReviewedAt      *time.Time                 `json:"reviewedAt"`
	RejectionReason *string                    `json:"rejectionReason"`
	Version         int64                      `json:"version"`
	SortOrder       int                        `json:"sortOrder"`
	CategoryID      *uint                      `json:"categoryId"`
	CreatedAt       time.Time                  `json:"createdAt"`
	UpdatedAt       time.Time                  `json:"updatedAt"`
}

func toPublicLink(l *linkmodel.Link) LinkRes {
	out := LinkRes{
		ID:        l.ID,
		Name:      l.Name,
		URL:       l.URL,
		Status:    l.Status,
		SortOrder: l.SortOrder,
	}
	if l.Description != nil {
		out.Description = *l.Description
	}
	if l.Avatar != nil {
		out.Avatar = *l.Avatar
	}
	return out
}

func toPublicLinks(links []*linkmodel.Link) []LinkRes {
	out := make([]LinkRes, len(links))
	for i, l := range links {
		out[i] = toPublicLink(l)
	}
	return out
}

func toManagedLink(l *linkmodel.Link) ManagedLinkRes {
	return ManagedLinkRes{
		ID:              l.ID,
		SubmitterID:     l.SubmitterID,
		Name:            l.Name,
		URL:             l.URL,
		Description:     l.Description,
		Avatar:          l.Avatar,
		Status:          l.Status,
		ReviewStatus:    l.ReviewStatus,
		ReviewedAt:      l.ReviewedAt,
		RejectionReason: l.RejectionReason,
		Version:         l.Version,
		SortOrder:       l.SortOrder,
		CategoryID:      l.CategoryID,
		CreatedAt:       l.CreatedAt,
		UpdatedAt:       l.UpdatedAt,
	}
}

func toManagedLinks(links []*linkmodel.Link) []ManagedLinkRes {
	out := make([]ManagedLinkRes, len(links))
	for i, l := range links {
		out[i] = toManagedLink(l)
	}
	return out
}
