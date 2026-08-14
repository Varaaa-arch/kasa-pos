# ADR-003: Print Job sebagai State Machine

**Status:** Accepted
**Date:** 2026-08-14

---

## Konteks

Printer adalah hardware — bisa offline, kehabisan kertas, atau disconnected. Kalau print langsung dilakukan sebagai side effect dari payment, kegagalan printer bisa terlihat seperti kegagalan transaksi. Keduanya harus dipisahkan.

## Keputusan

Definisikan `PrintJob` sebagai state machine dengan status eksplisit:

```
PENDING → RUNNING → COMPLETED
                 ↘ FAILED
```

```go
type PrintJob struct {
    ID        string
    Receipt   domainreceipt.Receipt
    Status    PrintJobStatus
    CreatedAt time.Time
    UpdatedAt time.Time
    Error     error
}
```

Transisi hanya boleh terjadi via method (`Start`, `Complete`, `Fail`). Transisi yang tidak valid mengembalikan error.

`PrintJob.Run(printer, renderer)` mengatur seluruh lifecycle: start → print → complete/fail.

## Prinsip utama

> **Printer failure must never silently become transaction failure.**

Transaksi disimpan dulu. Print job dibuat setelah itu. Kalau print gagal, job berstatus `FAILED` dan bisa di-retry — transaksi tetap valid.

## Konsekuensi

- Status print job selalu eksplisit dan bisa di-observe
- Retry bisa dilakukan tanpa mengulang transaksi
- Test bisa verifikasi lifecycle tanpa menyentuh printer nyata
