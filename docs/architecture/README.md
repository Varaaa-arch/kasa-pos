# Architecture Decision Records

Dokumen ini mencatat keputusan arsitektur yang sudah dibuat selama development KASA. Setiap ADR bersifat immutable — kalau keputusan berubah, buat ADR baru yang supersede ADR lama.

---

| ADR | Judul | Status |
|---|---|---|
| [ADR-001](ADR-001-domain-receipt-model.md) | Pisahkan Domain Receipt dari Printer Layer | Accepted |
| [ADR-002](ADR-002-money-type.md) | Gunakan Money Type untuk Nominal Keuangan | Accepted |
| [ADR-003](ADR-003-print-job-lifecycle.md) | Print Job sebagai State Machine | Accepted |
| [ADR-004](ADR-004-print-target-interface.md) | PrintTarget Interface — PrintJob tidak boleh tahu Hardware | Accepted |
| [ADR-005](ADR-005-print-idempotency.md) | Idempotency Key untuk POST /print | Accepted |
| [ADR-006](ADR-006-api-error-model.md) | API Error Model — Structured JSON Error Response | Accepted |

---

## Format ADR

```
# ADR-XXX: Judul

**Status:** Proposed | Accepted | Deprecated | Superseded by ADR-XXX
**Date:** YYYY-MM-DD

## Konteks
Masalah atau situasi yang memaksa keputusan ini.

## Keputusan
Apa yang diputuskan dan bagaimana bentuknya.

## Konsekuensi
Trade-off, dampak positif dan negatif dari keputusan ini.
```
