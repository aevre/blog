package service

import (
	"context"

	"blog-server/entity"
	"blog-server/pkg/errx"
	"blog-server/pkg/txmgr"
	"blog-server/repository"
	"blog-server/utils"

	"github.com/google/uuid"
)

type CreateUserInput struct {
	Email    string
	Username string
	Password string
}

type UserService interface {
	CreateAdmin(ctx context.Context, in *CreateUserInput) (*entity.User, error)
	GetByID(ctx context.Context, id uint) (*entity.User, error)
}

type userService struct {
	tx   txmgr.TxManager
	user repository.UserRepo
}

func (s *userService) GetByID(ctx context.Context, id uint) (*entity.User, error) {
	return s.user.GetByID(ctx, id)
}

func (s *userService) CreateAdmin(ctx context.Context, in *CreateUserInput) (*entity.User, error) {
	hashedPassword, err := utils.HashPassword(in.Password)
	if err != nil {
		return nil, errx.New(errx.CodeInternalError, err)
	}

	user := &entity.User{
		UUID:     uuid.New(),
		Email:    in.Email,
		Username: in.Username,
		Role:     entity.UserRoleAdmin,
	}

	return s.user.Create(ctx, user, hashedPassword)
}

func NewUserService(tx txmgr.TxManager, user repository.UserRepo) UserService {
	return &userService{
		tx,
		user,
	}
}
