package http

import (
	"mpc/internal/delivery/http/handler"
	"mpc/internal/delivery/http/middleware"
	"mpc/internal/infrastructure/auth"
	"mpc/internal/usecase"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func NewRouter(
	userUC *usecase.UserUseCase,
	walletUC *usecase.WalletUseCase,
	txnUC *usecase.TxnUseCase,
	authUC *usecase.AuthUseCase,
	balanceUC *usecase.BalanceUseCase,
	jwtService *auth.JWTService,
	log *logrus.Logger,
) *gin.Engine {

	router := gin.Default()
	// router.Use(middleware.LoggerMiddleware(log))
	router.Use(middleware.CORS())
	router.Use(gin.Recovery())

	healthHandler := handler.NewHealthHandler()
	authHandler := handler.NewAuthHandler(*authUC)
	userHandler := handler.NewUserHandler(*userUC, *balanceUC)
	walletHandler := handler.NewWalletHandler(walletUC)
	txnHandler := handler.NewTxnHandler(*txnUC)
	balanceHanlder := handler.NewBalanceHandler(*balanceUC)

	v1 := router.Group("/api/v1")
	{
		health := v1.Group("/health")
		{
			health.GET("/", healthHandler.HealthCheck)
		}

		v1.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

		auth := v1.Group("/auth")
		{
			auth.POST("/login", authHandler.Login)
			auth.POST("/signup", authHandler.Signup)
			auth.POST("/logout", authHandler.Logout)
			auth.POST("/refresh", authHandler.Refresh)
		}

		users := v1.Group("/users")
		users.Use(middleware.AuthMiddleware(*jwtService))
		{
			users.GET("/profile", userHandler.GetUser)
		}

		wallets := v1.Group("/wallets")
		wallets.Use(middleware.AuthMiddleware(*jwtService))
		{
			wallets.POST("/", walletHandler.CreateWallet)
			wallets.GET("/:id", walletHandler.GetWallet)
		}

		transactions := v1.Group("/transactions")
		transactions.Use(middleware.AuthMiddleware(*jwtService))
		{
			transactions.GET("/", txnHandler.GetTransactions)
			transactions.POST("/", txnHandler.CreateAndSubmitTransaction)
			// transactions.POST("/create", txnHandler.CreateTransaction)
			// transactions.POST("/submit", txnHandler.SubmitTransaction)
		}
	
		balance := v1.Group("/balances")
		balance.Use(middleware.AuthMiddleware(*jwtService))
		{
			balance.GET("/", balanceHanlder.GetBalances)
		}
	}

	// Redirect to swagger docs
	router.GET("/docs", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/api/v1/swagger/index.html")
	})
	return router
}
