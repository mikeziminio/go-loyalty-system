package accrual

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/mikeziminio/go-loyalty-system/internal/model"
)

type Client struct {
	httpClient *http.Client
	baseURL    string
}

func NewClient(baseURL string, timeout time.Duration) *Client {
	httpClient := http.Client{
		Timeout: timeout,
	}
	return &Client{
		httpClient: &httpClient,
		baseURL:    baseURL,
	}
}

type orderInfoResponse struct {
	ID      string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual,omitempty"`
}

func (c *Client) Order(ctx context.Context, id string) (*model.Order, error) {
	path := fmt.Sprintf("/api/orders/%s", id)
	fullURL, err := url.JoinPath(c.baseURL, path)
	if err != nil {
		return nil, fmt.Errorf("failed to make url: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch order info: %w", err)
	}
	// nilaway ругается в этом месте,
	// считает что возможны ситуации когда err == nil и res == nil
	if res == nil {
		return nil, fmt.Errorf("failed to fetch order info: nil response")
	}

	if res.StatusCode != http.StatusOK {
		switch res.StatusCode {
		case http.StatusNoContent:
			return nil, fmt.Errorf("order is not registered")
		case http.StatusTooManyRequests:
			return nil, fmt.Errorf("too many request")
		default:
			return nil, fmt.Errorf("error %d: %s", res.StatusCode, http.StatusText(res.StatusCode))
		}
	}

	body := res.Body
	defer body.Close()
	data, err := io.ReadAll(body)
	if err != nil {
		return nil, fmt.Errorf("failed to read body: %w", err)
	}

	var oi orderInfoResponse
	err = json.Unmarshal(data, &oi)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal: %w", err)
	}

	// todo: validate
	return &model.Order{
		ID:      oi.ID,
		Status:  model.OrderStatus(oi.Status),
		Accrual: oi.Accrual,
	}, nil
}
