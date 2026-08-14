# ADR-005: Idempotency Key untuk POST /print

**Status:** Accepted
**Date:** 2026-08-14

---

## Konteks

`POST /print` adalah operasi yang punya side effect fisik: printer mengeluarkan kertas. Kalau request dikirim dua kali (network retry, double-click, client timeout), struk bisa tercetak dua kali. Ini tidak bisa di-undo.

## Keputusan

Wajibkan header `Idempotency-Key` di setiap `POST /print`. Request pertama dengan key tertentu diproses; request berikutnya dengan key yang sama ditolak dengan `409 Conflict`.

```
POST /print
  ↓
Idempotency-Key header (required)
  ↓
Claim key di IdempotencyStore
  ├── key sudah ada → 409 Conflict
  └── key baru
        ↓
        PrintJob
        ↓
        Run
        ↓
        SetStatus(key, job.Status)
```

`IdempotencyStore` menyimpan status per key:

```go
type IdempotencyStore struct {
    mu      sync.Mutex
    claimed map[string]PrintJobStatus
}
```

Status di-update di 3 titik: setelah job dibuat (`PENDING`), setelah error (`FAILED`), setelah sukses (`COMPLETED`).

## Response codes

| Kondisi | Status |
|---|---|
| Key baru, print berhasil | `200 OK` |
| Key duplikat | `409 Conflict` |
| Idempotency store nil | `500 Internal Server Error` |
| Key tidak ada di header | `400 Bad Request` |

## Konsekuensi

- Double-submit aman: struk tidak tercetak dua kali
- Client harus generate unique key per request (UUID, invoice number, dll)
- Store saat ini in-memory — restart server akan reset semua key. Untuk persistent idempotency, store perlu dipindah ke database (future work)

## Alternatif yang Ditolak

- **Cek duplikat berdasarkan invoice number**: Invoice number bisa sama kalau ada bug di client. Key eksplisit lebih deterministic
- **No idempotency**: Tidak acceptable untuk operasi yang punya physical side effect
