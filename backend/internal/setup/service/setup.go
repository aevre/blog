// Package service implements initial site setup.
package service

import (
	"context"
	"fmt"
	"time"

	accountservice "blog-server/internal/account/service"
	setupmodel "blog-server/internal/setup/model"
	"blog-server/internal/setup/repository"
	"blog-server/internal/shared/errx"
	"blog-server/internal/shared/txmgr"
)

type InitializeInput struct {
	Username    string
	Email       string
	Password    string
	SiteName    string
	Logo        string
	Greeting    string
	Description string
}

type SetupService interface {
	Status(ctx context.Context) (bool, error)
	Initialize(ctx context.Context, input *InitializeInput) error
}

type setupService struct {
	tx   txmgr.TxManager
	sr   repository.SiteRepo
	sysr repository.SystemRepo
	us   accountservice.UserService
}

func NewSetupService(
	tx txmgr.TxManager,
	sr repository.SiteRepo,
	sysr repository.SystemRepo,
	us accountservice.UserService,
) SetupService {
	return &setupService{tx, sr, sysr, us}
}

func (s *setupService) Status(ctx context.Context) (bool, error) {
	return s.sysr.IsInitialized(ctx)
}

func (s *setupService) Initialize(ctx context.Context, input *InitializeInput) error {
	return s.tx.WithTx(ctx, func(ctx context.Context) error {
		initialized, err := s.sysr.IsInitializedForUpdate(ctx)
		if err != nil {
			return err
		}

		if initialized {
			return errx.New(errx.CodeConflict, fmt.Errorf("already initialized"))
		}

		if _, err = s.us.CreateAdmin(ctx, &accountservice.CreateUserInput{
			Username: input.Username,
			Password: input.Password,
			Email:    input.Email,
		}); err != nil {
			return err
		}

		if err = s.sr.Create(ctx, &setupmodel.Site{
			Name:        input.SiteName,
			Logo:        &input.Logo,
			Greeting:    &input.Greeting,
			Description: &input.Description,
		}); err != nil {
			return err
		}

		return s.sysr.MarkInitialized(ctx, time.Now())
	})
}
