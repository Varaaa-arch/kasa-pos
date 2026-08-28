# Spesifikasi Printer

## Device

| Informasi | Nilai |
|---|---|
| Merek | BLUEPRINT |
| Series | LITE |
| Model | BP-LITE58 |
| Jenis | Thermal Receipt Printer |
| Paper Width | 58mm |
| Interface | USB + Bluetooth |
| Protocol | ESC/POS |
| Device Node | `/dev/usb/lp0` (Linux) |

## USB Configuration

| Informasi | Nilai |
|---|---|
| Default Device | `/dev/usb/lp0` |
| Permissions | `666` (rw-rw-rw-) |
| User Group | `lp` |
| Detection | `lsusb` for device identification |

## Print Agent Configuration

| Informasi | Nilai |
|---|---|
| Default Port | `8081` |
| Default Address | `127.0.0.1:8081` |
| Protocol | HTTP |
| Content-Type | `application/json` |
| Idempotency | Header: `Idempotency-Key` |

## Operating System Support

| OS | Status | Notes |
|---|---|---|
| Linux | ✅ Supported | Primary development platform |
| Windows | ⚠️ Planned | Requires different USB driver |
| macOS | ⚠️ Planned | Requires different USB driver |

## Paper Specifications

| Informasi | Nilai |
|---|---|
| Type | Thermal paper |
| Width | 58mm (±0.5mm) |
| Diameter | 80mm (standard roll) |
| Core | 12.7mm (0.5 inch) |
| Length | ~50-80 meters per roll |

## Environmental Conditions

| Parameter | Range |
|---|---|
| Operating Temperature | 5°C - 45°C |
| Storage Temperature | -20°C - 60°C |
| Humidity | 20% - 85% RH (non-condensing) |

## Power Requirements

| Parameter | Value |
|---|---|
| Input Voltage | DC 5V - 9V |
| Current | 1.5A - 2A |
| Power Source | USB power or external adapter |

## Performance Specifications

| Parameter | Value |
|---|---|
| Print Speed | ~90mm/s |
| Resolution | 203 DPI (8 dots/mm) |
| Interface Speed | USB 2.0 Full Speed |
| Warm-up Time | < 3 seconds |

## Supported Barcode Types

| Barcode Type | Support |
|---|---|
| EAN-13 | ✅ Supported |
| Code 128 | ✅ Supported |
| Code 39 | ✅ Supported |
| QR Code | ⚠️ Planned |
| PDF417 | ❌ Not Supported |

## Font Support

| Font Type | Support |
|---|---|
| Standard ASCII | ✅ Supported |
| Double Width | ✅ Supported |
| Double Height | ✅ Supported |
| Bold | ✅ Supported |
| Underline | ✅ Supported |
| Custom Fonts | ❌ Not Supported |

## ESC/POS Commands

KASA Print Agent supports standard ESC/POS commands:

- Text formatting (bold, underline, double size)
- Alignment (left, center, right)
- Barcode printing
- Paper cutting (full/partial)
- Line feeds and form feeds
- Graphics and images (limited)

## Maintenance

### Recommended Maintenance Schedule

- **Daily**: Clean print head with alcohol wipe
- **Weekly**: Check paper roll and replace if low
- **Monthly**: Clean paper path and sensors
- **Quarterly**: Update firmware if available
- **Annually**: Full printer calibration

### Common Maintenance Tasks

**Paper Replacement**
1. Open printer cover
2. Remove empty roll core
3. Insert new paper roll with correct orientation
4. Feed paper through print mechanism
5. Close cover and test print

**Print Head Cleaning**
1. Turn off printer
2. Open printer cover
3. Gently clean print head with isopropyl alcohol
4. Allow to dry completely
5. Close cover and power on

**Sensor Cleaning**
1. Turn off printer
2. Locate paper sensors near print head
3. Clean with compressed air or soft brush
4. Power on and test

## Troubleshooting Guide

### Paper Jams

**Symptoms**: Printer stops mid-print, paper stuck

**Solutions**:
1. Turn off printer
2. Open printer cover
3. Gently remove jammed paper
4. Check for torn paper pieces
5. Reload paper properly
6. Test with single print

### Poor Print Quality

**Symptoms**: Faded print, missing lines, poor contrast

**Solutions**:
1. Check paper quality (use thermal paper)
2. Clean print head
3. Check paper alignment
4. Replace paper roll if old/damaged
5. Adjust print density settings (if available)

### Connectivity Issues

**Symptoms**: Printer not detected, communication errors

**Solutions**:
1. Check USB cable connection
2. Try different USB port
3. Verify device permissions
4. Restart print agent
5. Check system logs for USB errors

### Paper Out Errors

**Symptoms**: Paper out error despite paper being loaded

**Solutions**:
1. Verify paper is loaded correctly
2. Check paper sensors for obstructions
3. Clean paper sensors
4. Ensure paper is thermal type
5. Check paper end detector

## Known Limitations

1. **Graphics**: Limited graphics support, text-based receipts only
2. **Color**: Monochrome thermal printing only
3. **Paper Size**: Fixed 58mm width only
4. **Speed**: Not suitable for high-volume receipt printing
5. **OS Support**: Currently Linux-only for USB access

## Future Enhancements

Planned improvements for printer integration:

- [ ] Windows and macOS USB support
- [ ] QR code printing
- [ ] Advanced graphics support
- [ ] Wireless printing via Bluetooth
- [ ] Mobile app integration
- [ ] Cloud printing services
- [ ] Advanced error recovery
- [ ] Performance monitoring dashboard
