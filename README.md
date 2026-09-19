# CommsPliant Go SDK

Official Go client for the [CommsPliant Customer Integration API](https://developer.commspliant.com/).

## Installation

```bash
go get github.com/commspliant/go-sdk/commspliant
```

Or clone this repository and use a local path in your `go.mod`.

## Quickstart

```go
package main

import (
    "context"
    "fmt"
    "os"

    "github.com/commspliant/go-sdk/commspliant"
)

func main() {
    client, err := commspliant.NewClient("ck_YOUR_API_KEY", nil)
    if err != nil {
        panic(err)
    }

    result, err := client.RenderHTML(context.Background(), commspliant.RenderRequest{
        TemplateID: "550e8400-e29b-41d4-a716-446655440000",
        Variables: map[string]any{
            "title": "Monthly Report",
            "user":  map[string]any{"name": "Jane Doe"},
        },
    })
    if err != nil {
        panic(err)
    }

    if err := os.WriteFile("document.html", result.Body, 0644); err != nil {
        panic(err)
    }

    fmt.Println("saved document.html")
}
```

## Authentication

Pass your organization API key when creating the client. By default the SDK sends `X-Api-Key: ck_...`.

To use `Authorization: Bearer ck_...` instead:

```go
client, err := commspliant.NewClient("ck_YOUR_API_KEY", &commspliant.ClientOptions{
    UseBearerAuth: true,
})
```

## Configuration

```go
client, err := commspliant.NewClient("ck_YOUR_API_KEY", &commspliant.ClientOptions{
    BaseURL: "http://localhost:8085", // optional override
})
```

## Errors

Non-success API responses return `*commspliant.APIError` with `StatusCode`, `Message`, and `RequestID`.

## Documentation

Endpoint guides and SDK usage examples: [doc/README.md](doc/README.md)

## Links

- [Developer Portal](https://developer.commspliant.com/)
- [About CommsPliant](https://commspliant.com/)
