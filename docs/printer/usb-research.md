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

Bus 001 Device 001: ID 1d6b:0002 Linux Foundation 2.0 root hub
Bus 002 Device 001: ID 1d6b:0003 Linux Foundation 3.0 root hub
Bus 003 Device 001: ID 1d6b:0002 Linux Foundation 2.0 root hub
Bus 003 Device 002: ID 3151:3020 YICHIP Wireless Device
Bus 003 Device 003: ID 13d3:5439 IMC Networks Integrated Camera
Bus 003 Device 004: ID 8087:0032 Intel Corp. AX210 Bluetooth
Bus 003 Device 005: ID 28e9:0289 GDMicroelectronics micro-printer **Ini yang berbeda**
Bus 004 Device 001: ID 1d6b:0003 Linux Foundation 3.0 root hub

## BP-LITE58 USB Identification

### USB Device

| Property | Value |
|---|---|
| Manufacturer | GEZHI |
| Product | micro-printer |
| VID | `0x28e9` |
| PID | `0x0289` |
| VID:PID | `28e9:0289` |
| Serial Number | `000000000004` |
| USB Version | 1.10 |
| Negotiated Speed | Full Speed (12 Mbps) |
| Bus | 003 |
| Device | 005 |

### USB Interface

| Property | Value |
|---|---|
| Interface Number | `0` |
| Interface Class | Printer (`7`) |
| Interface Subclass | Printer (`1`) |
| Interface Protocol | Bidirectional (`2`) |
| Number of Endpoints | `2` |

### Endpoints

| Endpoint | Direction | Transfer Type | Max Packet Size |
|---|---|---|---:|
| `0x01` | OUT | Bulk | 64 bytes |
| `0x81` | IN | Bulk | 64 bytes |

### Initial Conclusion

The BP-LITE58 is detected by Linux as a USB Printer Class device.

The printer exposes a bidirectional USB printer interface with:

- Bulk OUT endpoint `0x01` for host-to-printer communication.
- Bulk IN endpoint `0x81` for printer-to-host communication.

This indicates that the printer can potentially be accessed through the
standard USB Printer Class stack or directly through USB endpoints.

Further testing is required to determine the exact transport mechanism and
whether the printer can accept raw ESC/POS commands through the USB
interface.


## Communication Method

### Selected Method

For the initial KASA POS implementation, the printer will be accessed
through the Linux USB printer device:

`/dev/usb/lp0`

The communication path is:

Go application
    ↓
/dev/usb/lp0
    ↓
Linux `usblp`
    ↓
USB Printer Class
    ↓
BP-LITE58

### Reason

The BP-LITE58 is detected by Fedora as a USB Printer Class device and the
Linux kernel exposes it through `/dev/usb/lp0`.

This allows the application to send raw ESC/POS bytes through the printer
device without implementing the USB transport layer directly.

### USB Details

- VID: `0x28e9`
- PID: `0x0289`
- Interface Class: Printer (`7`)
- Interface Protocol: Bidirectional (`2`)
- OUT Endpoint: `0x01`
- IN Endpoint: `0x81`

### Alternative

Direct USB communication using `libusb` is considered an alternative for
future versions if direct endpoint control becomes necessary.

For V1, direct USB access is intentionally avoided because the Linux
`usblp` abstraction already provides the required printer transport.