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
