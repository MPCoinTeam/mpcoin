package domain

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID
	Email        string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type UserCredentials struct {
	Email    string
	Password string
}

type CreateUserParams UserCredentials

type LoginUserParams UserCredentials

type UpdateUserParams struct {
	ID uuid.UUID
	UserCredentials
}

type CreateHashedUserParams struct {
	Email        string
	PasswordHash string
}

type CreateUserResponse struct {
	ID        uuid.UUID
	Email     string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type LoginUserResponse struct {
	ID    uuid.UUID
	Email string
}

type UserWithWallet struct {
	UserId        uuid.UUID `json:"user_id"`
	Email         string    `json:"email"`
	WalletId      uuid.UUID `json:"wallet_id"`
	WalletAddress string    `json:"wallet_address"`
	Avatar        string    `json:"avatar"`
	Name          string    `json:"name"`
}
