package postgres

import (
	"context"
	"mpc/internal/domain"
	sqlc "mpc/internal/infrastructure/db/sqlc"
	"mpc/internal/repository"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type userRepository struct {
	repository.BaseRepository
}

func NewUserRepo(dbPool *pgxpool.Pool) repository.UserRepository {
	return &userRepository{
		BaseRepository: repository.NewBaseRepo(dbPool),
	}
}

// Ensure UserRepository implements UserRepository
var _ repository.UserRepository = (*userRepository)(nil)

// GetUserWithWalletAddress implements repository.UserRepository.
func (r *userRepository) GetUserWithWallet(ctx context.Context, id uuid.UUID) (domain.UserWithWallet, error) {
	q := sqlc.New(r.DB())
	dbUser, err := q.GetUserWithWallet(ctx, pgtype.UUID{Bytes: id, Valid: true})
	if err != nil {
		return domain.UserWithWallet{}, err
	}

	return domain.UserWithWallet{
		UserId:        dbUser.ID.Bytes,
		Email:         dbUser.Email,
		WalletId:      dbUser.ID_2.Bytes,
		WalletAddress: dbUser.Address.String,
	}, nil
}

func (r *userRepository) CreateUser(ctx context.Context, params domain.CreateHashedUserParams) (domain.User, error) {
	var user domain.User
	err := r.WithTx(ctx, func(tx pgx.Tx) error {
		q := sqlc.New(tx)
		createdUser, err := q.CreateUser(ctx, sqlc.CreateUserParams{
			Email:        params.Email,
			PasswordHash: params.PasswordHash,
		})
		if err != nil {
			return err
		}
		user = domain.User{
			ID:           createdUser.ID.Bytes,
			Email:        createdUser.Email,
			PasswordHash: createdUser.PasswordHash,
			CreatedAt:    createdUser.CreatedAt.Time,
			UpdatedAt:    createdUser.UpdatedAt.Time,
		}
		return nil
	})
	return user, err
}

func (r *userRepository) GetUser(ctx context.Context, id uuid.UUID) (domain.User, error) {
	q := sqlc.New(r.DB())
	dbUser, err := q.GetUser(ctx, pgtype.UUID{Bytes: id, Valid: true})
	if err != nil {
		return domain.User{}, err
	}

	return domain.User{
		ID:        dbUser.ID.Bytes,
		Email:     dbUser.Email,
		CreatedAt: dbUser.CreatedAt.Time,
		UpdatedAt: dbUser.UpdatedAt.Time,
	}, nil
}

func (r *userRepository) GetUserByEmail(ctx context.Context, email string) (domain.User, error) {
	q := sqlc.New(r.DB())
	dbUser, err := q.GetUserByEmail(ctx, email)
	if err != nil {
		return domain.User{}, err
	}

	return domain.User{
		ID:           dbUser.ID.Bytes,
		Email:        dbUser.Email,
		PasswordHash: dbUser.PasswordHash,
		CreatedAt:    dbUser.CreatedAt.Time,
		UpdatedAt:    dbUser.UpdatedAt.Time,
	}, nil
}

func (r *userRepository) GetUserByAddress(ctx context.Context, address string) (domain.UserWithWallet, error) {
	q := sqlc.New(r.DB())
	dbUser, err := q.GetUserByAddress(ctx, address)
	if err != nil {
		return domain.UserWithWallet{}, err
	}

	return domain.UserWithWallet{
		UserId:        dbUser.ID.Bytes,
		Email:         dbUser.Email,
		Name:          dbUser.Name,
		WalletAddress: address,
		WalletId:      dbUser.ID_2.Bytes,
	}, nil
}

func (r *userRepository) UpdateUser(ctx context.Context, user domain.User) (domain.User, error) {
	err := r.WithTx(ctx, func(tx pgx.Tx) error {
		q := sqlc.New(tx)
		dbUser, err := q.UpdateUser(ctx, sqlc.UpdateUserParams{
			ID:           pgtype.UUID{Bytes: user.ID, Valid: true},
			Email:        user.Email,
			PasswordHash: user.PasswordHash,
		})
		if err != nil {
			return err
		}

		user = domain.User{
			ID:           dbUser.ID.Bytes,
			Email:        dbUser.Email,
			PasswordHash: dbUser.PasswordHash,
			CreatedAt:    dbUser.CreatedAt.Time,
			UpdatedAt:    dbUser.UpdatedAt.Time,
		}
		return nil
	})
	return user, err
}
