// Package authz provides authorization (RBAC + ABAC) components.
package authz

import "go.uber.org/fx"

// Module registers authorization-related components into the fx graph.
func Module() fx.Option {
	return fx.Module(
		"authz",
		fx.Provide(
			NewAuthorizer,
		),
	)
}
