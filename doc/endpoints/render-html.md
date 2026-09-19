# Render HTML

## HTTP

- **Method:** `POST`
- **Path:** `/api/v1/render/html`
- **Permission:** `render.execute`

## Description

Resolves an **approved** template version and returns rendered HTML as a streamed response.

- `templateId` is required.
- `templateVersionId` is optional; when omitted, the latest approved version is used.
- `variables` supplies values for `{{placeholder}}` syntax in the template.

## Request

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `templateId` | UUID string | Yes | Template to render |
| `templateVersionId` | UUID string | No | Explicit approved version override |
| `variables` | object | Yes | Values for template placeholders |

## Response

- **Status:** `200 OK`
- **Content-Type:** `text/html`
- **Body:** Rendered HTML stream
- **Headers:** `X-Request-ID` (request correlation ID), `Content-Disposition`

## Errors

| Status | Meaning |
|--------|---------|
| 400 | Invalid request body or parameters |
| 401 | Missing or invalid API key |
| 403 | Valid API key but missing `render.execute` permission |
| 404 | Template not found or not visible in the key's organization |
| 422 | Template version not approved for rendering |
| 429 | Render quota exceeded |
| 500 | Unexpected server error |

## SDK example

```go
package main

import (
    "context"
    "os"

    "github.com/commspliant/go-sdk/commspliant"
)

func main() {
    ctx := context.Background()

    client, err := commspliant.NewClient("ck_YOUR_API_KEY", nil)
    if err != nil {
        panic(err)
    }

    result, err := client.RenderHTML(ctx, commspliant.RenderRequest{
        TemplateID: "550e8400-e29b-41d4-a716-446655440000",
        Variables: map[string]any{
            "title": "Monthly Report",
            "user":  map[string]any{"name": "Jane Doe"},
        },
    })
    if err != nil {
        panic(err)
    }

    out, err := os.Create("document.html")
    if err != nil {
        panic(err)
    }
    defer out.Close()

    if _, err := out.Write(result.Body); err != nil {
        panic(err)
    }
}
```
