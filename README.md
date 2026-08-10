# 🧾 KASA

> **Modern Point of Sale System — built from the cashier screen to the physical receipt.**

<p align="center">
  <img src="https://img.shields.io/badge/status-in%20development-F59E0B?style=for-the-badge" alt="Status">
  <img src="https://img.shields.io/badge/backend-Go-00ADD8?style=for-the-badge&logo=go&logoColor=white" alt="Go">
  <img src="https://img.shields.io/badge/frontend-Next.js-000000?style=for-the-badge&logo=next.js&logoColor=white" alt="Next.js">
  <img src="https://img.shields.io/badge/database-PostgreSQL-4169E1?style=for-the-badge&logo=postgresql&logoColor=white" alt="PostgreSQL">
  <img src="https://img.shields.io/badge/printer-ESC%2FPOS-333333?style=for-the-badge" alt="ESC/POS">
</p>

<p align="center">
  <strong>KASA</strong> is a local-first POS system designed to connect a modern cashier experience with real-world thermal printer hardware.
</p>

---

## ✦ Why KASA?

Most POS projects stop here:

```text
Product
   ↓
Cart
   ↓
Checkout
   ↓
Database
   ↓
Done.
```

KASA doesn't.

The interesting part begins when the transaction has to leave the screen:

```text
                    ┌──────────────┐
                    │    CASHIER   │
                    └──────┬───────┘
                           │
                           ▼
                    ┌──────────────┐
                    │   POS WEB    │
                    └──────┬───────┘
                           │
                           ▼
                    ┌──────────────┐
                    │   GO API     │
                    └──────┬───────┘
                           │
                  ┌────────┴────────┐
                  ▼                 ▼
            ┌──────────┐     ┌────────────┐
            │DATABASE  │     │ PRINT JOB  │
            └──────────┘     └─────┬──────┘
                                   │
                                   ▼
                           ┌───────────────┐
                           │ GO PRINT AGENT│
                           └───────┬───────┘
                                   │
                                  USB
                                   │
                                   ▼
                           ┌───────────────┐
                           │  BP-LITE58    │
                           │ 🧾 RECEIPT    │
                           └───────────────┘
```

> **A POS isn't finished when the button says "Paid".**
> **It's finished when the transaction is persisted and the physical workflow can reliably continue.**

---

# 🏗️ System Architecture

> **Status:** 🟠 Draft — architecture awal, belum final.
>
> Struktur ini akan terus berubah selama hardware research berlangsung.
> Detail komunikasi USB terhadap **BLUEPRINT BP-LITE58** belum dianggap final.

---

## Transaction Flow

```text
┌──────────────┐
│   POS Web    │
│              │
│  Cashier UI  │
└──────┬───────┘
       │
       │ HTTP
       ▼
┌──────────────┐
│    Go API    │
│              │
│ Business     │
│ Logic        │
└──────┬───────┘
       │
       │ SQL
       ▼
┌──────────────┐
│  PostgreSQL  │
│              │
│ Source of    │
│ Truth        │
└──────────────┘
```

---

## Print Flow

```text
┌──────────────┐
│   POS Web    │
└──────┬───────┘
       │
       │ localhost
       ▼
┌────────────────────┐
│  Go Print Agent    │
│                    │
│ • Print Queue      │
│ • ESC/POS Renderer │
│ • Printer Status   │
└─────────┬──────────┘
          │
          │ USB
          ▼
┌────────────────────┐
│ BLUEPRINT BP-LITE58│
│                    │
│ Thermal Printer    │
└────────────────────┘
```

---

# 🧩 Components

| Component | Responsibility |
|---|---|
| **POS Web** | Interface kasir untuk produk, cart, checkout, pembayaran, dan status transaksi. |
| **Go API** | Menangani business logic, transaksi, inventory, authentication, dan orchestration. |
| **PostgreSQL** | Persistent source of truth untuk data aplikasi. |
| **Go Print Agent** | Local service yang menjadi bridge antara POS dan printer fisik. |
| **ESC/POS Engine** | Mengubah receipt menjadi command yang dapat dimengerti thermal printer. |
| **BP-LITE58** | Target thermal receipt printer untuk hardware integration pertama. |

---

# 🖨️ Hardware Integration

## Target Hardware

```text
┌────────────────────────────────────┐
│          BLUEPRINT BP-LITE58       │
├────────────────────────────────────┤
│ Type       : Thermal Receipt       │
│ Paper      : ~57 mm                 │
│ Connection : USB + Bluetooth        │
│ Protocol   : ESC/POS                │
└────────────────────────────────────┘
```

> **Important:** KASA tidak mengasumsikan bagaimana printer bekerja hanya berdasarkan spesifikasi. Jalur USB akan ditentukan berdasarkan hasil eksperimen pada device sebenarnya.

Research:

```text
docs/printer/
├── spesifikasi.md
├── usb-research.md
├── hasil-pengujian.md
└── troubleshooting.md
```

---

# 🧾 Receipt Pipeline

Receipt tidak dibuat langsung oleh UI.

```text
Transaction
     │
     ▼
Receipt Data
     │
     ▼
Receipt Renderer
     │
     ▼
ESC/POS Commands
     │
     ▼
Print Agent
     │
     ▼
USB Transport
     │
     ▼
BP-LITE58
     │
     ▼
   🧾
REAL RECEIPT
```

Hal ini membuat printer layer bisa diganti tanpa mengubah business logic transaksi.

---

# 🔄 Transaction ≠ Printing

Printer adalah hardware.

Hardware bisa:

- offline
- kehabisan kertas
- disconnected
- busy
- gagal menerima data

Karena itu:

```text
              PAYMENT
                 │
                 ▼
        ┌─────────────────┐
        │ SAVE TRANSACTION│
        └────────┬────────┘
                 │
                 ▼
          CREATE PRINT JOB
                 │
          ┌──────┴──────┐
          ▼             ▼
       SUCCESS        FAILED
          │             │
          ▼             ▼
      COMPLETED       RETRY
```

### Prinsip utama

> **Printer failure must never silently become transaction failure.**

---

# 🗺️ Roadmap

## Phase 01 — 🔌 Hardware Recon

**Current**

- [ ] USB device discovery
- [ ] VID/PID identification
- [ ] Linux driver discovery
- [ ] Device node discovery
- [ ] USB interface research
- [ ] ESC/POS validation
- [ ] First successful print

---

## Phase 02 — 🧾 Print Engine

- [ ] Printer abstraction
- [ ] ESC/POS encoder
- [ ] Text
- [ ] Alignment
- [ ] Bold
- [ ] Font sizing
- [ ] Paper feed
- [ ] Paper cut
- [ ] Barcode
- [ ] QR Code
- [ ] Print status
- [ ] Retry handling

---

## Phase 03 — 🛒 POS Core

- [ ] Product management
- [ ] Categories
- [ ] Search
- [ ] Cart
- [ ] Checkout
- [ ] Cash payment
- [ ] Change calculation
- [ ] Transaction history
- [ ] Receipt printing

---

## Phase 04 — 📦 Operations

- [ ] Inventory
- [ ] Stock movement
- [ ] Barcode scanning
- [ ] Cashier session
- [ ] Roles & permissions
- [ ] Cash drawer tracking
- [ ] Reporting

---

## Phase 05 — ⚡ Reliability

- [ ] Local-first storage
- [ ] Offline transaction support
- [ ] Print queue
- [ ] Persistent retry
- [ ] Printer health monitoring
- [ ] Sync mechanism
- [ ] Conflict handling
- [ ] Recovery flows

---

# 📅 7-Day MVP Sprint

| Day | Focus | Definition of Done |
|---:|---|---|
| **01** | 🔌 Hardware Research | BP-LITE58 communication path understood |
| **02** | 🧾 Print PoC | Go successfully prints a real receipt |
| **03** | ⚙️ Receipt Engine | Reusable ESC/POS renderer |
| **04** | 🧠 Backend | Go API + PostgreSQL foundation |
| **05** | 🖥️ POS UI | Functional cashier flow |
| **06** | 🔗 Integration | Checkout → transaction → print |
| **07** | 🧪 Hardening | Error handling + reliable print flow |

### 7-Day Finish Line

```text
             PRODUCT
                │
                ▼
              CART
                │
                ▼
            CHECKOUT
                │
                ▼
             PAYMENT
                │
                ▼
        TRANSACTION SAVED
                │
                ▼
           PRINT JOB
                │
                ▼
         GO PRINT AGENT
                │
                ▼
               USB
                │
                ▼
           BP-LITE58
                │
                ▼
       ╔════════════════╗
       ║ 🧾 REAL RECEIPT║
       ╚════════════════╝
```

---

# 🛠️ Technology Stack

### Frontend

- **Next.js**
- **React**
- **TypeScript**
- **Tailwind CSS**
- **shadcn/ui**

### Backend

- **Go**
- REST API
- SQL

### Data

- **PostgreSQL**
- SQLite for local/offline components where appropriate

### Hardware

- **USB**
- **ESC/POS**
- **Go Print Agent**

### Infrastructure

- Docker
- GitHub Actions
- Prometheus / Grafana / Loki *(planned)*

> The stack is intentionally pragmatic. KASA does not introduce distributed infrastructure simply for the sake of complexity.

---

# 📂 Repository Structure

```text
kasa/
│
├── apps/
│   ├── web/
│   │   └── ...                  # POS Web
│   │
│   ├── api/
│   │   └── ...                  # Go API
│   │
│   └── print-agent/
│       └── ...                  # Local printer service
│
├── docs/
│   ├── architecture/
│   ├── printer/
│   └── api/
│
├── experiments/
│   └── hardware/
│
├── infra/
│   └── ...                      # Infrastructure
│
├── .github/
│   └── workflows/
│
├── README.md
└── LICENSE
```

---

# 🧪 Development Philosophy

### 01 — Research before abstraction

Jangan membuat abstraction berdasarkan asumsi hardware.

```text
Hardware
   ↓
Observe
   ↓
Experiment
   ↓
Understand
   ↓
Abstract
   ↓
Implement
```

---

### 02 — Keep the physical boundary explicit

```text
┌──────────────────────────────┐
│          POS DOMAIN          │
│                              │
│ Product / Cart / Transaction │
└──────────────┬───────────────┘
               │
          Print Contract
               │
               ▼
┌──────────────────────────────┐
│       HARDWARE DOMAIN        │
│                              │
│ Renderer / Queue / USB       │
└──────────────────────────────┘
```

Business logic should not know whether the printer is:

```text
USB
Bluetooth
Network
```

It should only know:

```text
"Print this receipt."
```

---

### 03 — Fail loudly, recover safely

Bad:

```text
Payment successful
Printer failed
Nothing happened
```

Good:

```text
✓ Payment successful

⚠ Receipt could not be printed.

[ Retry Printing ]
[ Continue ]
```

---

# 📌 Current Status

> **🟠 Early Development — Day 1**

The project is currently in the **hardware research stage**.

### Current priority

```text
BP-LITE58
    ↓
USB
    ↓
Fedora Linux
    ↓
USB Detection
    ↓
Communication Path
    ↓
Go
    ↓
ESC/POS
    ↓
🧾 First Print
```

Until this path is proven, the printer architecture remains **intentionally provisional**.

---

# 📚 Documentation

| Document | Purpose |
|---|---|
| [`docs/architecture/`](docs/architecture/) | System architecture & design decisions |
| [`docs/printer/`](docs/printer/) | Printer hardware research |
| [`docs/printer/usb-research.md`](docs/printer/usb-research.md) | USB communication investigation |
| [`docs/printer/hasil-pengujian.md`](docs/printer/hasil-pengujian.md) | Printer experiments & results |

---

# 🤝 Development

Conventional commit examples:

```text
feat: add printer USB discovery
feat: implement ESC/POS text printing
feat: add receipt renderer
feat: add transaction creation

fix: handle printer disconnection
fix: retry failed print jobs

docs: document BP-LITE58 USB interface
test: add receipt renderer tests

refactor: separate printer transport
```

---

# 📜 License

Licensed under the **Apache License 2.0**.

See [`LICENSE`](LICENSE) for details.

---

<p align="center">
  <strong>Built for the real world.</strong>
  <br />
  <sub>Web → Backend → Hardware → Receipt.</sub>
</p>

<p align="center">
  <strong>🧾 KASA</strong>
</p>
