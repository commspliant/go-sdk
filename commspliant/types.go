package commspliant

import "net/http"

const defaultBaseURL = "https://api.commspliant.com"

// ClientOptions configures a Client.
type ClientOptions struct {
	BaseURL       string
	HTTPClient    *http.Client
	UseBearerAuth bool
}

// RenderRequest is the body for sync render endpoints.
type RenderRequest struct {
	TemplateID        string
	TemplateVersionID string
	Variables         map[string]any
}

// RenderResult is a successful render response.
type RenderResult struct {
	Body                []byte
	ContentType         string
	RequestID           string
	ContentDisposition  string
}

// renderPayload is the JSON shape sent to the API.
type renderPayload struct {
	TemplateID        string         `json:"templateId"`
	TemplateVersionID string         `json:"templateVersionId,omitempty"`
	Variables         map[string]any `json:"variables"`
}
