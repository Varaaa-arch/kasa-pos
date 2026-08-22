# ADR-006: API Error Model

**Status:** Accepted
**Date:** 2026-08-22

---

## Konteks

Sebelum ADR ini, semua error di API ditulis dengan `http.Error()`:

```go
http.Error(w, "product not found", http.StatusNotFound)
```

Response yang dihasilkan adalah plain text:

```
HTTP 404
Content-Type: text/plain; charset=utf-8

product not found
```

Masalah dengan pendekatan ini:

1. **Tidak bisa di-parse oleh client** — frontend harus string-match pesan error yang bisa berubah sewaktu-waktu
2. **Tidak ada kode error yang stabil** — tidak ada cara yang andal untuk membedakan "product not found" vs "transaction not found" tanpa mengandalkan HTTP status saja
3. **Tidak ada request tracing** — ketika user melaporkan error, tidak ada identifier untuk mengaitkan request ke log server
4. **Content-Type tidak konsisten** — endpoint sukses mengembalikan `application/json`, tapi error mengembalikan `text/plain`

## Keputusan

Semua error response menggunakan envelope JSON yang seragam:

```json
{
  "error": {
    "code": "PRODUCT_NOT_FOUND",
    "message": "Product not found",
    "request_id": "req_a1b2c3d4"
  }
}
```

### Struktur Model

```go
// ErrorCode adalah string konstanta yang stabil dan machine-readable.
type ErrorCode string

// errorResponse adalah top-level envelope.
type errorResponse struct {
    Error errorDetail `json:"error"`
}

// errorDetail berisi informasi error.
type errorDetail struct {
    Code      ErrorCode `json:"code"`
    Message   string    `json:"message"`
    RequestID string    `json:"request_id,omitempty"`
}
```

### Helper WriteError

Semua handler menggunakan satu fungsi untuk menulis error:

```go
WriteError(w, r, http.StatusNotFound, ErrCodeProductNotFound, "Product not found")
```

`WriteError` secara otomatis:
- Set `Content-Type: application/json`
- Tulis HTTP status code
- Encode envelope JSON
- Inject `request_id` dari context (diisi oleh `RequestID` middleware)

### RequestID Middleware

`RequestID` middleware dijalankan di setiap request:

```
Request masuk
    ↓
RequestID middleware
    ├── Baca X-Request-ID header (jika ada dari client)
    └── Generate req_XXXXXXXX jika tidak ada
          ↓
    Inject ke context (requestIDKey)
    Set X-Request-ID response header
          ↓
    Handler
          ↓
    WriteError membaca dari context → masuk ke JSON
```

Format ID yang di-generate: `req_` + 8 karakter hex lowercase.

## Error Codes

| ErrorCode | HTTP Status | Deskripsi |
|---|---|---|
| `INVALID_REQUEST_BODY` | 400 | Request body tidak valid JSON atau tidak bisa di-decode |
| `VALIDATION_ERROR` | 400 | Input valid JSON tapi gagal validasi bisnis (misal: SKU kosong) |
| `CART_EMPTY` | 400 | Checkout dikirim tanpa item |
| `INSUFFICIENT_PAYMENT` | 400 | Jumlah bayar kurang dari total |
| `INSUFFICIENT_STOCK` | 400 | Stok produk tidak cukup |
| `PRODUCT_NOT_FOUND` | 404 | Produk dengan ID yang diminta tidak ada |
| `TRANSACTION_NOT_FOUND` | 404 | Transaksi dengan ID yang diminta tidak ada |
| `INTERNAL_SERVER_ERROR` | 500 | Error tak terduga di sisi server |

## Contoh Response

**404 — Product not found:**

```http
HTTP/1.1 404 Not Found
Content-Type: application/json
X-Request-ID: req_a1b2c3d4

{
  "error": {
    "code": "PRODUCT_NOT_FOUND",
    "message": "Product not found",
    "request_id": "req_a1b2c3d4"
  }
}
```

**400 — Cart kosong:**

```http
HTTP/1.1 400 Bad Request
Content-Type: application/json
X-Request-ID: req_ff001122

{
  "error": {
    "code": "CART_EMPTY",
    "message": "Cart is empty",
    "request_id": "req_ff001122"
  }
}
```

**400 — Bayar kurang:**

```http
HTTP/1.1 400 Bad Request
Content-Type: application/json

{
  "error": {
    "code": "INSUFFICIENT_PAYMENT",
    "message": "Payment is insufficient",
    "request_id": "req_dd998877"
  }
}
```

**500 — Internal error:**

```http
HTTP/1.1 500 Internal Server Error
Content-Type: application/json

{
  "error": {
    "code": "INTERNAL_SERVER_ERROR",
    "message": "Failed to list products",
    "request_id": "req_cc112233"
  }
}
```

## Cara Pakai di Handler

```go
// Sebelum
http.Error(w, "product not found", http.StatusNotFound)

// Sesudah
WriteError(w, r, http.StatusNotFound, ErrCodeProductNotFound, "Product not found")
```

## Cara Pakai di Frontend / Client

Client bisa handle error berdasarkan `code`, bukan HTTP status saja:

```typescript
const res = await fetch('/products/xyz')
if (!res.ok) {
  const { error } = await res.json()
  if (error.code === 'PRODUCT_NOT_FOUND') {
    // tampilkan "Produk tidak ditemukan"
  }
  // log error.request_id untuk debugging
  console.error('Request ID:', error.request_id)
}
```

## Cara Testing

`assertErrorCode` tersedia di semua handler test sebagai helper:

```go
// Cukup satu baris untuk assert format error sepenuhnya:
assertErrorCode(t, rec, ErrCodeProductNotFound)

// assertErrorCode memvalidasi:
// 1. Content-Type: application/json
// 2. error.code sesuai expected
// 3. error.message tidak kosong
```

`decodeErrorResponse` tersedia di `errors_test.go` untuk decode envelope:

```go
resp := decodeErrorResponse(t, rec.Body.String())
if resp.Error.RequestID != "req_test999" {
    t.Fatalf(...)
}
```

## Konsekuensi

- **Positif:** Response format konsisten di semua endpoint; client tidak perlu string-match; error bisa di-trace via `request_id`
- **Positif:** `Content-Type: application/json` konsisten antara sukses dan error — frontend tidak perlu conditional parsing
- **Positif:** Menambah error code baru cukup tambah konstanta di `errors.go` dan panggil `WriteError`
- **Negatif:** `request_id` saat ini in-memory dan tidak di-persist ke log secara structured — untuk correlation penuh perlu structured logging (future work)

## Alternatif yang Ditolak

- **Gunakan error type dengan `Error()` string** — tidak cukup karena tidak ada kode stabil yang bisa di-parse client
- **Kembalikan HTTP status saja tanpa body** — tidak informatif untuk debugging dan tidak bisa dibedakan oleh client secara programmatik
- **Gunakan library error middleware eksternal** — over-engineering untuk kebutuhan saat ini; implementasi manual lebih transparan dan tanpa dependency tambahan
