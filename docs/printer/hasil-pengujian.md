# Hasil Pengujian Printer

## 1. Informasi Eksperimen

- Project: POS System
- Komponen: Receipt Prototype
- Bahasa: Go
- Printer target: Blueprint BP-LITE58
- Printer interface: USB
- Printer protocol: ESC/POS
- Receipt width: 48 karakter
- Status koneksi printer: Belum diuji
- Status kabel USB-B: Belum tersedia

---

## 2. Tujuan Pengujian

Eksperimen ini bertujuan untuk memastikan bahwa sistem POS dapat menghasilkan
format struk yang terstruktur sebelum dihubungkan secara langsung dengan
printer thermal BP-LITE58.

Fokus pengujian pada tahap ini adalah:

1. Membuat model data transaksi.
2. Menghasilkan format struk menggunakan Go.
3. Mengatur alignment teks.
4. Mengatur format harga Rupiah.
5. Mengatur layout item dan subtotal.
6. Menghasilkan total transaksi.
7. Menghasilkan informasi pembayaran dan kembalian.
8. Menghasilkan footer struk.

---

## 3. Data Transaksi Pengujian

### Store

- Name: TOKO KASA
- Address: Jl. Contoh No. 123
- Phone: 081234567890

### Transaction

- Invoice: INV-000001
- Timestamp: 10/08/2026 19:00:00
- Cashier: Bizar

### Items

| Item | Quantity | Unit Price | Subtotal |
|---|---:|---:|---:|
| Kopi Susu | 2 | Rp 15.000 | Rp 30.000 |
| Roti Bakar | 1 | Rp 12.000 | Rp 12.000 |
| Air Mineral | 1 | Rp 5.000 | Rp 5.000 |

### Summary

- Subtotal: Rp 47.000
- Discount: Rp 0
- Tax: Rp 0
- Service Charge: Rp 0
- Total: Rp 47.000

### Payment

- Method: CASH
- Paid: Rp 50.000
- Change: Rp 3.000

---

## 4. Hasil Pengujian

Prototype berhasil menghasilkan output struk melalui terminal.

Output berhasil menampilkan:

- Nama toko
- Alamat toko
- Nomor telepon
- Nomor invoice
- Waktu transaksi
- Nama kasir
- Daftar barang
- Quantity
- Harga satuan
- Subtotal item
- Subtotal transaksi
- Total transaksi
- Jumlah pembayaran
- Kembalian
- Footer

Contoh output:

```text
                 TOKO KASA
             Jl. Contoh No. 123
               081234567890
================================================

Invoice: INV-000001
10/08/2026 19:00:00
Kasir: Bizar

Kopi Susu
2 x Rp 15.000                         Rp 30.000
Roti Bakar
1 x Rp 12.000                         Rp 12.000
Air Mineral
1 x Rp 5.000                           Rp 5.000
------------------------------------------------
Subtotal                              Rp 47.000
TOTAL                                 Rp 47.000
------------------------------------------------
Bayar                                 Rp 50.000
Kembali                                Rp 3.000
------------------------------------------------
                 TERIMA KASIH
```

## Paper Cut Test

### Method

Mengirim ESC/POS `GS V 0` untuk menguji automatic paper cut.

### Result

Printer melakukan paper feed, tetapi tidak melakukan
pemotongan otomatis.

### Conclusion

BP-LITE58 menggunakan mekanisme **manual tear-off** dan
tidak memiliki automatic paper cutter.

Aplikasi KASA tidak akan bergantung pada fitur automatic
paper cut. Receipt akan melakukan paper feed secukupnya
sehingga pengguna dapat merobek kertas secara manual.


## Barcode Test

### Method

Printer diuji menggunakan ESC/POS barcode command
`GS k` dengan barcode CODE39.

Data yang digunakan:

`KASA001`

### Result

Barcode berhasil dicetak oleh printer.

Barcode dapat digunakan sebagai output barcode
untuk kebutuhan POS.

### Conclusion

BP-LITE58 mendukung pencetakan barcode melalui
ESC/POS.

CODE39 berhasil diuji dan dapat digunakan sebagai
baseline untuk pengembangan barcode printer KASA.