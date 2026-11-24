// Package service implements initial site setup.
package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"time"

	accountservice "blog-server/internal/account/service"
	bootstrap "blog-server/internal/bootstrap/model"
	"blog-server/internal/bootstrap/repository"
	logger "blog-server/internal/platform/logging"
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
	Token       string
}

type BootstrapService interface {
	Status(ctx context.Context) (bool, error)
	Initialize(ctx context.Context, input *InitializeInput) error
	EnsureBootstrapToken(ctx context.Context) error
}

type bootstrapService struct {
	tx   txmgr.TxManager
	sr   repository.SiteRepo
	sysr repository.SystemRepo
	us   accountservice.UserService
	log  logger.Logger
}

func NewBootstrapService(
	tx txmgr.TxManager,
	sr repository.SiteRepo,
	sysr repository.SystemRepo,
	us accountservice.UserService,
	log logger.Logger,
) BootstrapService {
	return &bootstrapService{tx, sr, sysr, us, log}
}

func (s *bootstrapService) Status(ctx context.Context) (bool, error) {
	return s.sysr.IsInitialized(ctx)
}

func (s *bootstrapService) Initialize(ctx context.Context, input *InitializeInput) error {
	return s.tx.WithTx(ctx, func(ctx context.Context) error {
		initialized, err := s.sysr.IsInitializedForUpdate(ctx)
		if err != nil {
			return err
		}

		if initialized {
			return errx.New(errx.CodeConflict, fmt.Errorf("already initialized"))
		}

		storeHash, err := s.sysr.BootstrapTokenHash(ctx)
		if err != nil {
			return err
		}

		submitted := sha256.Sum256([]byte(input.Token))
		submittedHash := hex.EncodeToString(submitted[:])
		if storeHash == nil || subtle.ConstantTimeCompare([]byte(submittedHash), []byte(*storeHash)) != 1 {
			return errx.New(errx.CodeUnauthorized, fmt.Errorf("invalid bootstrap token"))
		}

		if _, err = s.us.CreateAdmin(ctx, &accountservice.CreateUserInput{
			Username: input.Username,
			Password: input.Password,
			Email:    input.Email,
		}); err != nil {
			return err
		}

		if err = s.sr.Create(ctx, &bootstrap.Site{
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

func (s *bootstrapService) EnsureBootstrapToken(ctx context.Context) error {
	return s.tx.WithTx(ctx, func(ctx context.Context) error {
		initialized, err := s.sysr.IsInitializedForUpdate(ctx)
		if err != nil {
			return err
		}

		if initialized {
			return nil
		}

		hash, err := s.sysr.BootstrapTokenHash(ctx)
		if err != nil {
			return err
		}
		if hash != nil && *hash != "" {
			return nil
		}

		token, tokenHash, err := s.newBootstrapToken()
		if err != nil {
			return err
		}

		if err := s.sysr.SetBootstrapTokenHash(ctx, tokenHash); err != nil {
			return err
		}

		s.log.Info("bootstrap token generated",
			logger.String("token", token))

		return nil
	})
}

func (s *bootstrapService) newBootstrapToken() (string, string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", "", err
	}

	token := hex.EncodeToString(raw)
	sum := sha256.Sum256([]byte(token))
	return token, hex.EncodeToString(sum[:]), nil
}
