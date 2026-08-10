# Receipt Data Model

## Purpose

Receipt Data Model mendefinisikan data yang dibutuhkan
oleh Receipt Renderer untuk menghasilkan receipt.

Receipt Data Model tidak bergantung pada printer,
USB, atau ESC/POS.

## Architecture

Transaction
    ↓
Receipt Data
    ↓
Receipt Renderer
    ↓
ESC/POS
    ↓
Print Transport
    ↓
Printer

## Receipt

### Store

- Name
- Address
- Phone
- Logo

### Transaction

- Invoice Number
- Date
- Cashier
- Order Type

### Items

Setiap item memiliki:

- Name
- SKU
- Quantity
- Unit Price
- Discount
- Total

### Summary

- Subtotal
- Discount
- Tax
- Service Charge
- Total

### Payment

- Method
- Amount
- Change

### Footer

- Message
- QR Code

## Design Principle

Receipt Data harus bersifat printer-independent.

Receipt Renderer bertanggung jawab mengubah
Receipt Data menjadi format output tertentu.

Contoh output:

- ESC/POS
- PDF
- HTML
- Digital Receipt

Printer tidak boleh mengetahui business logic transaksi.