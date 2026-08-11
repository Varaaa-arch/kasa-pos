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
EOF