package client

import (
	"bytes"
	"context"
	"fmt"
	"net/http"

	"github.com/erupshis/bonusbridge/internal/helpers"
	"github.com/erupshis/bonusbridge/internal/logger"
	"github.com/erupshis/bonusbridge/internal/orders/data"
)

const url = "/api/orders/"

type defaultClient struct {
	client *http.Client
	log    logger.BaseLogger
}

func CreateDefault(client *http.Client, log logger.BaseLogger) BaseClient {
	return &defaultClient{
		client: client,
		log:    log,
	}
}

func (c *defaultClient) RequestCalculationResult(ctx context.Context, order *data.Order) (int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url+order.Number, nil)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("create request to loyalty system: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return resp.StatusCode, fmt.Errorf("client request error: %w", err)
	}
	defer helpers.ExecuteWithLogError(resp.Body.Close, c.log)

	bufResp := bytes.Buffer{}
	_, err = bufResp.ReadFrom(resp.Body)
	if err != nil {
		return http.StatusInternalServerError, fmt.Errorf("read response body: %w", err)
	}

	return resp.StatusCode, nil
}
