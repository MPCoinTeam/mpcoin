package handler

import (
	"mpc/internal/usecase"
	"mpc/pkg/utils"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UserHandler struct {
	userUseCase usecase.UserUseCase
}

func NewUserHandler(userUseCase usecase.UserUseCase) *UserHandler {
	return &UserHandler{userUseCase: userUseCase}
}

// GetUser godoc
// @Summary Get user profile
// @Description Get user profile
// @Tags users
// @Accept json
// @Produce json
// @Success 200 {object} docs.GetUserWalletResponse "Successful response"
// @Failure 401 {string} string "Unauthorized error"
// @Failure 500 {string} string "Internal server error"
// @Security ApiKeyAuth
// @Router /users/profile [get]
func (h *UserHandler) GetUser(c *gin.Context) {
	//get user id from context
	userId := c.MustGet("userID").(uuid.UUID)
	if userId == uuid.Nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	//get user
	user, err := h.userUseCase.GetUserWallet(c, userId)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get user")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, gin.H{"user": user})
}

func (h *UserHandler) GetUserQuery(c *gin.Context) {
	//get user id from context
	address := c.Query("address")
	if address == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Address not found"})
		return
	}

	//get user
	user, err := h.userUseCase.GetUserByAddress(c, address)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to get user")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, gin.H{"user": user})
}
