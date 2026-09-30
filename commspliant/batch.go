package commspliant

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// BatchPurpose selects integration batch output mode.
type BatchPurpose string

const (
	BatchPurposeEmail      BatchPurpose = "email"
	BatchPurposePDF        BatchPurpose = "pdf"
	BatchPurposeRenderOnly BatchPurpose = "render_only"
)

// RenderBatchSubmitRequest is the async batch submit body.
type RenderBatchSubmitRequest struct {
	ProcessBatchID string              `json:"processBatchId,omitempty"`
	TemplateID     string              `json:"templateId"`
	Purpose        BatchPurpose        `json:"purpose"`
	Items          []RenderBatchItem   `json:"items"`
	CallbackURL    string              `json:"callbackUrl,omitempty"`
}

// RenderBatchItem is one row in a batch.
type RenderBatchItem struct {
	Variables map[string]any `json:"variables"`
}

// RenderBatchSubmitResponse is returned with HTTP 202.
type RenderBatchSubmitResponse struct {
	ProcessBatchID string `json:"processBatchId"`
	BatchID        string `json:"batchId"`
}

// ListRenderProcessOutputsOptions configures pagination.
type ListRenderProcessOutputsOptions struct {
	Page     int
	PageSize int
}

// IntegrationProcessOutputsPage is a paginated list of output metadata.
type IntegrationProcessOutputsPage struct {
	Data []IntegrationProcessOutput `json:"data"`
	Meta PageMeta                   `json:"meta"`
}

// IntegrationProcessOutput describes one stored render output.
type IntegrationProcessOutput struct {
	ID          string `json:"id"`
	BatchID     string `json:"batchId"`
	RowIndex    int    `json:"rowIndex"`
	Format      string `json:"format"`
	ContentType string `json:"contentType"`
	ByteSize    int64  `json:"byteSize"`
	ItemStatus  string `json:"itemStatus"`
	CreatedAt   string `json:"createdAt"`
}

// PageMeta is pagination metadata.
type PageMeta struct {
	Page     int   `json:"page"`
	PageSize int   `json:"pageSize"`
	Total    int64 `json:"total"`
}

// SubmitRenderBatch enqueues an async integration render batch.
func (c *Client) SubmitRenderBatch(ctx context.Context, req RenderBatchSubmitRequest) (*RenderBatchSubmitResponse, error) {
	if err := validateBatchSubmitRequest(req); err != nil {
		return nil, err
	}
	var out RenderBatchSubmitResponse
	if err := c.postJSON(ctx, "/api/v1/render/batches", req, http.StatusAccepted, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListRenderProcessOutputs lists outputs for a processBatchId.
func (c *Client) ListRenderProcessOutputs(ctx context.Context, processBatchID string, opts ListRenderProcessOutputsOptions) (*IntegrationProcessOutputsPage, error) {
	if processBatchID == "" {
		return nil, fmt.Errorf("processBatchId is required")
	}
	q := url.Values{}
	if opts.Page > 0 {
		q.Set("page", fmt.Sprintf("%d", opts.Page))
	}
	if opts.PageSize > 0 {
		q.Set("pageSize", fmt.Sprintf("%d", opts.PageSize))
	}
	path := "/api/v1/render/processes/" + processBatchID + "/outputs"
	if len(q) > 0 {
		path += "?" + q.Encode()
	}
	var out IntegrationProcessOutputsPage
	if err := c.getJSON(ctx, path, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DownloadRenderOutput streams one output by id.
func (c *Client) DownloadRenderOutput(ctx context.Context, outputID string) (*RenderResult, error) {
	if outputID == "" {
		return nil, fmt.Errorf("outputId is required")
	}
	return c.getBinary(ctx, "/api/v1/render/outputs/"+outputID)
}

// DownloadRenderOutputsZip downloads a ZIP of the given output ids.
func (c *Client) DownloadRenderOutputsZip(ctx context.Context, outputIDs []string) (*RenderResult, error) {
	if len(outputIDs) == 0 {
		return nil, fmt.Errorf("outputIds is required")
	}
	body, err := json.Marshal(map[string][]string{"outputIds": outputIDs})
	if err != nil {
		return nil, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api/v1/render/outputs/download", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	c.setAuth(httpReq)
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, parseAPIError(resp)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return &RenderResult{Body: data, ContentType: resp.Header.Get("Content-Type"), RequestID: resp.Header.Get("X-Request-ID")}, nil
}

func validateBatchSubmitRequest(req RenderBatchSubmitRequest) error {
	if req.TemplateID == "" {
		return fmt.Errorf("templateId is required")
	}
	if req.Purpose == "" {
		return fmt.Errorf("purpose is required")
	}
	if len(req.Items) == 0 {
		return fmt.Errorf("items is required")
	}
	return nil
}

func (c *Client) postJSON(ctx context.Context, path string, payload any, expectStatus int, out any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(body))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	c.setAuth(httpReq)
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != expectStatus {
		return parseAPIError(resp)
	}
	if out != nil {
		return json.NewDecoder(resp.Body).Decode(out)
	}
	return nil
}

func (c *Client) getJSON(ctx context.Context, path string, out any) error {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return err
	}
	c.setAuth(httpReq)
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return parseAPIError(resp)
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

func (c *Client) getBinary(ctx context.Context, path string) (*RenderResult, error) {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return nil, err
	}
	c.setAuth(httpReq)
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, parseAPIError(resp)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return &RenderResult{
		Body:               data,
		ContentType:        resp.Header.Get("Content-Type"),
		RequestID:          resp.Header.Get("X-Request-ID"),
		ContentDisposition: resp.Header.Get("Content-Disposition"),
	}, nil
}

func (c *Client) setAuth(httpReq *http.Request) {
	if c.useBearerAuth {
		httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	} else {
		httpReq.Header.Set("X-Api-Key", c.apiKey)
	}
}
