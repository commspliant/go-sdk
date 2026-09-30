# Submit async render batch

## HTTP

- **Method:** `POST`
- **Path:** `/api/v1/render/batches`
- **Permission:** `render.execute`

## Description

Accepts a batch of variable rows for one template. Returns `processBatchId` and `batchId` with HTTP 202.

## SDK example

```go
resp, err := client.SubmitRenderBatch(ctx, commspliant.RenderBatchSubmitRequest{
    TemplateID: "550e8400-e29b-41d4-a716-446655440000",
    Purpose:    commspliant.BatchPurposePDF,
    Items: []commspliant.RenderBatchItem{
        {Variables: map[string]any{"title": "Invoice", "amount": "99.00"}},
    },
})
```
