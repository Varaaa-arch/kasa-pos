# ESC/POS Research

## Overview

ESC/POS adalah command system yang digunakan
untuk mengontrol thermal receipt printer.

Pada project POS ini, ESC/POS digunakan sebagai
format command untuk berkomunikasi dengan
printer BLUEPRINT BP-LITE58.

## Printing Flow

Data Transaksi
    ↓
Receipt Data
    ↓
Receipt Renderer
    ↓
ESC/POS Commands
    ↓
Print Transport
    ↓
USB
    ↓
BP-LITE58

## Required Features

| Feature | Priority | Status |
|---|---|---|
| Text | MUST | Research |
| Line Feed | MUST | Research |
| Alignment | MUST | Research |
| Bold | MUST | Research |
| Font Size | MUST | Research |
| Initialize | MUST | Research |
| Paper Feed | MUST | Research |
| Cut | SHOULD | Research |
| Barcode | SHOULD | Research |
| QR Code | SHOULD | Research |
| Image | SHOULD | Research |
| Cash Drawer | COULD | Research |

## Printer

Model:
BLUEPRINT BP-LITE58

Protocol:
ESC/POS

## Notes

Command compatibility akan diuji secara langsung
setelah printer dapat terhubung melalui USB.

## V1 Command Set

### MUST

- Initialize
- Text
- Line Feed
- Alignment
- Bold
- Text Size
- Paper Feed

### SHOULD

- Paper Cut
- QR Code
- Barcode
- Image / Logo

### COULD

- Cash Drawer

## Planned Printer Abstraction

ReceiptPrinter
    |
    ├── Initialize()
    ├── Text()
    ├── NewLine()
    ├── SetAlignment()
    ├── SetBold()
    ├── SetTextSize()
    ├── Feed()
    ├── Cut()
    ├── QRCode()
    ├── Barcode()
    └── Image()

## Receipt V1 Structure

1. Store Header
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
12. Paper Feed
13. Paper Cut