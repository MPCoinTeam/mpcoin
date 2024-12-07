package handler

import (
	"mpc/internal/domain"
	"mpc/internal/usecase"
	"mpc/pkg/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type TxnHandler struct {
	txnUC usecase.TxnUseCase
}

func NewTxnHandler(txnUC usecase.TxnUseCase) *TxnHandler {
	return &TxnHandler{txnUC: txnUC}
}

// GetTransactions godoc
// @Summary Get Transactions
// @Description Get a list of transactions
// @Tags transaction
// @Accept json
// @Produce json
// @Param page query int false "Page number"
// @Param per_page query int false "Number of items per page"
// @Param token query string false "Token ID"
// @Success 200 {object} docs.GetTxnResponse "Successful response"
// @Failure 400 {string} string "Bad request error due to invalid input"
// @Failure 500 {string} string "Internal server error"
// @Router /transactions [get]
// @Security ApiKeyAuth
func (h *TxnHandler) GetTransactions(c *gin.Context) {

	userId := c.MustGet("userID").(uuid.UUID)
	if userId == uuid.Nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}
	tokenId, err := uuid.Parse(c.Query("token"))
	if err != nil {
		// utils.ErrorResponse(c, http.StatusBadRequest, "Invalid token parameter")
		// return
		tokenId = uuid.Nil
	}
	// Get page, per_page and adress query parameters
	page, err := strconv.Atoi(c.DefaultQuery("page", "1")) // Default to page 1
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid page parameter")
		return
	}
	perPage, err := strconv.Atoi(c.DefaultQuery("per_page", "10")) // Default to 10 items per page
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid per_page parameter")
		return
	}

	// Get transactions
	txnList, err := h.txnUC.GetPaginatedTransactions(c.Request.Context(), userId, tokenId, page, perPage)
	
	

	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get transactions: "+err.Error())
		return
	}
	// Send a successful response with the transaction list
	utils.SuccessResponse(c, http.StatusOK, gin.H{"transactions": txnList, "page": page, "per_page": perPage})
}

func (h *TxnHandler) GetDetail(c *gin.Context) {
	// Retrieve the "id" path parameter from the URL
	id := c.Param("id")

	// Use the id as needed
	c.JSON(200, gin.H{
		"id": id,
	})
}

// CreateTransaction godoc
// @Summary Create Transaction
// @Description Create a new transaction
// @Tags transaction
// @Accept json
// @Produce json
// @Param createTxnRequest body domain.CreateTxnRequest true "Create Transaction Request"
// @Success 201 {object} docs.CreateTxnResponse "Successful response"
// @Failure 400 {string} string "Bad request error due to invalid input"
// @Failure 401 {string} string "Unauthorized error due to invalid token"
// @Failure 500 {string} string "Internal server error"
// @Router /transactions/create [post]
// @Security ApiKeyAuth
func (h *TxnHandler) CreateTransaction(c *gin.Context) {
	// Get userID from auth middleware
	userIDInterface, exists := c.Get("userID")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	req, err := utils.ParseRequest[domain.CreateTxnRequest](c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	// Parse userID to uuid.UUID
	userID, ok := userIDInterface.(uuid.UUID)
	if !ok {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Invalid user ID type")
		return
	}

	// Create transaction
	txnID, err := h.txnUC.CreateTransaction(c.Request.Context(), userID, req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create transaction: "+err.Error())
		return
	}
	utils.SuccessResponse(c, http.StatusCreated, gin.H{
		"message": "Transaction created successfully",
		"txn_id":  txnID,
	})
}

// SubmitTransaction godoc
// @Summary Submit Transaction
// @Description Submit a transaction
// @Tags transaction
// @Accept json
// @Produce json
// @Param submitTxnRequest body domain.SubmitTxnRequest true "Submit Transaction Request"
// @Success 200 {object} docs.SubmitTnxResponse "Successful response"
// @Failure 400 {string} string "Bad request error due to invalid input"
// @Failure 401 {string} string "Unauthorized error due to invalid token"
// @Failure 500 {string} string "Internal server error"
// @Router /transactions/submit [post]
// @Security ApiKeyAuth
func (h *TxnHandler) SubmitTransaction(c *gin.Context) {
	userIDInterface, exists := c.Get("userID")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// Parse userID to uuid.UUID
	userID, ok := userIDInterface.(uuid.UUID)
	if !ok {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Invalid user ID type")
		return
	}

	req, err := utils.ParseRequest[domain.SubmitTxnRequest](c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	txn, err := h.txnUC.SubmitTransaction(c.Request.Context(), userID, req.ID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to submit transaction: "+err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusOK, gin.H{"message": "Transaction submitted", "tx_hash": txn.TxHash})
}

// CreateAndSubmitTransaction godoc
// @Summary Create and Submit Transaction
// @Description Create a new transaction and submit it
// @Tags transaction
// @Accept json
// @Produce json
// @Param createTxnRequest body domain.CreateTxnRequest true "Create Transaction Request"
// @Success 201 {object} docs.CreateTxnResponse "Successful response"
// @Failure 400 {string} string "Bad request error due to invalid input"
// @Failure 401 {string} string "Unauthorized error due to invalid token"
// @Failure 500 {string} string "Internal server error"
// @Router /transactions [post]
// @Security ApiKeyAuth
func (h *TxnHandler) CreateAndSubmitTransaction(c *gin.Context) {
	// Get userID from auth middleware
	userIDInterface, exists := c.Get("userID")
	if !exists {
		utils.ErrorResponse(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	req, err := utils.ParseRequest[domain.CreateTxnRequest](c)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	// Parse userID to uuid.UUID
	userID, ok := userIDInterface.(uuid.UUID)
	if !ok {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Invalid user ID type")
		return
	}

	// Create transaction
	txnID, err := h.txnUC.CreateTransaction(c.Request.Context(), userID, req)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create transaction: "+err.Error())
		return
	}

	// Immediately submit it
	txn, err := h.txnUC.SubmitTransaction(c.Request.Context(), userID, txnID)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to submit transaction: "+err.Error())
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, gin.H{"message": "Transaction created and submitted", "tx_hash": txn.TxHash})
}
