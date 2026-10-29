// Package mapper converts content persistence models into domain models.
package mapper

import (
	"blog-server/ent"
	contentmodel "blog-server/internal/content/model"
)

func extractAuthor(p *ent.Post) string {
	if p.Edges.Author == nil {
		return ""
	}
	return p.Edges.Author.Username
}

// ToPost converts an ent.Post to contentmodel.Post.
func ToPost(p *ent.Post) *contentmodel.Post {
	if p == nil {
		return &contentmodel.Post{}
	}

	return &contentmodel.Post{
		ID:      p.ID,
		Title:   p.Title,
		Summary: p.Summary,
		Cover:   p.Cover,

		User:   extractAuthor(p),
		UserID: p.UserID,

		Content:         p.Content,
		ReadTimeMinutes: p.ReadTimeMinutes,
		ViewCount:       p.ViewCount,

		Status: p.Status,

		PublishedAt: p.PublishedAt,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,

		Tags:       ToPostTags(p.Edges.Tags),
		Categories: ToPostCategories(p.Edges.Categories),
	}
}

func ToPosts(ps []*ent.Post) []*contentmodel.Post {
	if len(ps) == 0 {
		return nil
	}

	res := make([]*contentmodel.Post, len(ps))
	for i, p := range ps {
		res[i] = ToPost(p)
	}

	return res
}
