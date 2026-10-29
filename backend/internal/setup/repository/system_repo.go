package repository

import (
	"context"
	"time"

	"blog-server/ent/system"
	"blog-server/internal/platform/database"
	"blog-server/internal/shared/errx"
)

type SystemRepo interface {
	IsInitialized(ctx context.Context) (bool, error)
	IsInitializedForUpdate(ctx context.Context) (bool, error)
	MarkInitialized(ctx context.Context, initializedAt time.Time) error
}

type systemRepo struct {
	ds *database.DataStore
}

func NewSystemRepo(ds *database.DataStore) SystemRepo {
	return &systemRepo{
		ds,
	}
}

func (r *systemRepo) IsInitialized(ctx context.Context) (bool, error) {
	s, err := r.ds.Client(ctx).System.
		Query().
		Where(system.IDEQ(1)).
		Only(ctx)
	if err != nil {
		return false, errx.New(errx.CodeInternalError, err)
	}

	return s.Initialized, nil
}

func (r *systemRepo) IsInitializedForUpdate(ctx context.Context) (bool, error) {
	s, err := r.ds.Client(ctx).System.
		Query().
		Where(system.IDEQ(1)).
		ForUpdate().
		Only(ctx)
	if err != nil {
		return false, errx.New(errx.CodeInternalError, err)
	}

	return s.Initialized, nil
}

func (r *systemRepo) MarkInitialized(ctx context.Context, initializedAt time.Time) error {
	_, err := r.ds.Client(ctx).System.
		UpdateOneID(1).
		SetInitialized(true).
		SetInitializedAt(initializedAt).
		SetUpdatedAt(time.Now()).
		Save(ctx)
	if err != nil {
		return errx.New(errx.CodeInternalError, err)
	}
	return nil
}
