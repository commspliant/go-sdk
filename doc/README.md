# API documentation

Guides for calling the [CommsPliant Customer Integration API](https://github.com/commspliant/commspliant) with this SDK.

## Authentication

Pass your organization API key when creating the client:

- Header: `X-Api-Key: ck_...`
- Alternative: `Authorization: Bearer ck_...`

JWT user tokens are not accepted on these endpoints.

## Endpoints

| Method | Endpoint | Documentation |
|--------|----------|---------------|
| POST | `/api/v1/render/html` | [Render HTML](endpoints/render-html.md) |
| POST | `/api/v1/render/pdf` | [Render PDF](endpoints/render-pdf.md) |
| POST | `/api/v1/render/batches` | [Submit batch](endpoints/render-batches.md) |
| GET | `/api/v1/render/processes/{processBatchId}/outputs` | [List process outputs](endpoints/render-process-outputs.md) |
| GET | `/api/v1/render/outputs/{outputId}` | [Download output](endpoints/render-output-download.md) |
| POST | `/api/v1/render/outputs/download` | [Download ZIP](endpoints/render-outputs-download.md) |

## Integration batch (async)

| Method | Path | Doc |
|--------|------|-----|
| POST | `/api/v1/render/batches/html` | [Submit HTML batch](endpoints/render-batches.md) |
| POST | `/api/v1/render/batches/pdf` | [Submit PDF batch](endpoints/render-batches.md) |
| GET | `/api/v1/render/processes/{id}/outputs/html` | List HTML outputs |
| GET | `/api/v1/render/processes/{id}/outputs/pdf` | List PDF outputs |
| GET | `/api/v1/render/outputs/html/{outputId}` | Download HTML |
| GET | `/api/v1/render/outputs/pdf/{outputId}` | Download PDF |
| POST | `/api/v1/render/outputs/html/download` | ZIP HTML |
| POST | `/api/v1/render/outputs/pdf/download` | ZIP PDF |

