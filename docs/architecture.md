# 🏗️ System Architecture

> **Status:** Draft — arsitektur awal, belum final.
> Struktur ini kemungkinan besar akan berubah setelah riset USB terhadap **BP-LITE58** selesai (lihat `docs/printer/usb-research.md`).

---

## Planned Architecture

### Transaction Flow

```
┌────────────┐        HTTP        ┌────────────┐        ┌────────────┐
│  POS Web   │ ─────────────────► │   Go API   │ ─────► │ PostgreSQL │
└────────────┘                    └────────────┘        └────────────┘
```

### Print Flow

```
┌────────────┐     localhost     ┌────────────────┐     USB     ┌──────────────────┐
│  POS Web   │ ────────────────► │ Go Print Agent  │ ──────────► │ BLUEPRINT BP-LITE58 │
└────────────┘                   └────────────────┘             └──────────────────┘
```

---

## Components

| Component | Deskripsi |
|---|---|
| **POS Web** | Frontend yang digunakan kasir untuk melakukan transaksi. |
| **Go API** | Backend yang menangani business logic dan transaksi. |
| **PostgreSQL** | Database utama aplikasi POS. |
| **Go Print Agent** | Local service yang menangani komunikasi antara POS dengan thermal printer. |
| **BP-LITE58** | Thermal receipt printer yang digunakan untuk mencetak struk transaksi. |

---

## Catatan

Ini **hanya arsitektur awal**. Detail komunikasi USB, protokol printer, dan struktur `Go Print Agent` **masih bisa berubah total** setelah kita mengetahui bagaimana **BP-LITE58** benar-benar berkomunikasi melalui USB.

Jangan anggap dokumen ini final — treat as living document. 