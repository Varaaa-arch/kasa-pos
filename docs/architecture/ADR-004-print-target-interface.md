# ADR-004: PrintTarget Interface — PrintJob tidak boleh tahu Hardware

**Status:** Accepted
**Date:** 2026-08-14

---

## Konteks

`PrintJob` awalnya menerima `transport.Printer` (concrete USB printer type) sebagai parameter. Ini berarti `PrintJob` terikat pada implementasi hardware — sulit ditest dan melanggar prinsip dependency inversion.

## Keputusan

Definisikan interface kecil `PrintTarget` di package `receipt` (bukan di `transport`), karena interface sebaiknya didefinisikan oleh consumer:

```go
// internal/printer/receipt/print_target.go
type PrintTarget interface {
    Open() error
    Write([]byte) (int, error)
    Close() error
}
```

`PrintJob.Run()` dan `Print()` menerima `PrintTarget`, bukan `transport.Printer`.

```
PrintJob
  ↓
PrintTarget interface
  ↑
USBPrinter / MockPrinter / Future NetworkPrinter
```

## Konsekuensi

- `internal/printer/receipt` tidak lagi mengimpor `internal/printer/transport`
- `MockPrinter` dan `USBPrinter` keduanya satisfy interface ini secara implisit (Go structural typing)
- Mengganti transport (USB → Network) tidak butuh mengubah `PrintJob`
- Interface kecil (3 method) sesuai dengan prinsip Go: prefer small interfaces

## Alternatif yang Ditolak

- **Definisikan interface di `transport` package**: Interface sebaiknya di sisi consumer, bukan producer
- **Embed transport.Printer langsung**: Terlalu tight coupling ke concrete type
