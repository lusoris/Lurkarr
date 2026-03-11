# OpenAPI + ogen + Scalar

## ogen — Go OpenAPI v3 Code Generator

> Docs: https://ogen.dev/
> Repo: https://github.com/ogen-go/ogen

### Overview

ogen generates Go server and client code from OpenAPI v3 specifications. Unlike other generators, it produces idiomatic Go code with full type safety.

### Benefits for Lurkarr

- **Type-safe handlers** — request/response structs generated from spec
- **Built-in validation** — path params, query params, request body validation
- **Eliminates manual MaxBytesReader** — built into generated middleware
- **Documentation-driven** — spec is single source of truth
- **Client generation** — type-safe client for testing

### Installation

```bash
go install github.com/ogen-go/ogen/cmd/ogen@latest
```

### Usage

```bash
ogen -target internal/api/generated -package api -clean openapi.yaml
```

### Generated Code Structure

```
internal/api/generated/
├── oas_server_gen.go       # Server interface (implement this)
├── oas_handlers_gen.go     # HTTP routing + middleware
├── oas_request_decoders.go # Request parsing
├── oas_response_encoders.go # Response writing
├── oas_schemas_gen.go      # Request/response types
├── oas_validators_gen.go   # Input validation
└── oas_client_gen.go       # Type-safe client
```

### Handler Interface Pattern

```go
// Generated interface — implement in your code
type Handler interface {
    GetSettings(ctx context.Context) (*Settings, error)
    UpdateSettings(ctx context.Context, req *UpdateSettingsReq) error
    GetApps(ctx context.Context) ([]App, error)
    // ... all endpoints
}
```

### OpenAPI Spec Structure for Lurkarr

```yaml
openapi: "3.1.0"
info:
  title: Lurkarr API
  version: "2.0.0"
servers:
  - url: /api
paths:
  /settings:
    get:
      operationId: getSettings
      responses:
        '200':
          content:
            application/json:
              schema: { $ref: '#/components/schemas/Settings' }
    put:
      operationId: updateSettings
      requestBody:
        required: true
        content:
          application/json:
            schema: { $ref: '#/components/schemas/UpdateSettingsRequest' }
      responses:
        '200': ...
        '422': ...
```

## Scalar — Interactive API Documentation

> Repo: https://github.com/scalar/scalar
> CDN: `@scalar/api-reference`

### Integration (zero-config)

```go
// Serve spec
mux.HandleFunc("GET /api/spec", func(w http.ResponseWriter, r *http.Request) {
    http.ServeFile(w, r, "openapi.yaml")
})

// Serve Scalar UI
mux.HandleFunc("GET /api/docs", func(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "text/html")
    fmt.Fprint(w, `<!DOCTYPE html>
<html>
<head><title>Lurkarr API</title></head>
<body>
  <script id="api-reference" data-url="/api/spec"></script>
  <script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference"></script>
</body>
</html>`)
})
```

### Features

- Interactive "Try It" for all endpoints
- Auto-generated from the same openapi.yaml used by ogen
- Dark mode, search, code samples in 20+ languages
- No build step — single HTML page with CDN-loaded JS
