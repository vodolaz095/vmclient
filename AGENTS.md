# vmclient - Agent Guide

This document contains essential knowledge for AI agents working in the vmclient repository.

## Project Overview

Go module (`github.com/vodolaz095/vmclient`) providing a client for Victoria Metrics. Supports:
- Pushing metrics (gauge, counter, metric sets)
- Instant queries
- Range queries
- Connection health checking (ping)

Key dependencies:
- `github.com/VictoriaMetrics/metrics` - for metric handling
- OpenTelemetry (`go.opentelemetry.io/...`) - for tracing

## Essential Commands

```bash
# Build and verify dependencies
make deps

# Run tests
make test

# Run vulnerability check (requires govulncheck)
make vuln

# Run example applications
make push        # Run push example
make instant     # Run instant query example
make range       # Run range query example
```

## Code Organization

- `client.go`: Main `Client` struct and `New()` constructor
- `config.go`: `Config` struct for client configuration
- `constants.go`: Shared constants (endpoints, labels, defaults)
- `do.go`: Core HTTP request logic with URL building
- `errors.go`: Custom error types and response handling
- `ping.go`: Health check implementation
- `push.go`: Metrics pushing functionality
- `instant.go`: Instant query implementation
- `range.go`: Range query implementation
- `model.go`: Data models for query results

## Architecture and Data Flow

### Client Initialization

1. `New(ctx, cfg)` creates a `Client` with configuration
2. Uses `cfg.HttpClient` if provided, otherwise creates HTTP client with OpenTelemetry instrumentation
3. Performs `Ping()` to verify connectivity
4. Returns initialized client

### Request Flow

All operations follow this pattern:
1. Method (e.g., `Push`, `Instant`, `Range`) creates OpenTelemetry span
2. Calls `c.do()` with operation type and parameters
3. `c.do()` builds appropriate URL and query parameters
4. Performs HTTP request with headers and tracing
5. Returns response for processing

### Operation-Specific Flows

**Push**: 
- `Push()` → `c.do("push", ...)` → Victoria Metrics `/api/v1/import/prometheus`

**Instant Query**:
- `Instant()` → `c.do("instant", ...)` → `/prometheus/api/v1/query` with `time` parameter

**Range Query**:
- `Range()` → `c.do("range", ...)` → `/prometheus/api/v1/query_range` with `start`/`end` parameters

**Ping**:
- `Ping()` → `c.do("ping", ...)` → `/-/healthy` endpoint

## Key Patterns and Conventions

### OpenTelemetry Tracing

All methods use OpenTelemetry tracing with consistent patterns:
- Spans named after operation (`ping`, `push`, `instant`, `range`)
- Span kind: `trace.SpanKindClient`
- Common attributes:
  - `db.client.connection.pool.name` (endpoint)
  - `db.system` ("Victoria Metrics")
- Status set to `codes.Ok` on success, `codes.Error` on failure
- Errors recorded with `span.RecordError()`

### Error Handling

- Custom `Err` struct wraps underlying errors with additional context
- `handleErrorResponse()` processes HTTP responses with status != 200
- Specific error types:
  - `ErrUnexpectedResponse` - non-200 status codes
  - `ErrQueryError` - query processing errors (422 status)
- Errors include HTTP status code, message, and raw response

### URL Construction

The `do()` function handles all URL building:
- Uses `url.JoinPath()` for path construction
- Builds query parameters with `url.Values`
- Sets `timeout` parameter from context deadline when present

## Testing Approach

- Uses standard Go testing with `testing` package
- Leverages `github.com/jarcoal/httpmock` for HTTP mocking
- Uses `github.com/stretchr/testify` for assertions
- Example applications in `/examples` directory demonstrate usage

## Skills

The project includes several skills in the `/skills` directory that provide guided workflows for common tasks:

### lint

**Purpose**: Run linting on the Go codebase and automatically fix issues

**Usage**:
1. Verify required tools (`go`, `gopls`, `staticcheck`) are available
2. Run `make lint` to identify issues
3. Analyze output and apply fixes
4. Re-run linting to verify fixes

**Requirements**:
- Must verify tools before running
- Must run `make lint` as primary command
- Must attempt to fix all reported issues
- Must verify fixes by re-running lint

**Dependencies**:
- Go compiler
- Go language server (gopls)
- Staticcheck tool
- Make build system

### vuln-scan

**Purpose**: Scan for known vulnerabilities using `govulncheck`

**Usage**:
```bash
make tools
make deps
make vuln
```

**Fixing Vulnerabilities**:
- Update dependencies using `go get -u <module>@latest`
- Or update to specific version: `go get <module>@<fixed_version>`
- Run `make deps` to synchronize changes
- Re-run `make vuln` to verify resolution

**Workflow**:
1. Ensure tools available (`make tools`)
2. Download dependencies (`make deps`)
3. Run scan (`make vuln`)
4. Identify vulnerable modules
5. Update module with `go get`
6. Synchronize with `make deps`
7. Verify resolution with `make vuln`

**Requirements**:
- `govulncheck` installed (`go install golang.org/x/vuln/cmd/govulncheck@latest`)
- Go 1.26+
- Internet access for vulnerability database

**Notes**:
- Verify updates don't break functionality
- Run tests after updates
- For production, prefer pinning to specific patch versions over `@latest`

### run-tests

**Purpose**: Run unit tests for the project

**Usage**:
```bash
make test
```

**Additional Options**:
- `make cover`: Run tests with coverage reporting
- `make check`: Run linting before tests

**Prerequisites**:
```bash
make tools  # Check for go, golint, govulncheck
make deps   # Download dependencies
```

## Important Gotchas

### Insecure TLS Configuration

Setting `Config.Insecure = true` modifies `http.DefaultTransport` globally:
```go
http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
```
This affects all HTTP clients in the process, not just this client instance.

### Label Handling

- `LabelForName` constant is "`__name__`" (double underscore)
- `labelsToString()` function formats labels with sorted key order
- `ExtraLabels` from config are applied to all pushed metrics

### Time Handling

- Time parameters are converted to Unix timestamps (seconds) for API calls
- `Instant` and `Range` methods expect `time.Time` parameters
- `step` parameter uses `time.Duration` (default: 5 minutes)

### HTTP Method Consistency

Despite the Victoria Metrics API documentation suggesting POST for some endpoints, this client uses GET for all query operations (instant and range) as shown in the code.

## Configuration

The `Config` struct supports:
- `Address`: Base URL for Victoria Metrics (default: `http://127.0.0.1:8428`)
- `Headers`: Map of HTTP headers to include in requests
- `ExtraLabels`: Labels to add to all pushed metrics
- `HttpClient`: Optional custom HTTP client
- `Insecure`: Skip TLS verification (affects global transport)

## Examples

Example applications in `/examples` demonstrate:
- `push/main.go`: Pushing gauge, counter, and metric set
- `instant/main.go`: Performing instant query
- `range/main.go`: Performing range query

These can be run with `make push`, `make instant`, `make range` respectively.
