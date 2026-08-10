# Receipt Layout V1

## Target Printer

- Model: BLUEPRINT BP-LITE58
- Paper: ~58 mm
- Character Width: TBD
- Characters Per Line: TBD

## Design Principle

Receipt harus:

- mudah dibaca
- tidak overflow
- harga rata kanan
- mendukung nama produk panjang
- printer-independent

## Receipt Structure

1. Header
2. Store Information
3. Transaction Information
4. Items
5. Subtotal
6. Discount
7. Total
8. Payment
9. Change
10. Footer
11. QR Code
12. Feed
13. Cut

## Item Layout

Product Name
Quantity x Unit Price             Total

Example:

Kopi Susu
2 x Rp15.000                  Rp30.000

## Summary Layout

Subtotal                      Rp40.000
Discount                           Rp0
--------------------------------------
TOTAL                         Rp40.000

## Payment Layout

Payment                       Rp50.000
Change                        Rp10.000

## Header

Header menggunakan:

- Center alignment
- Bold
- Optional larger text

## Footer

Footer menggunakan:

- Center alignment
- Thank-you message
- Optional QR Code

## Printer Width

Characters per line masih TBD.

Nilai final akan ditentukan setelah
hardware testing dengan BP-LITE58.