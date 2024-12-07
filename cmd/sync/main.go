package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"

	"mpc/internal/domain"
	"mpc/internal/infrastructure/config"
	"mpc/internal/infrastructure/ethereum"
	"mpc/internal/repository"
	"mpc/internal/repository/postgres"
	"mpc/internal/usecase"

	customeKafka "mpc/internal/infrastructure/kafka"

	db "mpc/internal/infrastructure/db"
)

var (
	// wsApiKey = os.Getenv("WEBSOCKET_API_KEY")
	walletAddressMapId = make(map[string]uuid.UUID)
    // mu      sync.Mutex
)

func StartKafkaReaderTopicWalletCreated(cfg *config.Config) {
	reader, err := customeKafka.NewKafkaConsumer(cfg, customeKafka.WithTopic(cfg.Kafka.WalletCreatedTopic))
	if err != nil {
		log.Printf("Failed to initialize Kafka consumer: %v", err)
	}
	defer reader.Close()

	if err = reader.SetOffset(kafka.LastOffset); err != nil {
		log.Println("error listening to wallet-created topic:", err)
		return
	}

	log.Println("Start listening to wallet-created topic")

	for {
		m, err := reader.ReadMessage(context.Background())
		if err != nil {
			log.Print(err)
			log.Printf("Error reading kafka message: %v\n", err)
			continue
		}

		//Handle message
		walletCreatedEventHandle(string(m.Value))
	}
}
func StartKafkaReaderTopicTransactionFound(cfg *config.Config, balanceUC usecase.BalanceUseCase) {
	reader, err := customeKafka.NewKafkaConsumer(cfg, customeKafka.WithTopic(cfg.Kafka.TransactionFoundTopic))
	if err != nil {
		log.Printf("Failed to initialize Kafka consumer: %v", err)
	}
	defer reader.Close()

	if err = reader.SetOffset(kafka.LastOffset); err != nil {
		log.Println("error listening to transaction-found topic:", err)
		return
	}

	log.Println("Start listening to transaction-found topic")

	for {
		m, err := reader.ReadMessage(context.Background())
		if err != nil {
			log.Print(err)
			log.Printf("Error reading kafka message: %v\n", err)
			continue
		}
		//Handle message
		transactionFoundHandle(string(m.Value), balanceUC)
	}
}

func walletCreatedEventHandle(message string){
	log.Printf("New wallet created: %s\n", message)
	var data domain.Wallet
    err := json.Unmarshal([]byte(message), &data)
    if err != nil {
        log.Println("Error unmarshalling JSON:", err)
        return
    }
	//update wallet address hash table
	// mu.Lock()
	walletAddressMapId[data.Address] = data.ID
	// mu.Unlock()
}

func transactionFoundHandle(message string, balanceUC usecase.BalanceUseCase){
	//update balance
	var data domain.Transaction
	err := json.Unmarshal([]byte(message), &data)
	if err != nil {
		log.Println("Error unmarshalling JSON:", err)
		return
	}
	
	hasFrom := walletAddressMapId[data.FromAddress]
	hasTo := walletAddressMapId[data.ToAddress]

	if hasFrom != uuid.Nil {
		err = balanceUC.UpdateBalanceRPC(context.Background(), common.HexToAddress(data.FromAddress), data.TokenID)
		if err != nil {
			log.Println("Error updating balance:", err)
			return
		}
	}	
	if hasTo != uuid.Nil {
		err = balanceUC.UpdateBalanceRPC(context.Background(), common.HexToAddress(data.ToAddress), data.TokenID)
		if err != nil {
			log.Println("Error updating balance:", err)
			return
		}
	}
}

// lọc ra các transaction của user và persist vào db
func persistUsersTransactions(ctx context.Context, transactions []domain.Transaction, ethRepo repository.EthereumRepository, tnxRepo repository.TransactionRepository, chain domain.Chain, tnxFoundPublisher *customeKafka.Writer) error {
	usersTransactions := []domain.Transaction{}
	for _, tnx := range transactions {

		if(walletAddressMapId[tnx.ToAddress] == uuid.Nil && walletAddressMapId[tnx.FromAddress] == uuid.Nil){
			continue
		}

		tnx.ChainID = chain.ID
		tnx.TokenID = chain.NativeTokenID	
		//publish transaction found event
		messageJSON, err := json.Marshal(tnx)
		if err != nil {
			log.Println("Error marshalling JSON:", err)
		}
		err = tnxFoundPublisher.WriteMessages(ctx, kafka.Message{
			Key:   []byte(tnx.ID.String()),
			Value: messageJSON,
		})
		if err != nil {
			log.Printf("Failed to publish message to Kafka: %v", err)
		}else{
			log.Println("Published transaction found event to Topic: ", tnxFoundPublisher.Topic)
		}

		if(walletAddressMapId[tnx.ToAddress] == uuid.Nil && walletAddressMapId[tnx.FromAddress] == uuid.Nil){	
			continue
		}

		log.Printf("Transaction: %s\n", tnx.TxHash)
		log.Printf("From: %s\n", tnx.FromAddress)
		log.Printf("To: %s\n", tnx.ToAddress)

		tnxReceipt, err := ethRepo.GetTransactionReceipt(ctx, common.HexToHash(tnx.TxHash))
		if err != nil {
			log.Printf("Error getting transaction receipt: %v\n", err)
			return err
		}
		if(tnxReceipt.Status == 0){
			tnx.Status = domain.StatusFailed
		}else{
			tnx.Status = domain.StatusSuccess
		}
		
		tnx.WalletID = walletAddressMapId[tnx.ToAddress]
		usersTransactions = append(usersTransactions, tnx)
	}

	if len(usersTransactions) == 0 {
		return nil
	}

	//delete if exist
	for _, tnx := range usersTransactions {
		fmt.Print("Deleting transaction: ", tnx.TxHash)
		tnx2, err := tnxRepo.DeleteTransaction(ctx, tnx.TxHash)
		if err != nil {
			log.Printf("Error deleting transaction: %v\n", err)
		}
		log.Printf("Deleted transaction: %s\n", tnx.TxHash)
		
		//find index of tnx in usersTransactions
		index := -1
		for i, t := range usersTransactions {
			if t.TxHash == tnx2.TxHash {
				index = i
				break
			}
		}
		if index != -1 {
			usersTransactions[index].CreatedAt = tnx.CreatedAt
		}
	}

	
	err := tnxRepo.InsertSettledTransactions(ctx, usersTransactions)
	if err != nil {
		log.Printf("Error inserting transactions: %v\n", err)
		return err
	}

	return nil
}

func SyncChainData(	ctx context.Context, chain domain.Chain, ethRepo repository.EthereumRepository, wsEthRepo repository.EthereumRepository, tnxRepo repository.TransactionRepository, tnxFoundPublisher *customeKafka.Writer) error {
	// Initial scan from start block to the latest block
	if chain.LastScanBlock == -1 {
		log.Printf("Chain %s added.\n", chain.Name)
	} else {
		go func(blockNum uint64, ethRepo repository.EthereumRepository, tnxRepo repository.TransactionRepository, chain domain.Chain, tnxFoundPublisher *customeKafka.Writer) {
			log.Printf("Chain %s is scanning from block %d to end\n", chain.Name, chain.LastScanBlock)
			transactions, err := ethRepo.GetTransactionsStartFrom(blockNum)
			if err != nil {
				log.Printf("Error getting transactions: %v\n", err)
				return
			}

			err = persistUsersTransactions(ctx, transactions, ethRepo, tnxRepo, chain, tnxFoundPublisher)
			if err != nil {
				log.Printf("Error persisting transactions: %v\n", err)
				return
			}
		}(uint64(chain.LastScanBlock), ethRepo, tnxRepo, chain, tnxFoundPublisher)
	}

	log.Printf("Setting up WebSocket connection for chain: %s\n", chain.Name)
	headers := make(chan *types.Header)
	sub, err := wsEthRepo.SubscribeNewHead(context.Background(), headers)
	if err != nil {
		log.Printf("Failed to subscribe to new head: %v", err)
	}

	// Lắng nghe sự kiện khi có block mới
	log.Println("Listening for new blocks...")
	for {
		select {
		case <-ctx.Done():
			log.Printf("Context canceled, stopping block listener for chain %s\n", chain.Name)
			return nil
		case err := <-sub.Err():
			log.Printf("Subscription error on chain %s: %v. Retrying...\n", chain.Name, err)
			time.Sleep(5 * time.Second) // Retry delay before reconnecting
			sub, err = wsEthRepo.SubscribeNewHead(ctx, headers)
			if err != nil {
				log.Printf("Failed to resubscribe to new head for chain %s: %v", chain.Name, err)
				continue
			}
		case header := <-headers:
			go func(header *types.Header, ethRepo repository.EthereumRepository, tnxRepo repository.TransactionRepository, chain domain.Chain, tnxFoundPublisher *customeKafka.Writer) {
				log.Printf("Chain %s: New block %d\n", chain.Name, header.Number.Uint64())
				transactions, err := ethRepo.GetTransactionsInBlock(header.Number.Uint64())
				if err != nil {
					log.Printf("Error getting transactions for block %d on chain %s: %v\n", header.Number.Uint64(), chain.Name, err)
					return
				}

				err = persistUsersTransactions(ctx, transactions,ethRepo, tnxRepo, chain, tnxFoundPublisher)
				if err != nil {
					log.Printf("Error persisting transactions: %v\n", err)
					return
				}
			}(header, wsEthRepo, tnxRepo, chain, tnxFoundPublisher)
		}
	}
}

//to-do: add arguments hanlder to run specific chain sync job
func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Printf("Failed to load config: %v", err)
		return
	}

	dbPool, err := db.InitDB(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.CloseDB()

	transactionRepo := postgres.NewTransactionRepo(dbPool)
	walletRepo := postgres.NewWalletRepo(dbPool)
	chainRepo := postgres.NewChainRepo(dbPool)
	balanceRepo := postgres.NewBalanceRepo(dbPool)
	balanceUC := usecase.NewBalanceUC(balanceRepo, walletRepo, chainRepo, &ethereum.EthereumClient{})


	// kafka
	transactionFoundPublisher, err := customeKafka.NewKafkaProducer(cfg, customeKafka.WithTopic(cfg.Kafka.TransactionFoundTopic))
	if err != nil {
		log.Printf("Failed to initialize Kafka producer: %v", err)
	}
	defer transactionFoundPublisher.Close()
	go StartKafkaReaderTopicWalletCreated(cfg)
	go StartKafkaReaderTopicTransactionFound(cfg, balanceUC)

	chains, err := chainRepo.GetChains(context.Background())
	if err != nil {
		log.Printf("Failed to get chains: %v", err)
		return
	}
	wallets, err := walletRepo.GetWallets(context.Background())
	if err != nil {
		log.Printf("Failed to get chains: %v", err)
		return
	}

	for _, wallet := range wallets {
		walletAddressMapId[wallet.Address] = wallet.ID
	}

	var wg sync.WaitGroup
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for _, chain := range chains {
		wg.Add(1)
		ethRepo, err := ethereum.NewEthereumClient(chain.RPCURL, cfg.Ethereum.SecretKey)
		if err != nil {
			log.Printf("Chain %s: Failed to initialize Ethereum client: %v", chain.Name, err)
			wg.Done()
			continue
		}
		wsEthRepo, err := ethereum.NewEthereumClient(chain.WSURL, cfg.Ethereum.SecretKey)
		if err != nil {
			log.Printf("Chain %s: Failed to initialize Ethereum client: %v", chain.Name, err)
			wg.Done()
			continue
		}
		
		go func(chain domain.Chain, ethRepo repository.EthereumRepository, tnxRepo repository.TransactionRepository, tnxFoundPublisher *customeKafka.Writer) {
			defer wg.Done()
			for{
				log.Printf("Sync: Starting sync job for chain: %s\n", chain.Name)
				err := SyncChainData(ctx, chain, ethRepo,wsEthRepo, tnxRepo, tnxFoundPublisher)
				if err != nil {
					log.Printf("Sync: Error syncing chain %s: %v\n", chain.Name, err)
				}
				//sleep
				time.Sleep(5 * time.Second)
			}
			
		}(chain, ethRepo, transactionRepo, transactionFoundPublisher)
	}
	wg.Wait()
}

