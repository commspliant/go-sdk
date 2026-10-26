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

// IntegrationDeliveryChannel selects post-render delivery for integration batches.
type IntegrationDeliveryChannel string

const (
	IntegrationDeliveryChannelNone   IntegrationDeliveryChannel = "none"
	IntegrationDeliveryChannelEmail  IntegrationDeliveryChannel = "email"
	IntegrationDeliveryChannelSMS    IntegrationDeliveryChannel = "sms"
	IntegrationDeliveryChannelPostal IntegrationDeliveryChannel = "postal"
)

// RenderBatchSubmitRequest is the async batch submit body.
type RenderBatchSubmitRequest struct {
	ProcessBatchID  string                       `json:"processBatchId,omitempty"`
	TemplateID      string                       `json:"templateId"`
	DeliveryChannel IntegrationDeliveryChannel   `json:"deliveryChannel"`
	Items           []RenderBatchItem            `json:"items"`
	CallbackURL             string                       `json:"callbackUrl,omitempty"`
	ExpectedDocumentCount   *int                         `json:"expectedDocumentCount,omitempty"`
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

// SubmitRenderBatchHtml enqueues an async HTML integration render batch.
func (c *Client) SubmitRenderBatchHtml(ctx context.Context, req RenderBatchSubmitRequest) (*RenderBatchSubmitResponse, error) {
	return c.submitRenderBatch(ctx, "/api/v1/render/batches/html", req)
}

// SubmitRenderBatchPdf enqueues an async PDF integration render batch.
func (c *Client) SubmitRenderBatchPdf(ctx context.Context, req RenderBatchSubmitRequest) (*RenderBatchSubmitResponse, error) {
	return c.submitRenderBatch(ctx, "/api/v1/render/batches/pdf", req)
}

func (c *Client) submitRenderBatch(ctx context.Context, path string, req RenderBatchSubmitRequest) (*RenderBatchSubmitResponse, error) {
	if err := validateBatchSubmitRequest(req); err != nil {
		return nil, err
	}
	var out RenderBatchSubmitResponse
	if err := c.postJSON(ctx, path, req, http.StatusAccepted, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// ListRenderProcessHtmlOutputs lists HTML outputs for a processBatchId.
func (c *Client) ListRenderProcessHtmlOutputs(ctx context.Context, processBatchID string, opts ListRenderProcessOutputsOptions) (*IntegrationProcessOutputsPage, error) {
	return c.listRenderProcessOutputs(ctx, processBatchID, "html", opts)
}

// ListRenderProcessPdfOutputs lists PDF outputs for a processBatchId.
func (c *Client) ListRenderProcessPdfOutputs(ctx context.Context, processBatchID string, opts ListRenderProcessOutputsOptions) (*IntegrationProcessOutputsPage, error) {
	return c.listRenderProcessOutputs(ctx, processBatchID, "pdf", opts)
}

func (c *Client) listRenderProcessOutputs(ctx context.Context, processBatchID, format string, opts ListRenderProcessOutputsOptions) (*IntegrationProcessOutputsPage, error) {
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
	path := fmt.Sprintf("/api/v1/render/processes/%s/outputs/%s", processBatchID, format)
	if len(q) > 0 {
		path += "?" + q.Encode()
	}
	var out IntegrationProcessOutputsPage
	if err := c.getJSON(ctx, path, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

// DownloadRenderHtmlOutput streams one HTML output by id.
func (c *Client) DownloadRenderHtmlOutput(ctx context.Context, outputID string) (*RenderResult, error) {
	return c.downloadRenderOutput(ctx, "html", outputID)
}

// DownloadRenderPdfOutput streams one PDF output by id.
func (c *Client) DownloadRenderPdfOutput(ctx context.Context, outputID string) (*RenderResult, error) {
	return c.downloadRenderOutput(ctx, "pdf", outputID)
}

func (c *Client) downloadRenderOutput(ctx context.Context, format, outputID string) (*RenderResult, error) {
	if outputID == "" {
		return nil, fmt.Errorf("outputId is required")
	}
	return c.getBinary(ctx, fmt.Sprintf("/api/v1/render/outputs/%s/%s", format, outputID))
}

// DownloadRenderHtmlOutputsZip downloads a ZIP of the given HTML output ids.
func (c *Client) DownloadRenderHtmlOutputsZip(ctx context.Context, outputIDs []string) (*RenderResult, error) {
	return c.downloadRenderOutputsZip(ctx, "html", outputIDs)
}

// DownloadRenderPdfOutputsZip downloads a ZIP of the given PDF output ids.
func (c *Client) DownloadRenderPdfOutputsZip(ctx context.Context, outputIDs []string) (*RenderResult, error) {
	return c.downloadRenderOutputsZip(ctx, "pdf", outputIDs)
}

func (c *Client) downloadRenderOutputsZip(ctx context.Context, format string, outputIDs []string) (*RenderResult, error) {
	if len(outputIDs) == 0 {
		return nil, fmt.Errorf("outputIds is required")
	}
	body, err := json.Marshal(map[string][]string{"outputIds": outputIDs})
	if err != nil {
		return nil, err
	}
	path := fmt.Sprintf("/api/v1/render/outputs/%s/download", format)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(body))
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
	if req.DeliveryChannel == "" {
		return fmt.Errorf("deliveryChannel is required")
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
