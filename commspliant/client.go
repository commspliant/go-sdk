package commspliant

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Client calls the CommsPliant Customer Integration API.
type Client struct {
	apiKey        string
	baseURL       string
	httpClient    *http.Client
	useBearerAuth bool
}

// NewClient creates a client with the given organization API key.
func NewClient(apiKey string, opts *ClientOptions) (*Client, error) {
	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		return nil, fmt.Errorf("api key is required")
	}

	client := &Client{
		apiKey:     apiKey,
		baseURL:    defaultBaseURL,
		httpClient: http.DefaultClient,
	}
	if opts != nil {
		if opts.BaseURL != "" {
			client.baseURL = strings.TrimRight(opts.BaseURL, "/")
		}
		if opts.HTTPClient != nil {
			client.httpClient = opts.HTTPClient
		}
		client.useBearerAuth = opts.UseBearerAuth
	}
	return client, nil
}

// RenderHTML renders an approved template to HTML.
func (c *Client) RenderHTML(ctx context.Context, req RenderRequest) (*RenderResult, error) {
	return c.postRender(ctx, "/api/v1/render/html", req)
}

// RenderPDF renders an approved template to PDF.
func (c *Client) RenderPDF(ctx context.Context, req RenderRequest) (*RenderResult, error) {
	return c.postRender(ctx, "/api/v1/render/pdf", req)
}

func (c *Client) postRender(ctx context.Context, path string, req RenderRequest) (*RenderResult, error) {
	if err := validateRenderRequest(req); err != nil {
		return nil, err
	}

	body, err := json.Marshal(toRenderPayload(req))
	if err != nil {
		return nil, fmt.Errorf("encode render request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if c.useBearerAuth {
		httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	} else {
		httpReq.Header.Set("X-Api-Key", c.apiKey)
	}

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, parseAPIError(resp)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	return &RenderResult{
		Body:               data,
		ContentType:        resp.Header.Get("Content-Type"),
		RequestID:          resp.Header.Get("X-Request-ID"),
		ContentDisposition: resp.Header.Get("Content-Disposition"),
	}, nil
}

func validateRenderRequest(req RenderRequest) error {
	if strings.TrimSpace(req.TemplateID) == "" {
		return fmt.Errorf("templateId is required")
	}
	if req.Variables == nil {
		return fmt.Errorf("variables is required")
	}
	return nil
}

func toRenderPayload(req RenderRequest) renderPayload {
	return renderPayload{
		TemplateID: req.TemplateID,
		Variables:  req.Variables,
	}
}

func parseAPIError(resp *http.Response) error {
	requestID := resp.Header.Get("X-Request-ID")
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return &APIError{
			StatusCode: resp.StatusCode,
			Message:    http.StatusText(resp.StatusCode),
			RequestID:  requestID,
		}
	}

	apiErr := &APIError{
		StatusCode: resp.StatusCode,
		RequestID:  requestID,
	}
	if len(body) == 0 {
		apiErr.Message = http.StatusText(resp.StatusCode)
		return apiErr
	}

	var parsed struct {
		Error   string         `json:"error"`
		Code    string         `json:"code"`
		Details map[string]any `json:"details"`
	}
	if err := json.Unmarshal(body, &parsed); err != nil {
		apiErr.Message = strings.TrimSpace(string(body))
		return apiErr
	}
	apiErr.Message = parsed.Error
	apiErr.Code = parsed.Code
	apiErr.Details = parsed.Details
	if apiErr.Message == "" {
		apiErr.Message = http.StatusText(resp.StatusCode)
	}
	return apiErr
}
