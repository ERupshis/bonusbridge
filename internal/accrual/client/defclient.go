package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/erupshis/bonusbridge/internal/logger"
	"github.com/erupshis/bonusbridge/internal/orders/data"
)

const url = "/api/orders/"

type defaultClient struct {
	client *http.Client
	log    logger.BaseLogger
}

func CreateDefault(log logger.BaseLogger) BaseClient {
	return &defaultClient{
		client: &http.Client{},
		log:    log,
	}
}

func (c *defaultClient) RequestCalculationResult(ctx context.Context, host string, order *data.Order) (ResponseStatus, RetryInterval, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, host+url+order.Number, nil)
	if err != nil {
		return http.StatusInternalServerError, 0, fmt.Errorf("create request to loyalty system: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return http.StatusInternalServerError, 0, fmt.Errorf("client request: %w", err)
	}
	defer func() {
		if err = resp.Body.Close(); err != nil {
			c.log.Info("[accrual:defaultClient:RequestCalculationResult] failed to close response body: %v", err)
		}
	}()

	pause := 0
	if resp.StatusCode == http.StatusTooManyRequests {
		pauseStr := resp.Header.Get("Retry-After")
		pause, err = strconv.Atoi(pauseStr)
		if err != nil {
			return http.StatusInternalServerError, 0, fmt.Errorf("parse 'Retry-After' header value: %w", err)
		}
	}

	if resp.StatusCode != http.StatusOK {
		return ResponseStatus(resp.StatusCode), RetryInterval(pause), nil
	}

	bufResp := bytes.Buffer{}
	_, err = bufResp.ReadFrom(resp.Body)
	if err != nil {
		return http.StatusInternalServerError, 0, fmt.Errorf("read response body: %w", err)
	}

	if err = json.Unmarshal(bufResp.Bytes(), order); err != nil {
		return http.StatusInternalServerError, 0, fmt.Errorf("parse response body: %w", err)
	}

	return ResponseStatus(resp.StatusCode), 0, nil
}
