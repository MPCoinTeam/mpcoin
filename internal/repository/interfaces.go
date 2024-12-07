package repository

import (
	"context"
	"crypto/ecdsa"
	"math/big"
	"mpc/internal/domain"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/google/uuid"
)

type UserRepository interface {
	CreateUser(ctx context.Context, params domain.CreateHashedUserParams) (domain.User, error)
	GetUser(ctx context.Context, id uuid.UUID) (domain.User, error)
	GetUserByEmail(ctx context.Context, email string) (domain.User, error)
	UpdateUser(ctx context.Context, user domain.User) (domain.User, error)
	GetUserWithWallet(ctx context.Context, id uuid.UUID) (domain.UserWithWallet, error)
	DBTransaction
}

type WalletRepository interface {
	CreateWallet(ctx context.Context, params domain.CreateWalletParams) (domain.Wallet, error)
	GetWallet(ctx context.Context, id uuid.UUID) (domain.Wallet, error)
	GetWalletByUserID(ctx context.Context, userID uuid.UUID) (domain.Wallet, error)
	GetWalletByAddress(ctx context.Context, address string) (domain.Wallet, error)
	GetWallets(ctx context.Context) ([]domain.Wallet, error)
	DBTransaction
}

type TransactionRepository interface {
	CreateTransaction(ctx context.Context, params domain.CreateTransactionParams) (domain.Transaction, error)
	GetTransaction(ctx context.Context, id uuid.UUID) (domain.Transaction, error)
	GetTransactionsByWalletID(ctx context.Context, walletID uuid.UUID) ([]domain.Transaction, error)
	GetPaginatedTransactions(ctx context.Context, userId uuid.UUID, tokenId uuid.UUID, page, limit int) ([]domain.Transaction, error)
	GetPaginatedAllTokenTransactions(ctx context.Context, userId uuid.UUID, page, limit int) ([]domain.Transaction, error)
	UpdateTransaction(ctx context.Context, transaction domain.Transaction) error
	InsertSettledTransactions(ctx context.Context, transactions []domain.Transaction) error
	DeleteTransaction(ctx context.Context, tx_hash string) (domain.Transaction, error)
	DBTransaction
}

type BalanceRepository interface {
	GetBalancesByUserId(ctx context.Context, userId uuid.UUID) ([]domain.GetBalanceResponse, error)
	UpdateBalance(ctx context.Context, params domain.UpdateBalanceParams) error
}
type ChainRepository interface {
	GetChains(ctx context.Context) ([]domain.Chain, error)
}

type EthereumRepository interface {
	CreateWallet() (*ecdsa.PrivateKey, common.Address, error)
	GetBalance(address common.Address) (*big.Int, error)
	CreateUnsignedTransaction(from common.Address, to common.Address, amount *big.Int) (*types.Transaction, error)
	SignTransaction(tx *types.Transaction, privateKey *ecdsa.PrivateKey) (*types.Transaction, error)
	SubmitTransaction(signedTx *types.Transaction) (common.Hash, error)
	WaitForTxn(hash common.Hash) (*types.Receipt, error)
	EncryptPrivateKey(data []byte) ([]byte, error)
	DecryptPrivateKey(ciphertext []byte) ([]byte, error)
	GetTransactionsStartFrom(blockNumber uint64) ([]domain.Transaction, error)
	GetTransactionsInBlock(blockNumber uint64) ([]domain.Transaction, error)
	SubscribeNewHead(ctx context.Context, ch chan<- *types.Header) (ethereum.Subscription, error)
	GetTransactionReceipt(context context.Context, txHash common.Hash) (*types.Receipt, error)
	NewInstance(url string) (EthereumRepository, error)
}
