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

type transactionRepository struct {
	repository.BaseRepository
}

// DeleteTransaction implements repository.TransactionRepository.
func (r *transactionRepository) DeleteTransaction(ctx context.Context, tnx_hash string) (domain.Transaction, error) {
	q := sqlc.New(r.DB())
	transaction, err := q.DeleteTransaction(ctx, pgtype.Text{String: tnx_hash, Valid: true})
	if err != nil {
		return domain.Transaction{}, err
	}
	return domain.Transaction{
		ID:        transaction.ID.Bytes,
		WalletID:  transaction.WalletID.Bytes,
		ChainID:   transaction.ChainID.Bytes,
		ToAddress: transaction.ToAddress,
		Amount:    transaction.Amount,
		TokenID:   transaction.TokenID.Bytes,
		TxHash:    transaction.TxHash.String,
		GasPrice:  transaction.GasPrice.String,
		GasLimit:  transaction.GasLimit.String,
		Nonce:     transaction.Nonce.Int64,
		Status:    domain.Status(transaction.Status),
	} , nil
}


func NewTransactionRepo(dbPool *pgxpool.Pool) repository.TransactionRepository {
	return &transactionRepository{
		BaseRepository: repository.NewBaseRepo(dbPool),
	}
}

// Ensure TransactionRepository implements TransactionRepository
var _ repository.TransactionRepository = (*transactionRepository)(nil)

// GetPaginatedTransactions implements repository.TransactionRepository.
func (r *transactionRepository) GetPaginatedTransactions(ctx context.Context, userId uuid.UUID, tokenId uuid.UUID, page int, limit int) ([]domain.Transaction, error) {

	q := sqlc.New(r.DB())
	transactions, err := q.GetPaginatedTransactions(ctx, sqlc.GetPaginatedTransactionsParams{
		ID:      pgtype.UUID{Bytes: userId, Valid: true}, //userId
		TokenID: pgtype.UUID{Bytes: tokenId, Valid: true},
		Offset:  int32((page - 1) * limit),
		Limit:   int32(limit),
	})
	if err != nil {
		return nil, err
	}

	var result []domain.Transaction
	for _, t := range transactions {
		result = append(result, domain.Transaction{
			ID:          t.ID.Bytes,
			WalletID:    t.WalletID.Bytes,
			ChainID:     t.ChainID.Bytes,
			ToAddress:   t.ToAddress,
			FromAddress: t.FromAddress.String,
			Amount:      t.Amount,
			TokenID:     t.TokenID.Bytes,
			TxHash:      t.TxHash.String,
			GasPrice:    t.GasPrice.String,
			GasLimit:    t.GasLimit.String,
			Nonce:       t.Nonce.Int64,
			Status:      domain.Status(t.Status),
			CreatedAt:   t.CreatedAt.Time,
		})
	}
	return result, nil
}

// GetPaginatedAllTokenTransactions implements repository.TransactionRepository.
func (r *transactionRepository) GetPaginatedAllTokenTransactions(ctx context.Context, userId uuid.UUID, page int, limit int) ([]domain.Transaction, error) {
	q := sqlc.New(r.DB())
	transactions, err := q.GetPaginatedAllTokenTransactions(ctx, sqlc.GetPaginatedAllTokenTransactionsParams{
		ID:     pgtype.UUID{Bytes: userId, Valid: true}, //userId
		Offset: int32((page - 1) * limit),
		Limit:  int32(limit),
	})
	if err != nil {
		return nil, err
	}

	var result []domain.Transaction
	for _, t := range transactions {
		result = append(result, domain.Transaction{
			ID:          t.ID.Bytes,
			WalletID:    t.WalletID.Bytes,
			ChainID:     t.ChainID.Bytes,
			ToAddress:   t.ToAddress,
			FromAddress: t.FromAddress.String,
			Amount:      t.Amount,
			TokenID:     t.TokenID.Bytes,
			TxHash:      t.TxHash.String,
			GasPrice:    t.GasPrice.String,
			GasLimit:    t.GasLimit.String,
			Nonce:       t.Nonce.Int64,
			Status:      domain.Status(t.Status),
			CreatedAt:   t.CreatedAt.Time,
		})
	}
	return result, nil
}

// Insert transactions that is persist on blockchain
func (r *transactionRepository) InsertSettledTransactions(ctx context.Context, transactions []domain.Transaction) error {
	err := r.WithTx(ctx, func(tx pgx.Tx) error {
		q := sqlc.New(tx)
		for _, t := range transactions {
			_, err := q.InsertSetteledTransaction(ctx, sqlc.InsertSetteledTransactionParams{
				ID:          pgtype.UUID{Bytes: t.ID, Valid: true},
				WalletID:    pgtype.UUID{Bytes: t.WalletID, Valid: true},
				ChainID:     pgtype.UUID{Bytes: t.ChainID, Valid: true},
				ToAddress:   t.ToAddress,
				Amount:      t.Amount,
				FromAddress: pgtype.Text{String: t.FromAddress, Valid: true},
				TokenID:     pgtype.UUID{Bytes: t.TokenID, Valid: true},
				GasPrice:    pgtype.Text{String: t.GasPrice, Valid: true},
				GasLimit:    pgtype.Text{String: t.GasLimit, Valid: true},
				Nonce:       pgtype.Int8{Int64: t.Nonce, Valid: true},
				Status:      string(t.Status),
				TxHash:      pgtype.Text{String: t.TxHash, Valid: true},
			})
			if err != nil {
				return err
			}

		}
		return nil
	})
	return err
}

func (r *transactionRepository) CreateTransaction(ctx context.Context, params domain.CreateTransactionParams) (domain.Transaction, error) {

	var transaction domain.Transaction
	err := r.WithTx(ctx, func(tx pgx.Tx) error {
		q := sqlc.New(tx)
		createdTransaction, err := q.CreateTransaction(ctx, sqlc.CreateTransactionParams{
			ID:          pgtype.UUID{Bytes: params.ID, Valid: true},
			WalletID:    pgtype.UUID{Bytes: params.WalletID, Valid: true},
			ChainID:     pgtype.UUID{Bytes: params.ChainID, Valid: true},
			ToAddress:   params.ToAddress,
			Amount:      params.Amount,
			TokenID:     pgtype.UUID{Bytes: params.TokenID, Valid: true},
			GasPrice:    pgtype.Text{String: params.GasPrice, Valid: true},
			GasLimit:    pgtype.Text{String: params.GasLimit, Valid: true},
			Nonce:       pgtype.Int8{Int64: params.Nonce, Valid: true},
			Status:      string(params.Status),
			FromAddress: pgtype.Text{String: params.FromAddress, Valid: true},
		})
		if err != nil {
			return err
		}
		transaction = domain.Transaction{
			ID:        createdTransaction.ID.Bytes,
			WalletID:  createdTransaction.WalletID.Bytes,
			ChainID:   createdTransaction.ChainID.Bytes,
			ToAddress: createdTransaction.ToAddress,
			Amount:    createdTransaction.Amount,
			TokenID:   createdTransaction.TokenID.Bytes,
			GasPrice:  createdTransaction.GasPrice.String,
			GasLimit:  createdTransaction.GasLimit.String,
			Nonce:     createdTransaction.Nonce.Int64,
			Status:    domain.Status(createdTransaction.Status),
		}
		return nil
	})
	return transaction, err
}

func (r *transactionRepository) GetTransaction(ctx context.Context, id uuid.UUID) (domain.Transaction, error) {
	q := sqlc.New(r.DB())
	transaction, err := q.GetTransaction(ctx, pgtype.UUID{Bytes: id, Valid: true})
	if err != nil {
		return domain.Transaction{}, err
	}
	return domain.Transaction{
		ID:        transaction.ID.Bytes,
		WalletID:  transaction.WalletID.Bytes,
		ChainID:   transaction.ChainID.Bytes,
		ToAddress: transaction.ToAddress,
		Amount:    transaction.Amount,
		TokenID:   transaction.TokenID.Bytes,
		TxHash:    transaction.TxHash.String,
		GasPrice:  transaction.GasPrice.String,
		GasLimit:  transaction.GasLimit.String,
		Nonce:     transaction.Nonce.Int64,
		Status:    domain.Status(transaction.Status),
	}, nil
}

func (r *transactionRepository) UpdateTransaction(ctx context.Context, transaction domain.Transaction) error {
	q := sqlc.New(r.DB())
	_, err := q.UpdateTransaction(ctx, sqlc.UpdateTransactionParams{
		ID:       pgtype.UUID{Bytes: transaction.ID, Valid: true},
		Status:   string(transaction.Status),
		TxHash:   pgtype.Text{String: transaction.TxHash, Valid: transaction.TxHash != ""},
		GasPrice: pgtype.Text{String: transaction.GasPrice, Valid: transaction.GasPrice != ""},
		GasLimit: pgtype.Text{String: transaction.GasLimit, Valid: transaction.GasLimit != ""},
		Nonce:    pgtype.Int8{Int64: transaction.Nonce, Valid: true},
	})
	if err != nil {
		return err
	}
	return nil
}

func (r *transactionRepository) GetTransactions(ctx context.Context, userID uuid.UUID) ([]domain.Transaction, error) {
	// Implement the database operation here
	panic("not implemented")
}

func (r *transactionRepository) GetTransactionsByWalletID(ctx context.Context, walletID uuid.UUID) ([]domain.Transaction, error) {
	// Implement the database operation here
	panic("not implemented")
}

//
