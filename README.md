# Calculator Services

HTTP API for common arithmetic operations: addition, subtraction, multiplication, division, exponentiation, square root, and percentage.

## Setup

Requires [Go 1.24+](https://go.dev/dl/).

```bash
git clone https://github.com/busrau/calc_services.git
cd calc_services
go test ./...
go run ./cmd/server
```

The server listens on `:8080` by default. Override the port with `PORT`:

```bash
# Windows PowerShell
$env:PORT="9090"; go run ./cmd/server

# Unix
PORT=9090 go run ./cmd/server
```

Build a binary:

```bash
go build -o calc-server ./cmd/server
./calc-server
```

## API usage

All operation endpoints accept `POST` with a JSON body and return JSON. Numbers are IEEE-754 `float64` values.

Success:

```json
{
  "operation": "add",
  "operands": [2, 3],
  "result": 5
}
```

Error:

```json
{
  "error": "division by zero"
}
```

| Method | Path | Body | Result |
| --- | --- | --- | --- |
| `GET` | `/health` | — | `{"status":"ok"}` |
| `POST` | `/v1/add` | `{"a":2,"b":3}` | `a + b` |
| `POST` | `/v1/subtract` | `{"a":10,"b":4}` | `a - b` |
| `POST` | `/v1/multiply` | `{"a":3,"b":7}` | `a * b` |
| `POST` | `/v1/divide` | `{"a":9,"b":3}` | `a / b` |
| `POST` | `/v1/power` | `{"a":2,"b":10}` | `a ^ b` |
| `POST` | `/v1/sqrt` | `{"a":81}` | `√a` |
| `POST` | `/v1/percentage` | `{"value":80,"percent":25}` | `percent%` of `value` |

Examples:

```bash
curl -s http://localhost:8080/health

curl -s -X POST http://localhost:8080/v1/add \
  -H "Content-Type: application/json" \
  -d "{\"a\":2,\"b\":3}"

curl -s -X POST http://localhost:8080/v1/divide \
  -H "Content-Type: application/json" \
  -d "{\"a\":10,\"b\":0}"

curl -s -X POST http://localhost:8080/v1/sqrt \
  -H "Content-Type: application/json" \
  -d "{\"a\":16}"

curl -s -X POST http://localhost:8080/v1/percentage \
  -H "Content-Type: application/json" \
  -d "{\"value\":200,\"percent\":10}"
```

### Validation and status codes

| Situation | Status |
| --- | --- |
| Valid request | `200` |
| Malformed JSON, unknown fields, missing operands, non-numeric values | `400` |
| Division by zero | `400` |
| Square root of a negative number | `400` |
| Power that is undefined in real numbers (for example `(-4)^0.5`) | `400` |
| Overflow (`Inf`) | `400` |
| Unexpected calculator failure | `500` |
| Wrong HTTP method | `405` |

Zero is a valid operand except as a divisor. Missing fields are distinct from zero: `{"a":0,"b":0}` is accepted for addition; `{"a":1}` is rejected.

Percentage is defined as `value * percent / 100`. For example, 25% of 80 is `20`.

## Design rationale

The service is split into two layers so math rules stay independent of HTTP:

- `internal/calculator` owns the operations and domain errors (division by zero, negative square roots, non-finite inputs and results).
- `internal/httpapi` owns transport concerns: routing, JSON decoding, required-field checks, and mapping domain errors to HTTP status codes.

Handlers depend on a `Calculator` interface rather than the concrete service. That keeps HTTP tests able to inject a stub for unexpected errors while using the real calculator for request validation and happy paths.

The HTTP layer uses the standard library (`net/http`, `encoding/json`) and Go 1.22 method-aware routing. Request bodies are size-capped and unknown JSON fields are rejected so clients get a clear `400` instead of silently ignored input.

Operations live on dedicated endpoints instead of a single catch-all `operation` field. Each route has a stable contract (binary `a`/`b`, unary `a`, or named `value`/`percent`), which keeps validation simple and documentation explicit.
