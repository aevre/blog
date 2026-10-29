// Package repository persists and queries link data.
package repository

import (
	"context"
	"strings"

	"blog-server/ent/link"
	"blog-server/internal/link/mapper"
	linkmodel "blog-server/internal/link/model"
	"blog-server/internal/platform/database"
	"blog-server/internal/shared/errx"

	"entgo.io/ent/dialect/sql"
)

// LinkRepo defines persistence operations for the Link aggregate.
//
// All read operations implicitly apply:
// - soft-delete filter (DeletedAt IS NULL)
type LinkRepo interface {
	Create(ctx context.Context, link *linkmodel.Link) (*linkmodel.Link, error)
	GetAll(ctx context.Context) ([]*linkmodel.Link, error)
	GetAllEnabled(ctx context.Context) ([]*linkmodel.Link, error)
	UpdateStatusBatch(ctx context.Context, updates map[uint]linkmodel.LinkStatus) error
	IsOwner(ctx context.Context, userID uint, linkID uint) (bool, error)
}

type linkRepo struct {
	ds *database.DataStore
}

// NewLinkRepo creates a LinkRepo instance backed by database.
func NewLinkRepo(ds *database.DataStore) LinkRepo {
	return &linkRepo{ds: ds}
}

// Create inserts a new link record.
//
// Optional fields (description, avatar) are only persisted if non-empty after trimming.
func (r *linkRepo) Create(ctx context.Context, l *linkmodel.Link) (*linkmodel.Link, error) {
	c := r.ds.Client(ctx).Link.
		Create().
		SetName(l.Name).
		SetURL(l.URL)

	if l.Description != nil && strings.TrimSpace(*l.Description) != "" {
		c.SetDescription(*l.Description)
	}
	if l.Avatar != nil && strings.TrimSpace(*l.Avatar) != "" {
		c.SetAvatar(*l.Avatar)
	}

	created, err := c.Save(ctx)
	if err != nil {
		return nil, errx.New(errx.CodeInternalError, err)
	}

	return mapper.ToLink(created), nil
}

// GetAll returns all non-deleted links ordered by ID descending.
func (r *linkRepo) GetAll(ctx context.Context) ([]*linkmodel.Link, error) {
	links, err := r.ds.Client(ctx).Link.
		Query().
		Where(
			link.DeletedAtIsNil(),
		).
		Order(
			link.ByID(sql.OrderDesc()),
			link.BySortOrder(sql.OrderAsc()),
		).
		All(ctx)
	if err != nil {
		return nil, errx.New(errx.CodeInternalError, err)
	}
	return mapper.ToLinks(links), nil
}

// GetAllEnabled returns all enabled and non-deleted links ordered by ID descending.
func (r *linkRepo) GetAllEnabled(ctx context.Context) ([]*linkmodel.Link, error) {
	links, err := r.ds.Client(ctx).Link.
		Query().
		Where(
			link.EnabledEQ(true),
			link.DeletedAtIsNil(),
		).
		Order(
			link.ByID(sql.OrderDesc()),
			link.BySortOrder(sql.OrderAsc()),
		).
		All(ctx)
	if err != nil {
		return nil, errx.New(errx.CodeInternalError, err)
	}
	return mapper.ToLinks(links), nil
}

// UpdateStatusBatch performs a bulk update of link status values.
//
// Implementation details:
// - Uses a single SQL UPDATE with CASE expression
// - Each ID is mapped to its corresponding status value
// - Empty input results in a no-op
func (r *linkRepo) UpdateStatusBatch(ctx context.Context, updates map[uint]linkmodel.LinkStatus) error {
	if len(updates) == 0 {
		return nil
	}

	ids := make([]uint, 0, len(updates))
	for id := range updates {
		ids = append(ids, id)
	}

	err := r.ds.Client(ctx).Link.
		Update().
		Where(link.IDIn(ids...)).
		Modify(func(u *sql.UpdateBuilder) {
			caseExpr := func(b *sql.Builder) {
				b.WriteString("CASE ")
				b.Ident(link.FieldID)
				b.WriteString(" ")

				for _, id := range ids {
					b.WriteString("WHEN ")
					b.Arg(id)
					b.WriteString(" THEN ")
					b.Arg(updates[id])
					b.WriteString(" ")
				}

				b.WriteString("ELSE ")
				b.Ident(link.FieldStatus)
				b.WriteString(" END")
			}
			u.Set(link.FieldStatus, sql.ExprFunc(caseExpr))
		}).
		Exec(ctx)
	if err != nil {
		return errx.New(errx.CodeInternalError, err)
	}

	return nil
}

// IsOwner checks whether a user owns the specified link.
func (r *linkRepo) IsOwner(ctx context.Context, userID uint, linkID uint) (bool, error) {
	count, err := r.ds.Client(ctx).Link.
		Query().
		Where(
			link.IDEQ(linkID),
		).
		Count(ctx)
	if err != nil {
		return false, errx.New(errx.CodeInternalError, err)
	}

	return count > 0, nil
}
