package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

// System holds the schema definition for the System entity.
type System struct {
	ent.Schema
}

// Fields of the System.
func (System) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").
			Default(1).
			Immutable(),
		field.Bool("initialized").
			Default(false),
		field.String("bootstrap_token_hash").
			Optional().
			Nillable(),
		field.Time("initialized_at").
			Optional(),
		field.Time("created_at").
			Default(time.Now),
		field.Time("updated_at").
			Default(time.Now),
	}
}

// Edges of the System.
func (System) Edges() []ent.Edge {
	return nil
}
