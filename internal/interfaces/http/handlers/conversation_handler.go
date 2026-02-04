package handlers

import (
	"net/http"

	"customer-support-api/internal/application/dto"
	"customer-support-api/internal/application/usecases"

	"github.com/gin-gonic/gin"
)

type ConversationHandler struct {
	createConvUseCase  *usecases.CreateConversationUseCase
	getConvUseCase     *usecases.GetConversationUseCase
	sendMessageUseCase *usecases.SendMessageUseCase
}

func NewConversationHandler(
	createConvUseCase *usecases.CreateConversationUseCase,
	getConvUseCase *usecases.GetConversationUseCase,
	sendMessageUseCase *usecases.SendMessageUseCase,
) *ConversationHandler {
	return &ConversationHandler{
		createConvUseCase:  createConvUseCase,
		getConvUseCase:     getConvUseCase,
		sendMessageUseCase: sendMessageUseCase,
	}
}

// CreateConversation godoc
// @Summary Create a new conversation
// @Description Create a new customer support conversation (natural language or guided flow)
// @Tags Conversations
// @Accept json
// @Produce json
// @Param request body dto.CreateConversationRequest true "Conversation creation request"
// @Success 201 {object} dto.CreateConversationResponse
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /conversations [post]
func (h *ConversationHandler) CreateConversation(c *gin.Context) {
	var req dto.CreateConversationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get customer ID from context (set by auth middleware)
	// For MVP, use a demo customer ID
	customerID := c.GetString("customer_id")
	if customerID == "" {
		customerID = "demo-customer-123"
	}

	resp, err := h.createConvUseCase.Execute(c.Request.Context(), customerID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// GetConversation godoc
// @Summary Get conversation details
// @Description Retrieve a conversation with all its messages
// @Tags Conversations
// @Produce json
// @Param id path string true "Conversation ID"
// @Success 200 {object} dto.ConversationDetailsResponse
// @Failure 404 {object} map[string]string "Conversation not found"
// @Router /conversations/{id} [get]
func (h *ConversationHandler) GetConversation(c *gin.Context) {
	conversationID := c.Param("id")

	resp, err := h.getConvUseCase.Execute(c.Request.Context(), conversationID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetMessages godoc
// @Summary Get conversation messages
// @Description Retrieve all messages in a conversation
// @Tags Conversations
// @Produce json
// @Param id path string true "Conversation ID"
// @Success 200 {object} map[string][]dto.MessageDTO
// @Failure 404 {object} map[string]string "Conversation not found"
// @Router /conversations/{id}/messages [get]
func (h *ConversationHandler) GetMessages(c *gin.Context) {
	conversationID := c.Param("id")

	resp, err := h.getConvUseCase.Execute(c.Request.Context(), conversationID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"messages": resp.Messages})
}

// SendMessage godoc
// @Summary Send a message in a conversation
// @Description Send a user message and receive an AI assistant response
// @Tags Conversations
// @Accept json
// @Produce json
// @Param id path string true "Conversation ID"
// @Param request body dto.SendMessageRequest true "Message content"
// @Success 200 {object} dto.SendMessageResponse
// @Failure 400 {object} map[string]string "Invalid request"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /conversations/{id}/messages [post]
func (h *ConversationHandler) SendMessage(c *gin.Context) {
	conversationID := c.Param("id")

	var req dto.SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.sendMessageUseCase.Execute(c.Request.Context(), conversationID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}
