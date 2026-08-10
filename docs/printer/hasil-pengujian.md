# Hasil Pengujian Printer

Printer: BLUEPRINT BP-LITE58

---

## Test 01 — USB Detection

Status: ⬜ Belum diuji

### Tujuan

Memastikan Fedora dapat mendeteksi printer
ketika terhubung melalui USB.

### Command

lsusb

### Result

TBD

---

## Test 02 — Linux Driver

Status: ⬜ Belum diuji

### Tujuan

Mengetahui driver yang digunakan Fedora
untuk printer.

### Command

lsusb -t

### Result

TBD

---

## Test 03 — Device Node

Status: ⬜ Belum diuji

### Tujuan

Mengetahui device node yang digunakan Linux
untuk mengakses printer.

### Command

ls -la /dev/usb/ /dev/lp* /dev/tty* 2>/dev/null

### Result

TBD

---

## Test 04 — First Print

Status: ⬜ Belum diuji

### Tujuan

Mencetak teks sederhana melalui BP-LITE58.

### Expected

HELLO WORLD

### Actual

TBD

---

## Test 05 — ESC/POS

Status: ⬜ Belum diuji

### Tujuan

Memastikan printer menerima command ESC/POS.

### Result

TBD