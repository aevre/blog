package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// AccountSession holds the schema definition for the AccountSession entity.
type AccountSession struct {
	ent.Schema
}

// Fields of the AccountSession.
func (AccountSession) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Immutable(),
		field.Uint("user_id").
			Immutable(),
		field.Time("expires_at"),
		field.Time("revoked_at").
			Optional().
			Nillable(),
	}
}

// Edges of the AccountSession.
func (AccountSession) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("sessions").
			Field("user_id").
			Unique().
			Required().
			Immutable(),

		edge.To("refresh_tokens", AccountRefreshToken.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}

func (AccountSession) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("expires_at", "id").
			StorageKey("account_sessions_expires_at"),
	}
}

func (AccountSession) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table: "account_sessions",
		},
	}
}
