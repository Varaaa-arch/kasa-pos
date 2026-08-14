# ADR-001: Pisahkan Domain Receipt dari Printer Layer

**Status:** Accepted
**Date:** 2026-08-14

---

## Konteks

Awalnya data receipt didefinisikan langsung di dalam package printer. Ini membuat business logic receipt terikat pada implementasi printer, sehingga sulit ditest secara terpisah dan susah diganti kalau printer berubah.

## Keputusan

Buat package `internal/domain/receipt` yang menjadi satu-satunya sumber kebenaran untuk model data receipt:

```
internal/
  domain/
    receipt/        ← domain model, tidak tahu printer
      receipt.go
      money.go
  printer/
    receipt/        ← printer logic, tahu domain
```

Package domain tidak boleh mengimpor apapun dari `printer/`.

## Konsekuensi

- Receipt domain bisa ditest tanpa menyentuh printer
- Printer layer bergantung ke domain, bukan sebaliknya
- Kalau model receipt berubah, perubahan terisolasi di `domain/receipt`
