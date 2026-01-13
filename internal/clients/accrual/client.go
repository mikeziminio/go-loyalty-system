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

func NewClient(baseURL string) *Client {
	httpClient := http.Client{
		Timeout: 10 * time.Second,
	}
	return &Client{
		httpClient: &httpClient,
		baseURL:    baseURL,
	}
}

type orderInfoResponse struct {
	Id      string `json:"order"`
	Status  string `json:"status"`
	Accrual int    `json:"accrual"`
}

func (c *Client) OrderInfo(ctx context.Context, id string) (*model.Order, error) {
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

	var oi orderInfoResponse
	err = json.Unmarshal(data, &oi)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal: %w", err)
	}

	// todo: validate
	return &model.Order{
		Id:      oi.Id,
		Status:  model.OrderStatus(oi.Status),
		Accrual: oi.Accrual,
	}, nil
}
