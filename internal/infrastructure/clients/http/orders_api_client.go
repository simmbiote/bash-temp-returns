package http

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"customer-support-api/internal/domain/models"
	"customer-support-api/internal/domain/services"
	"customer-support-api/internal/pkg/errors"
)

type ordersAPIClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// NewOrdersAPIClient creates a new Orders API client
func NewOrdersAPIClient(baseURL string) services.OrdersAPIClient {
	return &ordersAPIClient{
		baseURL: baseURL,
		apiKey:  "", // TODO: Add API key from config
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *ordersAPIClient) GetOrderByNumber(ctx context.Context, orderNumber string) (*models.Order, error) {
	url := fmt.Sprintf("%s/api/v1/orders?order_number=%s", c.baseURL, orderNumber)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, errors.NewDatabaseError("failed to create request", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errors.NewDatabaseError("failed to call orders API", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, errors.NewNotFoundError("order", orderNumber)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, errors.NewDatabaseError(
			fmt.Sprintf("orders API returned status %d: %s", resp.StatusCode, string(body)),
			nil,
		)
	}

	var order models.Order
	if err := json.NewDecoder(resp.Body).Decode(&order); err != nil {
		return nil, errors.NewDatabaseError("failed to decode order response", err)
	}

	return &order, nil
}

func (c *ordersAPIClient) GetOrderByID(ctx context.Context, orderID string) (*models.Order, error) {
	url := fmt.Sprintf("%s/api/v1/orders/%s", c.baseURL, orderID)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, errors.NewDatabaseError("failed to create request", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, errors.NewDatabaseError("failed to call orders API", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, errors.NewNotFoundError("order", orderID)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, errors.NewDatabaseError(
			fmt.Sprintf("orders API returned status %d: %s", resp.StatusCode, string(body)),
			nil,
		)
	}

	var order models.Order
	if err := json.NewDecoder(resp.Body).Decode(&order); err != nil {
		return nil, errors.NewDatabaseError("failed to decode order response", err)
	}

	return &order, nil
}

func (c *ordersAPIClient) ValidateOrderItems(ctx context.Context, orderNumber string, itemIDs []string) error {
	order, err := c.GetOrderByNumber(ctx, orderNumber)
	if err != nil {
		return err
	}

	// Check each item exists and can be returned
	for _, itemID := range itemIDs {
		item := order.GetItemByID(itemID)
		if item == nil {
			return errors.NewInvalidInputError(
				fmt.Sprintf("item %s not found in order %s", itemID, orderNumber),
				nil,
			)
		}

		if !order.CanReturnItem(itemID) {
			return errors.NewInvalidInputError(
				fmt.Sprintf("item %s cannot be returned", itemID),
				nil,
			)
		}
	}

	return nil
}

func (c *ordersAPIClient) CalculateRefundAmount(ctx context.Context, orderNumber string, itemIDs []string) (int, error) {
	order, err := c.GetOrderByNumber(ctx, orderNumber)
	if err != nil {
		return 0, err
	}

	return order.CalculateRefundAmount(itemIDs), nil
}
