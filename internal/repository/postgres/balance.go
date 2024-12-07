package postgres

import (
	"context"
	"math/big"
	"mpc/internal/domain"
	sqlc "mpc/internal/infrastructure/db/sqlc"
	"mpc/internal/repository"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type balanceRepository struct {
	repository.BaseRepository
}

// UpdateBalance implements repository.BalanceRepository.
func (b *balanceRepository) UpdateBalance(ctx context.Context, params domain.UpdateBalanceParams) error {
	q := sqlc.New(b.DB())
	_, err := q.UpdateBalance(ctx, sqlc.UpdateBalanceParams{
		Address: params.Address.String(),
		TokenID: pgtype.UUID{Bytes: params.TokenID, Valid: true},
		Balance: params.Balance.Int64(),
	})
	if err != nil {
		return err
	}

	return nil
}

// GetBalance implements repository.BalanceRepository.
func (b *balanceRepository) GetBalancesByUserId(ctx context.Context, userId uuid.UUID) ([]domain.GetBalanceResponse, error) {
	q := sqlc.New(b.DB())
	balances, err := q.GetBalancesByUserId(ctx, pgtype.UUID{Bytes: userId, Valid: true})
	if err != nil {
		return []domain.GetBalanceResponse{}, err
	}

	var result []domain.GetBalanceResponse
	for _, b := range balances {
		result = append(result, domain.GetBalanceResponse{
			TokenID: b.ID.Bytes,
			Balance: *big.NewInt(b.Balance),
			ChainID: b.ChainID.Bytes,
			ContractAddress: b.ContractAddress,
			UpdatedAt:       b.UpdatedAt.Time,
			TokenName:       b.Name,
			TokenSymbol:     b.Symbol,
			Decimals:        int64(b.Decimals),
		})
	}

	return result, nil

}

func NewBalanceRepo(dbPool *pgxpool.Pool) repository.BalanceRepository {
	return &balanceRepository{
		BaseRepository: repository.NewBaseRepo(dbPool),
	}
}

var _ repository.BalanceRepository = (*balanceRepository)(nil)
