package commspliant

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRenderHTML_success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v1/render/html" {
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		if r.Header.Get("X-Api-Key") != "ck_test" {
			t.Fatalf("unexpected api key: %s", r.Header.Get("X-Api-Key"))
		}

		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if body["templateId"] != "550e8400-e29b-41d4-a716-446655440000" {
			t.Fatalf("unexpected templateId: %v", body["templateId"])
		}

		w.Header().Set("Content-Type", "text/html")
		w.Header().Set("X-Request-ID", "req-123")
		w.Header().Set("Content-Disposition", `inline; filename="report.html"`)
		_, _ = w.Write([]byte("<html>ok</html>"))
	}))
	defer server.Close()

	client, err := NewClient("ck_test", &ClientOptions{BaseURL: server.URL})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	result, err := client.RenderHTML(context.Background(), RenderRequest{
		TemplateID: "550e8400-e29b-41d4-a716-446655440000",
		Variables:  map[string]any{"title": "Monthly Report"},
	})
	if err != nil {
		t.Fatalf("RenderHTML: %v", err)
	}
	if string(result.Body) != "<html>ok</html>" {
		t.Fatalf("unexpected body: %s", result.Body)
	}
	if result.RequestID != "req-123" {
		t.Fatalf("unexpected request id: %s", result.RequestID)
	}
}

func TestRenderPDF_success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/render/pdf" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/pdf")
		_, _ = w.Write([]byte("%PDF-1.4"))
	}))
	defer server.Close()

	client, err := NewClient("ck_test", &ClientOptions{BaseURL: server.URL})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	result, err := client.RenderPDF(context.Background(), RenderRequest{
		TemplateID: "550e8400-e29b-41d4-a716-446655440000",
		Variables:  map[string]any{"title": "Invoice"},
	})
	if err != nil {
		t.Fatalf("RenderPDF: %v", err)
	}
	if !strings.HasPrefix(string(result.Body), "%PDF") {
		t.Fatalf("unexpected body: %s", result.Body)
	}
}

func TestRenderHTML_missingTemplateID(t *testing.T) {
	client, err := NewClient("ck_test", nil)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	_, err = client.RenderHTML(context.Background(), RenderRequest{
		Variables: map[string]any{},
	})
	if err == nil || !strings.Contains(err.Error(), "templateId is required") {
		t.Fatalf("expected templateId validation error, got %v", err)
	}
}

func TestRenderHTML_apiError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Request-ID", "req-404")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "template not found"})
	}))
	defer server.Close()

	client, err := NewClient("ck_test", &ClientOptions{BaseURL: server.URL})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	_, err = client.RenderHTML(context.Background(), RenderRequest{
		TemplateID: "550e8400-e29b-41d4-a716-446655440000",
		Variables:  map[string]any{},
	})
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected APIError, got %T (%v)", err, err)
	}
	if apiErr.StatusCode != http.StatusNotFound {
		t.Fatalf("unexpected status: %d", apiErr.StatusCode)
	}
	if apiErr.Message != "template not found" {
		t.Fatalf("unexpected message: %s", apiErr.Message)
	}
	if apiErr.RequestID != "req-404" {
		t.Fatalf("unexpected request id: %s", apiErr.RequestID)
	}
}

func TestNewClient_requiresApiKey(t *testing.T) {
	_, err := NewClient(" ", nil)
	if err == nil {
		t.Fatal("expected error for empty api key")
	}
}

func TestBearerAuth(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer ck_test" {
			t.Fatalf("unexpected authorization: %s", r.Header.Get("Authorization"))
		}
		_, _ = io.Copy(w, r.Body)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	client, err := NewClient("ck_test", &ClientOptions{
		BaseURL:       server.URL,
		UseBearerAuth: true,
	})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	_, err = client.RenderHTML(context.Background(), RenderRequest{
		TemplateID: "550e8400-e29b-41d4-a716-446655440000",
		Variables:  map[string]any{},
	})
	if err != nil {
		t.Fatalf("RenderHTML: %v", err)
	}
}
