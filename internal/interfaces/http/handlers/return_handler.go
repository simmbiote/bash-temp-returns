package handlers

import (
	"net/http"

	"customer-support-api/internal/application/dto"
	"customer-support-api/internal/application/usecases"

	"github.com/gin-gonic/gin"
)

type ReturnHandler struct {
	createReturnUseCase *usecases.CreateReturnRequestUseCase
	getReturnUseCase    *usecases.GetReturnRequestUseCase
	listReturnsUseCase  *usecases.ListReturnRequestsUseCase
}

func NewReturnHandler(
	createReturnUseCase *usecases.CreateReturnRequestUseCase,
	getReturnUseCase *usecases.GetReturnRequestUseCase,
	listReturnsUseCase *usecases.ListReturnRequestsUseCase,
) *ReturnHandler {
	return &ReturnHandler{
		createReturnUseCase: createReturnUseCase,
		getReturnUseCase:    getReturnUseCase,
		listReturnsUseCase:  listReturnsUseCase,
	}
}

func (h *ReturnHandler) CreateReturn(c *gin.Context) {
	var req dto.CreateReturnRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	customerID := c.GetString("customer_id")
	if customerID == "" {
		customerID = "demo-customer-123"
	}

	resp, err := h.createReturnUseCase.Execute(c.Request.Context(), customerID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

func (h *ReturnHandler) GetReturn(c *gin.Context) {
	returnID := c.Param("id")

	resp, err := h.getReturnUseCase.Execute(c.Request.Context(), returnID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *ReturnHandler) ListReturns(c *gin.Context) {
	// Get customer ID from context (set by auth middleware)
	// For MVP, use a demo customer ID
	customerID := c.GetString("customer_id")
	if customerID == "" {
		customerID = "demo-customer-123"
	}

	resp, err := h.listReturnsUseCase.Execute(c.Request.Context(), customerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"returns": resp})
}
