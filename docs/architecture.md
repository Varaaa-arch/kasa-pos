# System Architecture

## Planned Architecture

POS Web
    |
    | HTTP
    v
Go API
    |
    v
PostgreSQL


POS Web
    |
    | localhost
    v
Go Print Agent
    |
    | USB
    v
BLUEPRINT BP-LITE58

## Components

### POS Web

Frontend yang digunakan kasir untuk melakukan transaksi.

### Go API

Backend yang menangani business logic dan transaksi.

### PostgreSQL

Database utama aplikasi POS.

### Go Print Agent

Local service yang menangani komunikasi
antara POS dengan thermal printer.

### BP-LITE58

Thermal receipt printer yang digunakan
untuk mencetak struk transaksi.