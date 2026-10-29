package schema

import (
	contentmodel "blog-server/internal/content/model"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Post holds the schema definition for the Post entity.
type Post struct {
	ent.Schema
}

func (Post) Mixin() []ent.Mixin {
	return []ent.Mixin{
		TimeMixin{},
	}
}

// Fields of the Post.
func (Post) Fields() []ent.Field {
	return []ent.Field{
		field.Uint("user_id"),

		field.String("title").
			MaxLen(255),

		field.String("summary").
			MaxLen(255).
			Optional().
			Nillable(),

		field.Text("content"),

		field.String("cover").
			MaxLen(255).
			Optional().
			Nillable(),

		field.Uint("read_time_minutes"),

		field.Uint("view_count").
			Default(0),

		field.Enum("status").
			GoType(contentmodel.PostStatus("")).
			Default(string(contentmodel.PostStatusDraft)),

		field.Time("published_at").
			Optional().
			Nillable(),
	}
}

// Edges of the Post.
func (Post) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("author", User.Type).
			Ref("posts").
			Field("user_id").
			Unique().
			Required(),

		edge.To("categories", PostCategory.Type).
			Through("post_category_relations", PostCategoryRelation.Type),

		edge.To("tags", PostTag.Type).
			Through("post_tag_relations", PostTagRelation.Type),
	}
}
