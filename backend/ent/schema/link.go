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

		field.Bool("enabled").
			Default(false),

		field.String("name").
			MaxLen(100),

		field.Int("sort_order").
			Default(0),

		field.String("url").
			MaxLen(255).
			Unique(),

		field.String("avatar").
			MaxLen(255).
			Optional(),

		field.Enum("status").
			GoType(linkmodel.LinkStatus("")).
			Default(string(linkmodel.LinkStatusNormal)),

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
	}
}
