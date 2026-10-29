package schema

import (
	accountmodel "blog-server/internal/account/model"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// User holds the schema definition for the User entity.
type User struct {
	ent.Schema
}

func (User) Mixin() []ent.Mixin {
	return []ent.Mixin{
		TimeMixin{},
	}
}

// Fields of the User.
func (User) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("uuid", uuid.UUID{}).
			Default(uuid.New).
			Unique(),

		field.String("avatar").
			MaxLen(255).
			Optional().
			Nillable(),

		field.String("email").
			MaxLen(100).
			Unique(),

		field.String("password").
			Sensitive().
			MaxLen(255),

		field.Enum("role").
			GoType(accountmodel.UserRole("")).
			Default(string(accountmodel.UserRoleReader)),

		field.String("username").
			MaxLen(50),
	}
}

// Edges of the User.
func (User) Edges() []ent.Edge {
	return []ent.Edge{
		edge.
			To("posts", Post.Type),
	}
}
