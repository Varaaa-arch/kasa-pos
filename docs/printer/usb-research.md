# USB Research — BLUEPRINT BP-LITE58

## Objective

Mengidentifikasi bagaimana Fedora Linux mengenali
dan berkomunikasi dengan printer BP-LITE58 melalui USB.

## Environment

- OS: Fedora Linux
- Printer: BLUEPRINT BP-LITE58
- Connection: USB

## USB Identification

### Before Printer Connected

Command:

lsusb

Result:

Bus 001 Device 001: ID 1d6b:0002 Linux Foundation 2.0 root hub
Bus 002 Device 001: ID 1d6b:0003 Linux Foundation 3.0 root hub
Bus 003 Device 001: ID 1d6b:0002 Linux Foundation 2.0 root hub
Bus 003 Device 002: ID 3151:3020 YICHIP Wireless Device
Bus 003 Device 003: ID 13d3:5439 IMC Networks Integrated Camera
Bus 003 Device 004: ID 8087:0032 Intel Corp. AX210 Bluetooth
Bus 004 Device 001: ID 1d6b:0003 Linux Foundation 3.0 root hub

### After Printer Connected

Command:

lsusb

Result:

TBD

## USB Identity

| Informasi | Nilai |
|---|---|
| Vendor ID | TBD |
| Product ID | TBD |
| Device Name | TBD |

## Linux Detection

### Kernel

Command:

dmesg | tail -50

Result:

TBD

### USB Tree

Command:

lsusb -t

Result:

TBD

### Device Nodes

Command:

ls -la /dev/usb/ /dev/lp* /dev/tty* 2>/dev/null

Result:

TBD

## Communication Method

TBD

## Notes

TBD