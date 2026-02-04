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

func (h *ConversationHandler) GetConversation(c *gin.Context) {
	conversationID := c.Param("id")

	resp, err := h.getConvUseCase.Execute(c.Request.Context(), conversationID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *ConversationHandler) GetMessages(c *gin.Context) {
	conversationID := c.Param("id")

	resp, err := h.getConvUseCase.Execute(c.Request.Context(), conversationID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"messages": resp.Messages})
}

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
