// Package repository persists and queries link data.
package repository

import (
	"context"
	"errors"
	"strings"
	"time"

	"blog-server/ent"
	"blog-server/ent/link"
	"blog-server/ent/predicate"
	"blog-server/internal/link/mapper"
	linkmodel "blog-server/internal/link/model"
	"blog-server/internal/platform/database"
	"blog-server/internal/platform/field"
	"blog-server/internal/shared/errx"

	"entgo.io/ent/dialect/sql"
)

type CreateParams struct {
	SubmitterID uint

	Name        string
	URL         string
	Description *string
	Avatar      *string
	CategoryID  *uint
}

type UpdateContentParams struct {
	ID              uint
	ActorID         uint
	IsAdmin         bool
	ExpectedVersion int64

	Name        field.Optional[string]
	URL         field.Optional[string]
	Description field.Optional[string]
	Avatar      field.Optional[string]
	SortOrder   field.Optional[int]
	CategoryID  field.Optional[uint]
}

type ReviewParams struct {
	ID              uint
	ExpectedVersion int64
	Decision        linkmodel.LinkReviewStatus
	Reason          *string
}

// LinkRepo defines persistence operations for the Link aggregate.
//
// All read operations implicitly apply:
// - soft-delete filter (DeletedAt IS NULL)
type LinkRepo interface {
	ListPublic(ctx context.Context) ([]*linkmodel.Link, error)

	GetMineByID(ctx context.Context, submitterID uint, linkID uint) (*linkmodel.Link, error)

	ListAll(ctx context.Context) ([]*linkmodel.Link, error)

	GetByID(ctx context.Context, id uint) (*linkmodel.Link, error)

	Create(ctx context.Context, in *CreateParams) (*linkmodel.Link, error)

	UpdateContent(ctx context.Context, in *UpdateContentParams) (*linkmodel.Link, error)

	Review(ctx context.Context, in *ReviewParams) error

	DeleteMine(ctx context.Context, submitterID uint, linkID uint, expectedVersion int64) error

	DeleteAdmin(ctx context.Context, linkID uint, expectedVersion int64) error

	UpdateHealthStatusBatch(ctx context.Context, updates map[uint]linkmodel.LinkStatus) error
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
func (r *linkRepo) Create(ctx context.Context, in *CreateParams) (*linkmodel.Link, error) {
	if in.SubmitterID == 0 {
		return nil, errx.New(errx.CodeInvalidParam, errors.New("submitter ID is required"))
	}

	builder := r.ds.Client(ctx).Link.
		Create().
		SetSubmitterID(in.SubmitterID).
		SetName(in.Name).
		SetURL(in.URL).
		SetReviewStatus(linkmodel.LinkReviewPending).
		SetVersion(1)

	if in.Description != nil {
		builder.SetDescription(*in.Description)
	}
	if in.Avatar != nil {
		builder.SetAvatar(*in.Avatar)
	}
	if in.CategoryID != nil {
		builder.SetCategoryID(*in.CategoryID)
	}

	created, err := builder.Save(ctx)
	if err != nil {
		return nil, errx.New(errx.CodeInternalError, err)
	}

	return mapper.ToLink(created), nil
}

func (r *linkRepo) ListPublic(ctx context.Context) ([]*linkmodel.Link, error) {
	links, err := r.ds.Client(ctx).Link.
		Query().
		Where(
			link.ReviewStatusEQ(linkmodel.LinkReviewApproved),
			link.DeletedAtIsNil(),
		).
		Order(
			link.BySortOrder(sql.OrderAsc()),
			link.ByID(sql.OrderDesc()),
		).
		All(ctx)
	if err != nil {
		return nil, errx.New(errx.CodeInternalError, err)
	}
	return mapper.ToLinks(links), nil
}

// ListAll returns all non-deleted links ordered by ID descending.
func (r *linkRepo) ListAll(ctx context.Context) ([]*linkmodel.Link, error) {
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

func (r *linkRepo) DeleteAdmin(ctx context.Context, linkID uint, expectedVersion int64) error {
	if linkID == 0 || expectedVersion < 1 {
		return errx.New(errx.CodeInvalidParam, errors.New("invalid delete parameters"))
	}

	now := time.Now()
	affected, err := r.ds.Client(ctx).Link.
		Update().
		Where(
			link.IDEQ(linkID),
			link.DeletedAtIsNil(),
			link.VersionEQ(expectedVersion),
		).
		SetDeletedAt(now).
		SetUpdatedAt(now).
		SetVersion(expectedVersion + 1).
		Save(ctx)
	if err != nil {
		return errx.New(errx.CodeInternalError, err)
	}
	if affected == 0 {
		return errx.New(errx.CodeConflict, errors.New("link changed or not found"))
	}
	return nil
}

func (r *linkRepo) DeleteMine(ctx context.Context, submitterID uint, linkID uint, expectedVersion int64) error {
	if submitterID == 0 || linkID == 0 || expectedVersion < 1 {
		return errx.New(errx.CodeInvalidParam, errors.New("invalid delete parameters"))
	}

	now := time.Now()
	affected, err := r.ds.Client(ctx).Link.
		Update().
		Where(
			link.IDEQ(linkID),
			link.SubmitterIDEQ(submitterID),
			link.DeletedAtIsNil(),
			link.VersionEQ(expectedVersion),
		).
		SetDeletedAt(now).
		SetUpdatedAt(now).
		SetVersion(expectedVersion + 1).
		Save(ctx)
	if err != nil {
		return errx.New(errx.CodeInternalError, err)
	}
	if affected == 0 {
		return errx.New(errx.CodeNotFound, errors.New("link not found"))
	}
	return nil
}

func (r *linkRepo) Review(ctx context.Context, in *ReviewParams) error {
	if in.ID == 0 || in.ExpectedVersion < 1 {
		return errx.New(errx.CodeInvalidParam, errors.New("invalid link ID or version"))
	}

	switch in.Decision {
	case linkmodel.LinkReviewApproved:
	case linkmodel.LinkReviewRejected:
		if in.Reason == nil || strings.TrimSpace(*in.Reason) == "" {
			return errx.New(errx.CodeValidationFailed, errors.New("rejection reason is required"))
		}
	default:
		return errx.New(errx.CodeValidationFailed, errors.New("rejection review decision"))
	}

	now := time.Now()
	builder := r.ds.Client(ctx).Link.
		Update().
		Where(
			link.IDEQ(in.ID),
			link.DeletedAtIsNil(),
			link.ReviewStatusNEQ(in.Decision),
			link.VersionEQ(in.ExpectedVersion),
		).
		SetReviewStatus(in.Decision).
		SetReviewedAt(now).
		SetUpdatedAt(now).
		SetVersion(in.ExpectedVersion + 1)

	switch in.Decision {
	case linkmodel.LinkReviewApproved:
		builder.ClearRejectionReason()
	case linkmodel.LinkReviewRejected:
		builder.SetRejectionReason(*in.Reason)

	}

	affected, err := builder.Save(ctx)
	if err != nil {
		return errx.New(errx.CodeInternalError, err)
	}
	if affected == 0 {
		return errx.New(errx.CodeConflict, errors.New("link changed or review already applied"))
	}

	return nil
}

func (r *linkRepo) GetMineByID(ctx context.Context, submitterID uint, linkID uint) (*linkmodel.Link, error) {
	if submitterID == 0 || linkID == 0 {
		return nil, errx.New(errx.CodeInvalidParam, errors.New("both submitterID and linkID must be provided"))
	}
	entity, err := r.ds.Client(ctx).Link.
		Query().
		Where(
			link.IDEQ(linkID),
			link.SubmitterIDEQ(submitterID),
			link.DeletedAtIsNil(),
		).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errx.New(errx.CodeNotFound, err)
		}
		return nil, errx.New(errx.CodeInternalError, err)
	}
	return mapper.ToLink(entity), nil
}

func (r *linkRepo) GetByID(ctx context.Context, id uint) (*linkmodel.Link, error) {
	if id == 0 {
		return nil, errx.New(errx.CodeInvalidParam, errors.New("link ID is required"))
	}

	entity, err := r.ds.Client(ctx).Link.
		Query().
		Where(
			link.IDEQ(id),
			link.DeletedAtIsNil(),
		).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errx.New(errx.CodeNotFound, err)
		}
		return nil, errx.New(errx.CodeInternalError, err)
	}
	return mapper.ToLink(entity), nil
}

// UpdateHealthStatusBatch performs a bulk update of link status values.
//
// Implementation details:
// - Uses a single SQL UPDATE with CASE expression
// - Each ID is mapped to its corresponding status value
// - Empty input results in a no-op
func (r *linkRepo) UpdateHealthStatusBatch(ctx context.Context, updates map[uint]linkmodel.LinkStatus) error {
	if len(updates) == 0 {
		return nil
	}

	ids := make([]uint, 0, len(updates))
	for id := range updates {
		ids = append(ids, id)
	}

	if len(ids) == 0 {
		return nil
	}

	err := r.ds.Client(ctx).Link.
		Update().
		Where(
			link.IDIn(ids...),
			link.DeletedAtIsNil(),
		).
		Modify(func(builder *sql.UpdateBuilder) {
			builder.Set(
				link.FieldStatus,
				sql.ExprFunc(func(b *sql.Builder) {
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
				}),
			)
		}).
		Exec(ctx)
	if err != nil {
		return errx.New(errx.CodeInternalError, err)
	}

	return nil
}

func (r *linkRepo) UpdateContent(ctx context.Context, in *UpdateContentParams) (*linkmodel.Link, error) {
	if in == nil || in.ID == 0 || in.ActorID == 0 || in.ExpectedVersion < 1 {
		return nil, errx.New(errx.CodeInvalidParam, errors.New("invalid update parameters"))
	}

	predicates := []predicate.Link{
		link.DeletedAtIsNil(),
		link.VersionEQ(in.ExpectedVersion),
	}

	now := time.Now()

	if !in.IsAdmin {
		predicates = append(predicates, link.SubmitterIDEQ(in.ActorID))
	}

	builder := r.ds.Client(ctx).Link.
		UpdateOneID(in.ID).
		Where(predicates...).
		SetVersion(in.ExpectedVersion + 1).
		SetUpdatedAt(now)

	if !in.IsAdmin {
		builder.SetReviewStatus(linkmodel.LinkReviewPending).
			ClearReviewedAt().
			ClearRejectionReason()
	}

	if in.Name.Set {
		if in.Name.Value == nil || strings.TrimSpace(*in.Name.Value) == "" {
			return nil, errx.New(errx.CodeValidationFailed, errors.New("name cannot be empty"))
		}
		builder.SetName(strings.TrimSpace(*in.Name.Value))
	}

	if in.URL.Set {
		if in.URL.Value == nil || strings.TrimSpace(*in.URL.Value) == "" {
			return nil, errx.New(errx.CodeValidationFailed, errors.New("URL cannot be empty"))
		}
		builder.SetURL(strings.TrimSpace(*in.URL.Value))
	}

	if in.Description.Set {
		if in.Description.Value == nil {
			builder.ClearDescription()
		} else {
			builder.SetDescription(strings.TrimSpace(*in.Description.Value))
		}
	}
	if in.Avatar.Set {
		if in.Avatar.Value == nil {
			builder.ClearAvatar()
		} else {
			builder.SetAvatar(strings.TrimSpace(*in.Avatar.Value))
		}
	}
	if in.CategoryID.Set {
		if in.CategoryID.Value == nil {
			builder.ClearCategoryID()
		} else {
			builder.SetCategoryID(*in.CategoryID.Value)
		}
	}

	if in.SortOrder.Set {
		if in.SortOrder.Value == nil {
			return nil, errx.New(errx.CodeValidationFailed, errors.New("soreOrder cannot be null"))
		}
		builder.SetSortOrder(*in.SortOrder.Value)
	}

	affected, err := builder.Save(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			if !in.IsAdmin {
				return nil, errx.New(errx.CodeNotFound, errors.New("link not found"))
			}
			return nil, errx.New(errx.CodeConflict, errors.New("link changed or not found"))
		}
		return nil, errx.New(errx.CodeInternalError, err)
	}

	return mapper.ToLink(affected), nil
}
