package service

import (
	"context"

	accountmodel "blog-server/internal/account/model"
	"blog-server/internal/account/repository"
	"blog-server/internal/shared/errx"
	"blog-server/internal/shared/txmgr"

	"github.com/google/uuid"
)

type CreateUserInput struct {
	Email    string
	Username string
	Password string
}

type UserService interface {
	CreateAdmin(ctx context.Context, in *CreateUserInput) (*accountmodel.User, error)
	GetByID(ctx context.Context, id uint) (*accountmodel.User, error)
}

type userService struct {
	tx   txmgr.TxManager
	user repository.UserRepo
}

func (s *userService) GetByID(ctx context.Context, id uint) (*accountmodel.User, error) {
	return s.user.GetByID(ctx, id)
}

func (s *userService) CreateAdmin(ctx context.Context, in *CreateUserInput) (*accountmodel.User, error) {
	hashedPassword, err := hashPassword(in.Password)
	if err != nil {
		return nil, errx.New(errx.CodeInternalError, err)
	}

	user := &accountmodel.User{
		UUID:     uuid.New(),
		Email:    in.Email,
		Username: in.Username,
		Role:     accountmodel.UserRoleAdmin,
	}

	return s.user.Create(ctx, user, hashedPassword)
}

func NewUserService(tx txmgr.TxManager, user repository.UserRepo) UserService {
	return &userService{
		tx,
		user,
	}
}
