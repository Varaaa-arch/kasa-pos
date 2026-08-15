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

## Checkout Flow

1. Kasir tambah produk ke keranjang
2. `POST /checkout` → atomic commit + stock reduction
3. Receipt dibuat dari transaksi
4. Print Agent menerima `POST /print` dengan idempotency key
5. Response `201` + `print_job.status` (`COMPLETED` atau `FAILED`)

---

## Testing

```bash
# Go
go test ./...
go test -race ./...

# Web
cd apps/web
npm run lint
npx tsc --noEmit
npm run build
npm run test:e2e   # butuh API + web sudah running
```

---

## Struktur Project

```text
cmd/pos-api/              # REST API server
experiments/printer/      # Print Agent & hardware POC
internal/
  api/                    # HTTP handlers
  service/checkout/       # Atomic checkout + print orchestration
  printer/                # ESC/POS, receipt, print agent client
apps/web/                 # Next.js cashier UI
docs/                     # Arsitektur & riset printer
migrations/               # SQL migrations
```

---

## Status

🟡 **In development** core checkout, stock, transaksi, dan integrasi Print Agent sudah berjalan. Autentikasi dan fitur lanjutan masih dalam roadmap.
