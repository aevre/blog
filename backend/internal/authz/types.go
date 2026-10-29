package authz

type Resource string

const (
	ResourcePost Resource = "post"
	ResourceLink Resource = "link"
)

type Action string

const (
	ActionCreate Action = "create"
	ActionRead   Action = "read"
	ActionUpdate Action = "update"
	ActionDelete Action = "delete"
)

type Permission struct {
	Resource Resource
	Action   Action
}
