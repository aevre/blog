// Package repository persists site setup data.
package repository

import (
	"context"
	"time"

	"blog-server/ent/site"
	"blog-server/internal/platform/database"
	setupmodel "blog-server/internal/setup/model"
	"blog-server/internal/shared/errx"
)

type SiteRepo interface {
	Create(ctx context.Context, site *setupmodel.Site) error
	Get(ctx context.Context) (*setupmodel.Site, error)
	Update(ctx context.Context, site *setupmodel.Site) error
}

type siteRepo struct {
	ds *database.DataStore
}

func NewSiteRepo(ds *database.DataStore) SiteRepo {
	return &siteRepo{
		ds,
	}
}

func (r *siteRepo) Create(ctx context.Context, site *setupmodel.Site) error {
	_, err := r.ds.Client(ctx).Site.
		Create().
		SetID(1).
		SetName(site.Name).
		SetNillableLogo(site.Logo).
		SetNillableGreeting(site.Greeting).
		SetNillableDescription(site.Description).
		SetCreatedAt(time.Now()).
		SetUpdatedAt(time.Now()).
		Save(ctx)
	if err != nil {
		return errx.New(errx.CodeInternalError, err)
	}

	return nil
}

func (r *siteRepo) Get(ctx context.Context) (*setupmodel.Site, error) {
	s, err := r.ds.Client(ctx).Site.
		Query().
		Where(site.IDEQ(1)).
		Only(ctx)
	if err != nil {
		return nil, errx.New(errx.CodeInternalError, err)
	}

	return &setupmodel.Site{
		Name:        s.Name,
		Logo:        &s.Logo,
		Greeting:    &s.Greeting,
		Description: &s.Description,
		CreatedAt:   &s.CreatedAt,
	}, nil
}

func (r *siteRepo) Update(ctx context.Context, site *setupmodel.Site) error {
	_, err := r.ds.Client(ctx).Site.
		UpdateOneID(1).
		SetName(site.Name).
		SetNillableLogo(site.Logo).
		SetNillableGreeting(site.Greeting).
		SetNillableDescription(site.Description).
		SetUpdatedAt(time.Now()).
		Save(ctx)
	if err != nil {
		return errx.New(errx.CodeInternalError, err)
	}

	return nil
}
