// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0

package db

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type Balance struct {
	ID        pgtype.UUID
	WalletID  pgtype.UUID
	ChainID   pgtype.UUID
	TokenID   pgtype.UUID
	Balance   pgtype.Numeric
	UpdatedAt pgtype.Timestamptz
}

type Chain struct {
	ID             pgtype.UUID
	Name           string
	ChainID        string
	RpcUrl         string
	NativeCurrency string
	CreatedAt      pgtype.Timestamptz
	UpdatedAt      pgtype.Timestamptz
	ExplorerUrl    pgtype.Text
}

type Token struct {
	ID              pgtype.UUID
	ChainID         pgtype.UUID
	ContractAddress string
	Name            string
	Symbol          string
	Decimals        int32
	CreatedAt       pgtype.Timestamptz
	UpdatedAt       pgtype.Timestamptz
}

type Transaction struct {
	ID        pgtype.UUID
	WalletID  pgtype.UUID
	ChainID   pgtype.UUID
	ToAddress string
	Amount    string
	TokenID   pgtype.UUID
	GasPrice  pgtype.Text
	GasLimit  pgtype.Text
	Nonce     pgtype.Int8
	Status    string
	TxHash    pgtype.Text
	CreatedAt pgtype.Timestamptz
	UpdatedAt pgtype.Timestamptz
}

type User struct {
	ID           pgtype.UUID
	Email        string
	PasswordHash string
	CreatedAt    pgtype.Timestamptz
	UpdatedAt    pgtype.Timestamptz
}

type Wallet struct {
	ID                  pgtype.UUID
	UserID              pgtype.UUID
	Address             string
	EncryptedPrivateKey []byte
	CreatedAt           pgtype.Timestamptz
	UpdatedAt           pgtype.Timestamptz
}
