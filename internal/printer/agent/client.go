package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	domainreceipt "pos-system/internal/domain/receipt"
)

// ErrPrintTimeout is returned when the print agent times out or returns PRINT_TIMEOUT.
var ErrPrintTimeout = errors.New("PRINT_TIMEOUT")

type PrintResponse struct {
	JobID   string `json:"job_id"`
	Message string `json:"message"`
}

type PrintAgentClient interface {
	Print(
		ctx context.Context,
		receipt domainreceipt.Receipt,
		idempotencyKey string,
	) (PrintResponse, error)
}

type HTTPClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewHTTPClient(baseURL string) *HTTPClient {
	return &HTTPClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *HTTPClient) Print(
	ctx context.Context,
	receipt domainreceipt.Receipt,
	idempotencyKey string,
) (PrintResponse, error) {
	if strings.TrimSpace(idempotencyKey) == "" {
		return PrintResponse{}, fmt.Errorf("idempotency key is required")
	}

	body, err := json.Marshal(toPrintRequest(receipt))
	if err != nil {
		return PrintResponse{}, fmt.Errorf("marshal print request: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.baseURL+"/print",
		bytes.NewReader(body),
	)
	if err != nil {
		return PrintResponse{}, fmt.Errorf("create print request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Idempotency-Key", idempotencyKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(ctx.Err(), context.DeadlineExceeded) {
			return PrintResponse{}, fmt.Errorf("%w: print request context deadline exceeded: %v", ErrPrintTimeout, err)
		}
		return PrintResponse{}, fmt.Errorf("print agent request failed: %w", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return PrintResponse{}, fmt.Errorf("read print agent response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		respStr := strings.TrimSpace(string(responseBody))
		if resp.StatusCode == http.StatusGatewayTimeout || strings.Contains(respStr, "PRINT_TIMEOUT") {
			return PrintResponse{}, fmt.Errorf("%w: print agent returned %d: %s", ErrPrintTimeout, resp.StatusCode, respStr)
		}
		return PrintResponse{}, fmt.Errorf(
			"print agent returned %d: %s",
			resp.StatusCode,
			respStr,
		)
	}

	var printResp PrintResponse
	if err := json.Unmarshal(responseBody, &printResp); err != nil {
		return PrintResponse{}, fmt.Errorf("decode print agent response: %w", err)
	}

	return printResp, nil
}
