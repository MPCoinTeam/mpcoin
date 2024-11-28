package usecase

import (
	"context"
	"errors"
	"mpc/internal/domain"
	"mpc/internal/repository"
	"mpc/pkg/utils"

	"github.com/google/uuid"
)

type UserUseCase interface {
	GetUser(ctx context.Context, id uuid.UUID) (domain.User, error)
	UpdateUser(ctx context.Context, params domain.UpdateUserParams) (domain.User, error)
	GetUserWallet(ctx context.Context, id uuid.UUID) (domain.UserWithWallet, error)
	GetUserByAddress(ctx context.Context, address string) (domain.UserWithWallet, error)
}

type userUseCase struct {
	userRepo repository.UserRepository
}

func NewUserUC(userRepo repository.UserRepository) UserUseCase {
	return &userUseCase{userRepo: userRepo}
}

var _ UserUseCase = (*userUseCase)(nil)

func (uc *userUseCase) GetUser(ctx context.Context, id uuid.UUID) (domain.User, error) {
	return uc.userRepo.GetUser(ctx, id)
}

func (uc *userUseCase) UpdateUser(ctx context.Context, params domain.UpdateUserParams) (domain.User, error) {
	// Check if user exists
	existingUser, err := uc.userRepo.GetUser(ctx, params.ID)
	if err != nil {
		return domain.User{}, errors.New("user not found")
	}

	// Update fields if provided
	if params.Email != "" {
		existingUser.Email = params.Email
	}
	if params.Password != "" {
		hashedPassword, err := utils.HashPassword(params.Password)
		if err != nil {
			return domain.User{}, err
		}
		existingUser.PasswordHash = hashedPassword
	}

	return uc.userRepo.UpdateUser(ctx, existingUser)
}

func (uc *userUseCase) GetUserWallet(ctx context.Context, id uuid.UUID) (domain.UserWithWallet, error) {
	user, err := uc.userRepo.GetUserWithWallet(ctx, id)

	user.Avatar = "https://robohash.org/af424cffba2a77572a76dc071c0799dc?set=set4&bgset=&size=400x400"
	user.Name = "John Doe"
	return user, err
}

func (uc *userUseCase) GetUserByAddress(ctx context.Context, address string) (domain.UserWithWallet, error) {
	user, err := uc.userRepo.GetUserByAddress(ctx, address)
	user.Avatar = "https://robohash.org/af424cffba2a77572a76dc071c0799dc?set=set4&bgset=&size=400x400"
	if user.Name == "" {
		user.Name = "John Doe"
	}
	return user, err
}
