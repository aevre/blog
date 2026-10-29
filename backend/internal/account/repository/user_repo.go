// Package repository persists account and session data.
package repository

import (
	"context"

	"blog-server/ent"
	"blog-server/ent/user"
	"blog-server/internal/account/mapper"
	accountmodel "blog-server/internal/account/model"
	"blog-server/internal/platform/database"
	"blog-server/internal/shared/errx"

	"github.com/google/uuid"
)

// UserRepo defines persistence operations for user aggregate.
//
// It encapsulates both write and read access, including authentication-related projections
// and existence checks. All queries automatically exclude soft-deleted records.
type UserRepo interface {
	Create(ctx context.Context, user *accountmodel.User, hashPassword string) (*accountmodel.User, error)
	GetByID(ctx context.Context, id uint) (*accountmodel.User, error)
	GetByEmail(ctx context.Context, email string) (*accountmodel.User, error)

	ExistsByEmail(ctx context.Context, email string) (bool, error)
	ExistsByUUID(ctx context.Context, uuidStr string) (bool, error)
	ExistsByID(ctx context.Context, id uint) (bool, error)

	GetAuthByEmail(ctx context.Context, email string) (*accountmodel.UserAuth, error)
	GetAuthByID(ctx context.Context, id uint) (*accountmodel.UserAuth, error)
}

// userRepo implements UserRepo using ent ORM.
type userRepo struct {
	ds *database.DataStore
}

// NewUserRepo constructs a UserRepo backed by DataStore.
func NewUserRepo(ds *database.DataStore) UserRepo {
	return &userRepo{ds: ds}
}

// baseQuery returns a query builder with global constraints applied.
//
// Currently enforces soft-delete filtering (DeletedAt IS NULL).
func (r *userRepo) baseQuery(ctx context.Context) *ent.UserQuery {
	return r.ds.Client(ctx).User.
		Query().
		Where(user.DeletedAtIsNil())
}

// Create inserts a new user record with hashed password.
//
// The input entity is not modified. Optional fields (e.g. Avatar) are applied only if present.
func (r *userRepo) Create(ctx context.Context, u *accountmodel.User, hashPassword string) (*accountmodel.User, error) {
	c := r.ds.Client(ctx).User.Create().
		SetUUID(u.UUID).
		SetEmail(u.Email).
		SetPassword(hashPassword).
		SetRole(u.Role).
		SetUsername(u.Username)

	if u.Avatar != nil {
		c.SetAvatar(*u.Avatar)
	}

	created, err := c.Save(ctx)
	if err != nil {
		return nil, errx.New(errx.CodeInternalError, err)
	}

	return mapper.ToUser(created), nil
}

// GetByID retrieves a user by ID.
func (r *userRepo) GetByID(ctx context.Context, id uint) (*accountmodel.User, error) {
	u, err := r.baseQuery(ctx).
		Where(user.IDEQ(id)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errx.New(errx.CodeNotFound, err)
		}
		return nil, errx.New(errx.CodeInternalError, err)
	}

	return mapper.ToUser(u), nil
}

// GetByEmail retrieves a user by email.
//
// Only non-soft-deleted users are returned.
// Returns ErrNotFound wrapped in domain error if no match exists.
func (r *userRepo) GetByEmail(ctx context.Context, email string) (*accountmodel.User, error) {
	u, err := r.baseQuery(ctx).
		Where(user.EmailEQ(email)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errx.New(errx.CodeNotFound, err)
		}
		return nil, errx.New(errx.CodeInternalError, err)
	}

	return mapper.ToUser(u), nil
}

// ExistsByID checks whether a user exists by ID.
//
// Soft-deleted users are excluded.
func (r *userRepo) ExistsByID(ctx context.Context, id uint) (bool, error) {
	exists, err := r.baseQuery(ctx).
		Where(user.IDEQ(id)).
		Exist(ctx)
	if err != nil {
		return false, errx.New(errx.CodeInternalError, err)
	}
	return exists, nil
}

// ExistsByUUID checks whether a user exists by UUID.
//
// Input UUID string must be valid RFC4122 format.
// Soft-deleted users are excluded.
func (r *userRepo) ExistsByUUID(ctx context.Context, uuidStr string) (bool, error) {
	uid, err := uuid.Parse(uuidStr)
	if err != nil {
		return false, errx.New(errx.CodeInternalError, err)
	}

	exists, err := r.baseQuery(ctx).
		Where(user.UUIDEQ(uid)).
		Exist(ctx)
	if err != nil {
		return false, errx.New(errx.CodeInternalError, err)
	}
	return exists, nil
}

// ExistsByEmail checks whether a user exists by email.
//
// Soft-deleted users are excluded.
func (r *userRepo) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	exists, err := r.baseQuery(ctx).
		Where(user.EmailEQ(email)).
		Exist(ctx)
	if err != nil {
		return false, errx.New(errx.CodeInternalError, err)
	}
	return exists, nil
}

// GetAuthByEmail returns authentication projection for a user identified by email.
//
// Only minimal fields required for authentication are selected (ID, Password, Role).
// Soft-deleted users are excluded.
func (r *userRepo) GetAuthByEmail(ctx context.Context, email string) (*accountmodel.UserAuth, error) {
	u, err := r.baseQuery(ctx).
		Where(user.EmailEQ(email)).
		Select(
			user.FieldID,
			user.FieldPassword,
			user.FieldRole,
		).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errx.New(errx.CodeNotFound, err)
		}
		return nil, errx.New(errx.CodeInternalError, err)
	}

	return &accountmodel.UserAuth{
		ID:       u.ID,
		Password: u.Password,
		Role:     u.Role,
	}, nil
}

// GetAuthByID returns authentication projection for a user identified by ID.
//
// Only minimal authentication fields are selected.
// Soft-deleted users are excluded.
func (r *userRepo) GetAuthByID(ctx context.Context, id uint) (*accountmodel.UserAuth, error) {
	u, err := r.baseQuery(ctx).
		Where(user.IDEQ(id)).
		Select(
			user.FieldID,
			user.FieldPassword,
			user.FieldRole,
		).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, errx.New(errx.CodeNotFound, err)
		}
		return nil, errx.New(errx.CodeInternalError, err)
	}

	return &accountmodel.UserAuth{
		ID:       u.ID,
		Password: u.Password,
		Role:     u.Role,
	}, nil
}
