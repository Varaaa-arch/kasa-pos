# ADR-002: Gunakan Money Type untuk Nominal Keuangan

**Status:** Accepted
**Date:** 2026-08-14

---

## Konteks

Semua field nominal keuangan (harga, subtotal, total, kembalian) awalnya disimpan sebagai `int64`. Ini rawan bug: operasi antar currency bisa terjadi tanpa deteksi, formatting harus dilakukan manual di setiap titik, dan tidak ada enkapsulasi logika aritmetika.

## Keputusan

Definisikan `Money` sebagai tipe domain di `internal/domain/receipt/money.go`:

```go
type Money struct {
    Amount   int64
    Currency Currency
}
```

Dengan method:
- `Add`, `Sub`, `Mul` — aritmetika, panik kalau currency berbeda
- `IsZero`, `IsNegative` — guard condition
- `Equal`, `Compare`, `LessThan`, dll — perbandingan
- `String` / `Format` — formatting canonical `Rp15.000`

## Konsekuensi

- Cross-currency arithmetic langsung ketahuan saat runtime (panic)
- `"Rp" + formatMoney(x)` tidak lagi tersebar di mana-mana — `x.String()` cukup
- Zero-value `Money{}` punya `Currency = ""`, bukan `IDR`. Calculator menggunakan `moneyOrZero()` untuk normalize field optional yang tidak di-set
- Test menggunakan `assertMoney(t, got, wantAmount, wantCurrency)` agar assertion eksplisit terhadap currency

## Alternatif yang Ditolak

- **`float64`**: Floating point tidak cocok untuk keuangan karena presisi
- **Third-party money library**: Overkill untuk satu currency (IDR). Bisa diadopsi nanti kalau multi-currency dibutuhkan
