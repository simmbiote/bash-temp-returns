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

// CreateReturn godoc
// @Summary Create a return request
// @Description Create a new return request with order validation and automatic refund calculation
// @Tags Returns
// @Accept json
// @Produce json
// @Param request body dto.CreateReturnRequest true "Return request details"
// @Success 201 {object} dto.CreateReturnResponse
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /returns [post]
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

// GetReturn godoc
// @Summary Get return request details
// @Description Retrieve detailed information about a specific return request
// @Tags Returns
// @Produce json
// @Param id path string true "Return Request ID"
// @Success 200 {object} dto.ReturnRequestDTO
// @Failure 404 {object} map[string]string "Return not found"
// @Router /returns/{id} [get]
func (h *ReturnHandler) GetReturn(c *gin.Context) {
	returnID := c.Param("id")

	resp, err := h.getReturnUseCase.Execute(c.Request.Context(), returnID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// ListReturns godoc
// @Summary List return requests
// @Description Retrieve all return requests for the authenticated customer
// @Tags Returns
// @Produce json
// @Param customer_id query string false "Customer ID (defaults to demo-customer-123)"
// @Success 200 {object} map[string][]dto.ReturnRequestDTO
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /returns [get]
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
