cat > hasil-pengujian-printer.md << 'EOF'
# Hasil Pengujian Printer

## Day 1 — USB & Hardware

### USB Device

Printer berhasil terdeteksi oleh Fedora melalui USB.

```text
Bus 003 Device 005: ID 28e9:0289 GDMicroelectronics micro-printer
```

Device descriptor:

```text
Vendor ID:  0x28e9
Product ID: 0x0289
Manufacturer: GEZHI
Product: micro-printer
Serial: 000000000004
USB Version: 1.10
Speed: Full Speed (12Mbps)
```

Printer menggunakan USB Printer Class:

```text
Interface Class:    7 Printer
Interface SubClass: 1 Printer
Interface Protocol: 2 Bidirectional
```

Endpoint:

```text
OUT: 0x01
IN:  0x81
Transfer: Bulk
Max Packet Size: 64 bytes
```

### Linux Device

Fedora mengenali printer melalui:

```text
/dev/usb/lp0
```

Kernel module:

```text
usblp
```

Status:

```text
$ lsmod | grep usblp

usblp
```

Device permission:

```text
crw-rw----. 1 root lp ... /dev/usb/lp0
```

User telah ditambahkan ke group:

```text
lp
```

### USB Communication

Komunikasi dilakukan melalui:

```text
Go
 ↓
/dev/usb/lp0
 ↓
Linux usblp
 ↓
USB
 ↓
Printer
```

Printer menerima raw ESC/POS byte stream.

## Day 2 — Go Print Agent

### Printer Transport

Implementasi transport USB dibuat menggunakan:

```go
os.OpenFile("/dev/usb/lp0", os.O_WRONLY, 0)
```

Transport menyediakan operasi:

```text
Open()
Write()
Close()
```

Test:

```bash
go test ./internal/printer/transport
```

Hasil:

```text
PASS
```

### ESC/POS Core

ESC/POS command dasar berhasil diimplementasikan dan diuji.

Command yang berhasil:

```text
Initialize
Text
Line Feed
Bold
Alignment
Font Size
Paper Feed
```

Printer berhasil menerima dan mengeksekusi command tersebut.

### Basic Print Test

Basic text printing berhasil.

Contoh:

```text
HELLO WORLD
```

Printer berhasil mencetak text secara fisik.

### Receipt Renderer

Receipt data model dibuat untuk merepresentasikan:

```text
Store
Transaction
Items
Summary
Payment
Footer
```

Flow:

```text
Receipt Data
    ↓
Receipt Renderer
    ↓
ESC/POS []byte
```

Renderer menghasilkan ESC/POS byte stream yang siap dikirim ke printer.

Unit test:

```bash
go test ./internal/printer/receipt
```

Hasil:

```text
PASS
```

### Prototype Receipt

Prototype receipt berhasil dicetak secara fisik.

Receipt berisi:

```text
Store Information
Invoice
Timestamp
Cashier
Items
Subtotal
Total
Payment
Change
Footer
```

Contoh data:

```text
TOKO KASA

Invoice: INV-000001
Kasir: Bizar

Kopi Susu
2 x Rp 15.000

Roti Bakar
1 x Rp 12.000

Air Mineral
1 x Rp 5.000

TOTAL
Rp 47.000

Bayar
Rp 50.000

Kembali
Rp 3.000

TERIMA KASIH
```

### USB Write Reliability

Single large write menghasilkan masalah di printer: printer hanya memproses sebagian receipt.

Untuk meningkatkan reliability, USB transport diubah menjadi chunked write:

```text
Receipt bytes
     ↓
64-byte chunks
     ↓
10ms delay
     ↓
/dev/usb/lp0
```

Dengan metode tersebut, prototype receipt berhasil dicetak secara lengkap.

Implementasi ini menggunakan ukuran chunk 64 byte yang sesuai dengan:

```text
wMaxPacketSize: 64 bytes
```

### Printer Disconnect Test

Printer diuji dalam kondisi USB disconnect.

Test:

```text
Printer connected
        ↓
USB cable disconnected
        ↓
Write operation
        ↓
Write failure detected
```

Setelah USB disambungkan kembali:

```text
USB reconnect
      ↓
/dev/usb/lp0 available
      ↓
Printer usable again
```

Hasil:

```text
PASS
```
# Hasil Pengujian Printer BP-LITE58

Dokumentasi ini mencatat hasil pengujian fisik (*physical testing*) integrasi Go dengan thermal printer BP-LITE58 melalui Linux USB printer device (`/dev/usb/lp0`).

## Physical Print Evidence

### 1. Basic Text

Printer berhasil mencetak:

```text
HELLO WORLD
```

Pengujian membuktikan pipeline berikut berhasil:

```text
Go
↓
ESC/POS Core
↓
Printer Transport
↓
/dev/usb/lp0
↓
BP-LITE58
```

### 2. ESC/POS Formatting

Pengujian fisik berhasil untuk:

- Bold
- Left alignment
- Center alignment
- Right alignment
- Font size
- Paper feed

### 3. Barcode

Printer berhasil mencetak barcode CODE39.

**Data pengujian:** `KASA001`

### 4. Paper Cut

Automatic paper cut diuji menggunakan ESC/POS `GS V 0`.

**Hasil pengujian:**

- Automatic cutter: tidak tersedia
- Manual tear-off: tersedia

Printer melakukan paper feed, tetapi tidak melakukan pemotongan otomatis.

### 5. Character Width

Character width diuji menggunakan beberapa panjang baris.

| Panjang baris | Hasil |
|---:|---|
| 32 characters | 1 line |
| 33 characters | Wrapped |
| 42 characters | Wrapped |
| 48 characters | Wrapped |
| 49 characters | Wrapped |

**Kesimpulan:**

Normal character width = **32 characters**.

Nilai ini digunakan sebagai constraint utama *Receipt Layout Engine*.

### 6. Prototype Receipt

Prototype receipt berhasil dicetak secara fisik. Receipt mencakup:

- Store information
- Invoice
- Timestamp
- Cashier
- Product items
- Quantity
- Unit price
- Subtotal
- Total
- Payment
- Change
- Footer

### 7. Long Product Name

Produk dengan nama panjang diuji menggunakan receipt width 32 karakter.

Layout engine berhasil melakukan wrapping sehingga setiap baris tetap berada dalam batas printer.

### 8. Multiple Items

Receipt dengan banyak item diuji.

**Hasil pengujian:**

- All items rendered
- All lines remained within printer width
- Subtotal remained correct

### 9. Large Price

Harga besar diuji sampai nominal jutaan dan ratusan juta.

**Hasil pengujian:**

- Price formatting remained correct
- Right-aligned price remained within printer width
- Layout remained within printer width

### 10. Large Quantity

Quantity besar diuji sampai:

- `99`
- `999`
- `9999`
- `100000`

**Hasil pengujian:**

- Quantity remained visible
- Subtotal remained correctly calculated
- Layout remained within printer width

### 11. Printer Disconnect

Printer dicabut ketika aplikasi sedang aktif.

**Hasil pengujian:**

```text
Write error: no such device
```

Error berhasil diteruskan ke aplikasi tanpa menyebabkan crash.

### 12. Printer Reconnect

Setelah printer disambungkan kembali:

```text
Plaintext
↓
USB reconnect
↓
/dev/usb/lp0 available
↓
Printer reopened successfully
↓
Write succeeded
```

**Physical test:**

`KASA RECONNECT OK` berhasil dicetak.

## Physical Test Conclusion

BP-LITE58 berhasil digunakan dari Go melalui Linux USB printer device.

### Verified Capabilities

| Capability | Status |
|---|---|
| USB connection | PASS |
| Printer transport | PASS |
| ESC/POS initialization | PASS |
| Text printing | PASS |
| Bold | PASS |
| Alignment | PASS |
| Font size | PASS |
| Paper feed | PASS |
| Barcode CODE39 | PASS |
| Prototype receipt | PASS |
| Manual tear-off | PASS |
| Disconnect detection | PASS |
| Reconnect | PASS |

Printer character width untuk normal printing ditetapkan sebesar **32 characters**.

## Barcode Compatibility

| Barcode | Result | Command |
|---|---|---|
| CODE39 | PASS | `GS k 4` |
| CODE128 | PASS | `GS k 73` |
| EAN-13 | PASS | `GS k 2` |

EAN-13 menggunakan legacy ESC/POS barcode command `GS k 2`.
The printer successfully printed the barcode using 12 data digits,
with the check digit generated by the printer.