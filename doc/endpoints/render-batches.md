# Submit async render batch (HTML / PDF)

## HTTP

- **POST** `/api/v1/render/batches/html` — `SubmitRenderBatchHtml`
- **POST** `/api/v1/render/batches/pdf` — `SubmitRenderBatchPdf`
- **Permission:** `render.execute` (API key)

## Request body

| Field | Required | Description |
|-------|----------|-------------|
| `templateId` | yes | Latest approved version is used |
| `deliveryChannel` | yes | `none`, `email`, `sms`, or `postal` — only `none` is accepted today |
| `items` | yes | Array of `{ "variables": { ... } }` |
| `processBatchId` | no | Omit for a server-assigned process id (returned in **202**). If set, use a stable client UUID: the platform creates the process when missing for your org, or appends when the process is still open. Not the same as `batchId`. |
| `callbackUrl` | no | Optional HTTPS URL for async completion notification when the process finishes. See [Completion callback](#completion-callback) below. |
| `expectedDocumentCount` | no | **Planned mode:** total rows across all chunk POSTs for this `processBatchId`. Same value on first submit (omit on follow-ups once set). Process stays open until this many rows are accepted and rendered. If omitted, **open-ended mode** uses a 15-minute idle grace (configurable) after rendering goes idle before the callback. |

## Completion modes

- **Planned** (`expectedDocumentCount` set): callback when accepted rows reach the planned total and all jobs finish (`completionReason: expected_reached`). You can submit further chunks while the process is open even if earlier jobs already completed.
- **Open-ended** (field omitted): after all in-flight jobs finish, the platform waits for the idle grace period (default 15 minutes, reset on each accepted chunk) before closing and calling `callbackUrl` (`completionReason: idle_timeout`).

## Completion callback

When the process reaches `completed` or `completed_with_errors`, the platform POSTs JSON to `callbackUrl` **outside** the render worker.

- **Poll first:** call list/download output endpoints until outputs exist; do not rely on the callback alone.
- **Retries:** exponential backoff for up to ~24 hours, then up to seven daily attempts; then `attention_required` until an org admin with `integration.callbacks.manage` retries or abandons the callback in the admin UI (not via API key).
- **At-least-once:** dedupe on `processBatchId` and `type`.
- **No signing:** requests do not include `X-CommsPliant-Signature` (unlike organization webhooks).

```json
{
  "type": "integration_render_process.completed",
  "processBatchId": "68463aea-7762-474b-a53c-d14742900b65",
  "status": "completed",
  "completionReason": "expected_reached",
  "expectedDocumentCount": 1500,
  "batches": { "total": 1, "completed": 1, "failed": 0 },
  "documents": { "total": 100, "rendered": 98, "failed": 2 }
}
```

OpenAPI schema: `IntegrationRenderProcessCompletedCallback`.

## Response **202**

Always includes both ids:

```json
{ "processBatchId": "...", "batchId": "..." }
```

When you omit `processBatchId`, `processBatchId` in the response is newly assigned. When you set `processBatchId` on first submit, the response echoes your client id.

## Errors

| Status | When |
|--------|------|
| `400` | `code: validation_failed` — see `details.field` / `details.reason` (e.g. `callbackUrl` + `https_required`) or `details.missingFields` |
| `401` / `403` | API key auth or permission |
| `409` | `processBatchId` refers to a process that is not open |
| `422` | Template not approved, or `expectedDocumentCount` mismatch / would exceed planned total |
| `429` | Quota, parallel process limit, or batch size exceeded |

Batch submit does not return **404** for an unknown client `processBatchId`; the platform creates the process instead. List and download endpoints may still return **404** when the process does not exist for your org.

Example **400** for `http://` callback URL:

```json
{
  "error": "callbackUrl must use HTTPS",
  "code": "validation_failed",
  "details": { "field": "callbackUrl", "reason": "https_required" }
}
```

## Go SDK

Server-assigned process id:

```go
resp, err := client.SubmitRenderBatchPdf(ctx, commspliant.RenderBatchSubmitRequest{
    TemplateID:      templateID,
    DeliveryChannel: commspliant.IntegrationDeliveryChannelNone,
    Items:           []commspliant.RenderBatchItem{{Variables: map[string]any{"title": "Invoice"}}},
})
// resp.ProcessBatchID, resp.BatchID
```

Stable client process id:

```go
resp, err := client.SubmitRenderBatchPdf(ctx, commspliant.RenderBatchSubmitRequest{
    ProcessBatchID:  "68463aea-7762-474b-a53c-d14742900b65",
    TemplateID:      templateID,
    DeliveryChannel: commspliant.IntegrationDeliveryChannelNone,
    Items:           []commspliant.RenderBatchItem{{Variables: map[string]any{}}},
})
```

Planned statement (1500 rows in three chunks):

```go
expected := 1500
resp, err := client.SubmitRenderBatchPdf(ctx, commspliant.RenderBatchSubmitRequest{
    ProcessBatchID:        "68463aea-7762-474b-a53c-d14742900b65",
    TemplateID:            templateID,
    DeliveryChannel:       commspliant.IntegrationDeliveryChannelNone,
    ExpectedDocumentCount: &expected,
    Items:                 []commspliant.RenderBatchItem{{Variables: map[string]any{}}},
})
```

Use `SubmitRenderBatchHtml` for HTML batches.
