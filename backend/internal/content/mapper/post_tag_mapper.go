package mapper

import (
	"blog-server/ent"
	contentmodel "blog-server/internal/content/model"
)

// ToPostTag converts an ent.PostTag to contentmodel.PostTag
func ToPostTag(t *ent.PostTag) contentmodel.PostTag {
	if t == nil {
		return contentmodel.PostTag{}
	}
	return contentmodel.PostTag{
		ID:        t.ID,
		Name:      t.Name,
		Slug:      t.Slug,
		CreatedAt: t.CreatedAt,
		UpdatedAt: t.UpdatedAt,
	}
}

// ToPostTags converts ent.PostTags to contentmodel.PostTags.
func ToPostTags(tags []*ent.PostTag) []contentmodel.PostTag {
	if len(tags) == 0 {
		return nil
	}

	res := make([]contentmodel.PostTag, len(tags))

	for i, t := range tags {
		res[i] = ToPostTag(t)
	}

	return res
}
