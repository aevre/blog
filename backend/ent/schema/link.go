package schema

import (
	linkmodel "blog-server/internal/link/model"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Link holds the schema definition for the Link entity.
type Link struct {
	ent.Schema
}

func (Link) Mixin() []ent.Mixin {
	return []ent.Mixin{
		TimeMixin{},
	}
}

// Fields of the Link.
func (Link) Fields() []ent.Field {
	return []ent.Field{
		field.String("description").
			MaxLen(255).
			Optional(),

		field.String("name").
			MaxLen(100),

		field.Int("sort_order").
			Default(0),

		field.String("url").
			MaxLen(255),

		field.String("avatar").
			MaxLen(255).
			Optional(),

		field.Uint("submitter_id"),

		field.Enum("status").
			GoType(linkmodel.LinkStatus("")).
			Default(string(linkmodel.LinkStatusNormal)),

		field.Enum("review_status").
			GoType(linkmodel.LinkReviewStatus("")).
			Default(string(linkmodel.LinkReviewPending)),

		field.Time("reviewed_at").
			Optional().
			Nillable(),

		field.String("rejection_reason").
			MaxLen(1000).
			Optional().
			Nillable(),

		field.Int64("version").
			Default(1),

		field.Uint("category_id").
			Optional(),
	}
}

// Edges of the Link.
func (Link) Edges() []ent.Edge {
	return []ent.Edge{
		edge.
			From("category", LinkCategory.Type).
			Ref("links").
			Unique().
			Field("category_id"),

		edge.From("submitter", User.Type).
			Ref("links").
			Field("submitter_id").
			Unique().
			Required(),
	}
}
