// Package mapper converts link persistence models into domain models.
package mapper

import (
	"blog-server/ent"
	linkmodel "blog-server/internal/link/model"
)

// ToLink converts an ent.Link to linkmodel.Link.
func ToLink(l *ent.Link) *linkmodel.Link {
	if l == nil {
		return &linkmodel.Link{}
	}
	link := &linkmodel.Link{
		ID:          l.ID,
		Name:        l.Name,
		URL:         l.URL,
		Enabled:     l.Enabled,
		Description: &l.Description,
		Avatar:      &l.Avatar,
		Status:      l.Status,
		SortOrder:   l.SortOrder,
		CreatedAt:   l.CreatedAt,
		UpdatedAt:   l.UpdatedAt,
	}

	// CategoryID is 0 when not set in the database (ent uses plain uint).
	// Convert to *uint only when non-zero to preserve nil semantics.
	if l.CategoryID != 0 {
		id := l.CategoryID
		link.CategoryID = &id
	}

	return link
}

func ToLinks(ls []*ent.Link) []*linkmodel.Link {
	if len(ls) == 0 {
		return nil
	}

	res := make([]*linkmodel.Link, len(ls))
	for i, l := range ls {
		res[i] = ToLink(l)
	}

	return res
}
