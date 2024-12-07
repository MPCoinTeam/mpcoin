// Package usecase contains the business logic for the application.
package usecase

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"math/big"
	"mpc/internal/domain"
	"mpc/internal/infrastructure/redis"
	"mpc/internal/repository"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
)

type TxnUseCase interface {
	CreateTransaction(ctx context.Context, userID uuid.UUID, params domain.CreateTxnRequest) (uuid.UUID, error)
	SubmitTransaction(ctx context.Context, userId uuid.UUID, txnId uuid.UUID) (domain.Transaction, error)
	GetTransactions(ctx context.Context, walletID uuid.UUID) ([]domain.Transaction, error)
	GetPaginatedTransactions(ctx context.Context, userId uuid.UUID, tokenId uuid.UUID, page, perPage int) ([]domain.Transaction, error)
}

type txnUseCase struct {
	txnRepo       repository.TransactionRepository
	ethRepo       repository.EthereumRepository
	walletUC      WalletUseCase
	redisClient   redis.RedisClient
	kafkaProducer *kafka.Writer
}

func NewTxnUC(txnRepo repository.TransactionRepository, ethRepo repository.EthereumRepository, walletUC WalletUseCase, redisClient redis.RedisClient, kafkaProducer *kafka.Writer) TxnUseCase {
	return &txnUseCase{txnRepo: txnRepo, ethRepo: ethRepo, walletUC: walletUC, redisClient: redisClient, kafkaProducer: kafkaProducer}
}

var _ TxnUseCase = (*txnUseCase)(nil)

// GetPaginatedTransactions implements TxnUseCase.
func (uc *txnUseCase) GetPaginatedTransactions(ctx context.Context, userId uuid.UUID, tokenId uuid.UUID, page int, perPage int) ([]domain.Transaction, error) {
	
	if(userId == uuid.Nil){
		return nil, errors.New("address cannot be nil")
	}
	if(page < 1){
		return nil, errors.New("page must be greater than 0")
	}
	if(perPage < 1){
		return nil, errors.New("perPage must be greater than 0")
	}

	var txnList []domain.Transaction
	var err error

	if(tokenId == uuid.Nil){
		txnList, err = uc.txnRepo.GetPaginatedAllTokenTransactions(ctx, userId, page, perPage)
	} else {
		txnList, err = uc.txnRepo.GetPaginatedTransactions(ctx, userId, tokenId, page, perPage)
	}

	if err != nil {
		return nil, fmt.Errorf("transactionUsecase.GetPaginatedTransactions: failed to call txnRepo: %w", err)
	}

	if txnList == nil {
		txnList = []domain.Transaction{}
	}

	return txnList, err
}

// CreateTransaction creates a new transaction and stores it in the database.
func (uc *txnUseCase) CreateTransaction(ctx context.Context, userID uuid.UUID, params domain.CreateTxnRequest) (uuid.UUID, error) {
	wallet, err := uc.walletUC.GetWallet(ctx, params.WalletID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to get wallet: %w", err)
	}

	fromAddress := common.HexToAddress(wallet.Address)
	toAddress := common.HexToAddress(params.ToAddress)

	// Convert amount string to big.Float first, then to big.Int
	amountFloat, _, err := new(big.Float).Parse(params.Amount, 10)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid amount: %s", params.Amount)
	}

	// Multiply by 1e18 to convert to Wei
	weiFloat := new(big.Float).Mul(amountFloat, big.NewFloat(1e18))

	// Convert to big.Int, truncating any fractional part
	amountInWei, _ := weiFloat.Int(nil)

	unsignedTx, err := uc.ethRepo.CreateUnsignedTransaction(fromAddress, toAddress, amountInWei)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to create unsigned transaction: %w", err)
	}

	// Print useful information about the unsigned transaction
	fmt.Printf("CreateTransaction: unsignedTx details:\n")
	fmt.Printf("  To: %s\n", unsignedTx.To().Hex())
	fmt.Printf("  Value: %s\n", unsignedTx.Value().String())
	fmt.Printf("  Gas: %d\n", unsignedTx.Gas())
	fmt.Printf("  GasPrice: %s\n", unsignedTx.GasPrice().String())
	fmt.Printf("  Nonce: %d\n", unsignedTx.Nonce())

	// Serialize the unsigned transaction
	unsignedTxData, err := unsignedTx.MarshalBinary()
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to serialize unsigned transaction: %w", err)
	}

	// Encrypt the unsigned transaction data
	encryptedData, err := uc.encryptData(unsignedTxData)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to encrypt transaction data: %w", err)
	}

	txID := uuid.New()
	err = uc.redisClient.Set(ctx, fmt.Sprintf("transaction:%s", txID), encryptedData, 24*time.Hour)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to save encrypted transaction to Redis: %w", err)
	}

	transaction := domain.CreateTransactionParams{
		ID:        txID,
		WalletID:  params.WalletID,
		ChainID:   params.ChainID,
		Amount:    amountInWei.String(),
		ToAddress: params.ToAddress,
		TokenID:   params.TokenID,
		Status:    domain.StatusPending,
		FromAddress: wallet.Address,
	}

	_, err = uc.txnRepo.CreateTransaction(ctx, transaction)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to save transaction to database: %w", err)
	}

	return txID, nil
}

// SubmitTransaction signs and submits a transaction to the Ethereum network.
func (uc *txnUseCase) SubmitTransaction(ctx context.Context, userId uuid.UUID, txnId uuid.UUID) (domain.Transaction, error) {
	// Retrieve the encrypted transaction from Redis
	encryptedData, err := uc.redisClient.Get(ctx, fmt.Sprintf("transaction:%s", txnId))
	if err != nil {
		return domain.Transaction{}, fmt.Errorf("failed to get transaction from Redis: %w", err)
	}

	// Decrypt the transaction data
	unsignedTxData, err := uc.decryptData(encryptedData)
	if err != nil {
		return domain.Transaction{}, fmt.Errorf("failed to decrypt transaction data: %w", err)
	}

	var unsignedTx types.Transaction
	err = unsignedTx.UnmarshalBinary(unsignedTxData)
	if err != nil {
		return domain.Transaction{}, fmt.Errorf("failed to deserialize unsigned transaction: %w", err)
	}

	// Get private key from user
	privateKey, err := uc.walletUC.GetPrivateKey(ctx, userId)
	if err != nil {
		return domain.Transaction{}, fmt.Errorf("failed to get private key: %w", err)
	}

	signedTx, err := uc.ethRepo.SignTransaction(&unsignedTx, privateKey)
	if err != nil {
		return uc.updateTransactionStatus(ctx, txnId, domain.StatusFailed, err)
	}

	txHash, err := uc.ethRepo.SubmitTransaction(signedTx)
	if err != nil {
		return uc.updateTransactionStatus(ctx, txnId, domain.StatusFailed, err)
	}

	transaction, err := uc.txnRepo.GetTransaction(ctx, txnId)
	if err != nil {
		return domain.Transaction{}, fmt.Errorf("failed to get transaction from database: %w", err)
	}

	transaction.Status = domain.StatusSubmitted
	transaction.TxHash = txHash.Hex()

	if err := uc.txnRepo.UpdateTransaction(ctx, transaction); err != nil {
		return domain.Transaction{}, fmt.Errorf("failed to update transaction in database: %w", err)
	}

	if err := uc.redisClient.Delete(ctx, fmt.Sprintf("transaction:%s", txnId.String())); err != nil {
		log.Printf("Failed to delete submitted transaction from Redis: %v", err)
	}

	uc.publishMessage(ctx, txnId, transaction.ChainID, txHash)

	return transaction, nil
}

// GetTransactions retrieves all transactions for a given user ID.
func (uc *txnUseCase) GetTransactions(ctx context.Context, walletID uuid.UUID) ([]domain.Transaction, error) {
	return uc.txnRepo.GetTransactionsByWalletID(ctx, walletID)
}

func (uc *txnUseCase) updateTransactionStatus(ctx context.Context, id uuid.UUID, status domain.Status, err error) (domain.Transaction, error) {
	transaction, dbErr := uc.txnRepo.GetTransaction(ctx, id)
	if dbErr != nil {
		return domain.Transaction{}, fmt.Errorf("failed to get transaction: %w (original error: %v)", dbErr, err)
	}

	transaction.Status = status

	if dbErr := uc.txnRepo.UpdateTransaction(ctx, transaction); dbErr != nil {
		return domain.Transaction{}, fmt.Errorf("failed to update transaction status: %w (original error: %v)", dbErr, err)
	}

	return transaction, err
}

// Encryption helper function
func (uc *txnUseCase) encryptData(data []byte) (string, error) {
	key := []byte("5ca44e2f52f9418f9b54e62f94d97f65") // Replace with a secure key
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decryption helper function
func (uc *txnUseCase) decryptData(encryptedData string) ([]byte, error) {
	key := []byte("5ca44e2f52f9418f9b54e62f94d97f65") // Replace with the same secure key used for encryption
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedData)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}

func (uc *txnUseCase) publishMessage(ctx context.Context, txnId uuid.UUID, chainID uuid.UUID, txHash common.Hash) {
	// Prepare the message
	message := domain.TxnMessage{
		ChainID: chainID,
		TxHash:  txHash.Hex(),
	}

	// Serialize the message to JSON
	messageJSON, err := json.Marshal(message)
	if err != nil {
		log.Printf("Failed to marshal transaction message: %v", err)
	}

	// Publish message to Kafka
	if err := uc.kafkaProducer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(txnId.String()),
		Value: messageJSON,
	}); err != nil {
		log.Printf("Failed to publish message to Kafka: %v", err)
	}
}
