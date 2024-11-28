package usecase

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"log"
	"mpc/internal/domain"
	"mpc/internal/repository"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
)

type WalletUseCase interface {
	CreateWallet(ctx context.Context, userID uuid.UUID) (domain.Wallet, error)
	GetWallet(ctx context.Context, id uuid.UUID) (domain.Wallet, error)
	GetPrivateKey(ctx context.Context, userID uuid.UUID) (*ecdsa.PrivateKey, error)
}

type walletUseCase struct {
	walletRepo    repository.WalletRepository
	ethRepo       repository.EthereumRepository
	kafkaProducer *kafka.Writer
}

func NewWalletUC(walletRepo repository.WalletRepository, ethRepo repository.EthereumRepository, kafkaProducer *kafka.Writer) WalletUseCase {
	return &walletUseCase{
		walletRepo:    walletRepo,
		ethRepo:       ethRepo,
		kafkaProducer: kafkaProducer,
	}
}

var _ WalletUseCase = (*walletUseCase)(nil)

func (uc *walletUseCase) CreateWallet(ctx context.Context, userID uuid.UUID) (domain.Wallet, error) {
	privateKey, address, err := uc.ethRepo.CreateWallet()
	if err != nil {
		return domain.Wallet{}, err
	}

	// Convert private key to bytes
	privateKeyBytes := crypto.FromECDSA(privateKey)

	// Encrypt the private key
	encryptedPrivateKey, err := uc.ethRepo.EncryptPrivateKey(privateKeyBytes)
	if err != nil {
		return domain.Wallet{}, err
	}

	wallet := domain.CreateWalletParams{
		UserID:              userID,
		Address:             address.Hex(),
		EncryptedPrivateKey: encryptedPrivateKey,
	}

	newWallet, err := uc.walletRepo.CreateWallet(ctx, wallet)
	if err != nil {
		return domain.Wallet{}, err
	}

	// Publish wallet created event to Kafka topic wallet-created
	messageJSON, err := json.Marshal(newWallet)
	if err != nil {
		return domain.Wallet{}, err
	}
	fmt.Println("messageJSON: ", string(messageJSON))

	err2 := uc.kafkaProducer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(newWallet.ID.String()),
		Value: messageJSON,
	})
	if err2 != nil {
		log.Printf("Failed to publish message to Kafka: %v", err)
	} else {
		fmt.Println("Published message to Topic: ", uc.kafkaProducer.Topic)
	}

	return newWallet, nil
}

func (uc *walletUseCase) GetWallet(ctx context.Context, id uuid.UUID) (domain.Wallet, error) {
	return uc.walletRepo.GetWallet(ctx, id)
}

func (uc *walletUseCase) GetPrivateKey(ctx context.Context, userID uuid.UUID) (*ecdsa.PrivateKey, error) {
	wallet, err := uc.walletRepo.GetWalletByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	privateKeyBytes, err := uc.ethRepo.DecryptPrivateKey(wallet.EncryptedPrivateKey)
	if err != nil {
		return nil, err
	}
	return crypto.ToECDSA(privateKeyBytes)
}
