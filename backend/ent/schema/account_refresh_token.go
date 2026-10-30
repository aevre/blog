package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// AccountRefreshToken holds the schema definition for the AccountRefreshToken entity.
type AccountRefreshToken struct {
	ent.Schema
}

// Fields of the AccountRefreshToken.
func (AccountRefreshToken) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").
			StorageKey("hash").
			SchemaType(map[string]string{
				dialect.Postgres: "text",
			}).
			Immutable().
			Sensitive(),
		field.UUID("session_id", uuid.UUID{}).
			Immutable(),
		field.Time("expires_at"),
		field.Time("used_at").
			Optional().
			Nillable(),
	}
}

// Edges of the AccountRefreshToken.
func (AccountRefreshToken) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("session", AccountSession.Type).
			Ref("refresh_tokens").
			Field("session_id").
			Unique().
			Required().
			Immutable(),
	}
}

func (AccountRefreshToken) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("session_id").
			StorageKey("account_refresh_tokens_session_id"),
		index.Fields("expires_at", "id").
			StorageKey("account_refresh _tokens_expires_at"),
	}
}

func (AccountRefreshToken) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{
			Table: "account_refresh_tokens",
		},
	}
}
