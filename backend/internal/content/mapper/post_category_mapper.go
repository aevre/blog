package mapper

import (
	"blog-server/ent"
	contentmodel "blog-server/internal/content/model"
)

// ToPostCategory converts an ent.PostCategory to contentmodel.PostCategory.
func ToPostCategory(c *ent.PostCategory) contentmodel.PostCategory {
	if c == nil {
		return contentmodel.PostCategory{}
	}
	return contentmodel.PostCategory{
		ID:        c.ID,
		Name:      c.Name,
		Slug:      c.Slug,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
}

// ToPostCategories converts ent.PostCategories to contentmodel.PostCategories.
func ToPostCategories(cats []*ent.PostCategory) []contentmodel.PostCategory {
	if len(cats) == 0 {
		return nil
	}

	res := make([]contentmodel.PostCategory, len(cats))
	for i, c := range cats {
		res[i] = ToPostCategory(c)
	}

	return res
}
