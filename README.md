# KASA

**Modern Point of Sale (dari layar kasir sampai struk fisik).**

KASA adalah sistem POS lokal yang menghubungkan checkout web, transaksi atomik di database, dan cetak struk thermal via Print Agent.

<p align="center">
  <img src="https://img.shields.io/badge/Go-00ADD8?style=flat-square&logo=go&logoColor=white" alt="Go">
  <img src="https://img.shields.io/badge/Next.js-000?style=flat-square&logo=next.js&logoColor=white" alt="Next.js">
  <img src="https://img.shields.io/badge/PostgreSQL-4169E1?style=flat-square&logo=postgresql&logoColor=white" alt="PostgreSQL">
  <img src="https://img.shields.io/badge/ESC%2FPOS-333?style=flat-square" alt="ESC/POS">
</p>

---

## Kenapa KASA?

Banyak POS berhenti di **checkout + database**. KASA melanjutkan ke **struk fisik**:

```text
Cashier UI  →  Go API  →  PostgreSQL (commit)
                    ↓
              Print Agent  →  USB  →  BP-LITE58
```

Transaksi **tetap sukses** meski printer gagal. Cetak terjadi **setelah** commit — tidak di dalam SQL transaction.

---

## Tech Stack

| Layer | Teknologi |
|-------|-----------|
| Frontend | Next.js, TypeScript, Tailwind |
| Backend | Go, REST API |
| Database | PostgreSQL |
| Printer | ESC/POS, Go Print Agent, USB |

**Target printer:** BLUEPRINT BP-LITE58 (58mm thermal)

---

## Arsitektur Singkat

```text
┌─────────────┐     ┌─────────────┐     ┌──────────────┐
│  POS Web    │────▶│  POS API    │────▶│ PostgreSQL   │
│  :3000      │     │  :8080      │     │              │
└─────────────┘     └──────┬──────┘     └──────────────┘
                           │ commit OK
                           ▼
                    ┌─────────────┐     ┌──────────────┐
                    │ Print Agent │────▶│ BP-LITE58    │
                    │  :8081      │ USB │ (thermal)    │
                    └─────────────┘     └──────────────┘
```

| Service | Port | Fungsi |
|---------|------|--------|
| POS Web | `3000` | UI kasir |
| POS API | `8080` | Produk, checkout, transaksi |
| Print Agent | `8081` | Render struk & kirim ke printer |

---

## Quick Start

### 1. Database

```bash
docker run -d --name pos-postgres \
  -e POSTGRES_USER=pos -e POSTGRES_PASSWORD=pos -e POSTGRES_DB=pos \
  -p 5434:5432 postgres:17
```

Jalankan migrasi (lihat folder `migrations/`).

### 2. Environment

```bash
cp .env.example .env
```

```env
DATABASE_URL=postgres://pos:pos@localhost:5434/pos?sslmode=disable
PRINT_AGENT_URL=http://127.0.0.1:8081
```

### 3. Jalankan services

```bash
# Terminal 1 — API
go run ./cmd/pos-api

# Terminal 2 — Print Agent (butuh printer USB atau POC)
go run ./experiments/printer/print-agent

# Terminal 3 — Web
cd apps/web
NEXT_PUBLIC_API_URL=http://localhost:8080 npm run dev
```

Buka **http://localhost:3000**

---

## Running Services

### POS API

```bash
# Development
go run ./cmd/pos-api

# Production build
go build -o bin/pos-api ./cmd/pos-api
./bin/pos-api
```

**API Endpoints:**
- `GET /products` - List all products
- `POST /products` - Create new product
- `GET /products/:id` - Get product by ID
- `PUT /products/:id` - Update product
- `DELETE /products/:id` - Delete product
- `POST /checkout` - Process checkout and print receipt
- `GET /transactions` - List transactions
- `GET /transactions/:id` - Get transaction details
- `POST /transactions/:id/reprint` - Reprint receipt

### Print Agent

```bash
# Development
go run ./experiments/printer/print-agent

# Production build
go build -o bin/print-agent ./experiments/printer/print-agent
./bin/print-agent
```

**Print Agent Endpoints:**
- `POST /print` - Print receipt with idempotency key
- `GET /status` - Check print agent status
- `GET /health` - Health check endpoint

**Environment Variables:**
- `PRINTER_DEVICE` - USB device path (default: `/dev/usb/lp0`)
- `LISTEN_ADDRESS` - Server address (default: `127.0.0.1:8081`)

---

## Testing

### Unit Tests

```bash
# Run all tests
go test ./...

# Run specific package tests
go test ./internal/service/checkout

# Run with verbose output
go test -v ./...

# Run with race detection
go test -race ./...

# Run with coverage
go test -cover ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Integration Tests

```bash
# Run integration tests (requires DATABASE_URL)
export DATABASE_URL="postgres://pos:pos@localhost:5434/pos?sslmode=disable"
go test -v ./internal/api
go test -v ./internal/service/checkout -run TestEndToEnd
```

### Failure Injection Tests

```bash
# Run all failure injection tests
go test -v ./internal/service/checkout -run TestFailure

# Run specific failure scenario
go test -v ./internal/service/checkout -run TestFailure_PrinterDisconnected
go test -v ./internal/service/checkout -run TestFailure_PostgreSQLDown
go test -v ./internal/service/checkout -run TestFailure_ConcurrentCheckout
```

**Failure Scenarios Covered:**
- Printer disconnected
- Print agent down
- Network timeout
- PostgreSQL down
- Insufficient stock
- Concurrent checkout
- Print agent idempotency
- Cascading failures

### Full E2E Regression Tests

```bash
# Requires DATABASE_URL and real print agent setup
export DATABASE_URL="postgres://pos:pos@localhost:5434/pos?sslmode=disable"
go test -v ./internal/service/checkout -run TestFullE2ERegression
```

### Web Tests

```bash
cd apps/web
npm run lint
npx tsc --noEmit
npm run build
npm run test:e2e   # butuh API + web sudah running
```

---

## Failure Simulation

### Simulating Printer Failure

```bash
# Stop print agent to simulate printer/agent failure
# Transaction will succeed but print job will fail
curl -X POST http://localhost:8080/checkout \
  -H "Content-Type: application/json" \
  -d '{
    "items": [{"product_id": "uuid", "quantity": 1}],
    "paid_amount": 10000,
    "payment_method": "CASH"
  }'
```

### Simulating Database Failure

```bash
# Stop PostgreSQL to simulate database failure
docker stop pos-postgres

# Checkout will fail with database connection error
curl -X POST http://localhost:8080/checkout \
  -H "Content-Type: application/json" \
  -d '{
    "items": [{"product_id": "uuid", "quantity": 1}],
    "paid_amount": 10000,
    "payment_method": "CASH"
  }'
```

### Simulating Network Timeout

```bash
# Use slow print agent to simulate timeout
# Implementation depends on your test setup
```

---

## Recovery Procedures

### Print Agent Recovery

If print agent fails:

1. **Check print agent status:**
```bash
curl http://localhost:8081/health
```

2. **Restart print agent:**
```bash
# Stop current process
# Start again
go run ./experiments/printer/print-agent
```

3. **Reprint failed receipts:**
```bash
curl -X POST http://localhost:8080/transactions/:id/reprint
```

### Database Recovery

If database fails:

1. **Check database connection:**
```bash
docker ps | grep pos-postgres
docker logs pos-postgres
```

2. **Restart database:**
```bash
docker restart pos-postgres
```

3. **Run migrations if needed:**
```bash
# Check migrations folder for specific commands
```

### API Recovery

If API fails:

1. **Check API logs for errors**
2. **Verify environment variables**
3. **Restart API service**
4. **Check database connectivity**

---

## Troubleshooting

### Common Issues

**Printer not detected:**
```bash
# Check USB devices
lsusb
# Check printer device
ls -la /dev/usb/lp0
# Adjust permissions if needed
sudo chmod 666 /dev/usb/lp0
```

**Database connection failed:**
```bash
# Check PostgreSQL is running
docker ps | grep postgres
# Check connection string in .env
# Test connection
psql postgres://pos:pos@localhost:5434/pos
```

**Print agent timeout:**
```bash
# Check print agent is running
curl http://localhost:8081/health
# Check network connectivity
curl http://localhost:8081/status
```

**Concurrent checkout issues:**
```bash
# Run concurrent tests to verify locking
go test -v ./internal/service/checkout -run TestConcurrent
```

---

## Documentation

Detailed documentation is available in the `docs/` directory:

- **[docs/architecture/](docs/architecture/)** - Architecture Decision Records (ADRs) and system architecture
- **[docs/api/](docs/api/)** - API documentation, endpoints, and usage examples
- **[docs/printer/](docs/printer/)** - Printer specifications, ESC/POS implementation, and hardware research

---

## Project Structure

```text
cmd/pos-api/              # REST API server
experiments/printer/      # Print Agent & hardware POC
internal/
  api/                    # HTTP handlers
  service/checkout/       # Atomic checkout + print orchestration
  printer/                # ESC/POS, receipt, print agent client
apps/web/                 # Next.js cashier UI
docs/                     # Architecture, API, and printer documentation
migrations/               # SQL migrations
```

---

## Checkout Flow

1. Kasir tambah produk ke keranjang
2. `POST /checkout` → atomic commit + stock reduction
3. Receipt dibuat dari transaksi
4. Print Agent menerima `POST /print` dengan idempotency key
5. Response `201` + `print_job.status` (`COMPLETED` atau `FAILED`)

---

## Status

✅ **Done V1.0** 
