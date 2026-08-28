# Printer Documentation

KASA POS system integrates with BLUEPRINT BP-LITE58 thermal printer for receipt printing. This documentation covers printer setup, operation, testing, and troubleshooting.

## Quick Start

### Running Print Agent

```bash
# Development
go run ./experiments/printer/print-agent

# Production build
go build -o bin/print-agent ./experiments/printer/print-agent
./bin/print-agent
```

### Environment Variables

```bash
# Printer device path (default: /dev/usb/lp0)
export PRINTER_DEVICE=/dev/usb/lp0

# Server address (default: 127.0.0.1:8081)
export LISTEN_ADDRESS=127.0.0.1:8081
```

### Default Configuration

```go
const (
    printerDevice = "/dev/usb/lp0"
    listenAddress = "127.0.0.1:8081"
    shutdownTimeout = 10 * time.Second
)
```

## Printer Specifications

### Device Information

| Specification | Value |
|---------------|-------|
| Brand | BLUEPRINT |
| Series | LITE |
| Model | BP-LITE58 |
| Type | Thermal Receipt Printer |
| Paper Width | 58mm |
| Interface | USB + Bluetooth |
| Protocol | ESC/POS |

### USB Configuration

```bash
# Check USB devices
lsusb

# Find printer device
ls -la /dev/usb/

# Set permissions (if needed)
sudo chmod 666 /dev/usb/lp0

# Add user to lp group (permanent fix)
sudo usermod -a -G lp $USER
```

## Print Agent Endpoints

### POST /print

Print receipt with idempotency support.

**Request:**
```http
POST /print
Content-Type: application/json
Idempotency-Key: transaction-uuid

{
  "store": {
    "name": "TOKO KASA",
    "address": "Jl. Contoh No. 123",
    "phone": "081234567890"
  },
  "transaction": {
    "id": "transaction-uuid",
    "invoice_number": "INV-001",
    "timestamp": "2026-08-28T12:00:00Z",
    "cashier": "Kasir"
  },
  "items": [
    {
      "product_id": "prod-uuid",
      "sku": "PROD-001",
      "name": "Kopi Susu",
      "quantity": 2,
      "unit_price": 15000,
      "subtotal": 30000
    }
  ],
  "summary": {
    "subtotal": 30000,
    "discount": 0,
    "tax": 0,
    "service_charge": 0,
    "total": 30000
  },
  "payment": {
    "method": "CASH",
    "paid": 30000,
    "change": 0
  },
  "footer": {
    "message": "Terima kasih"
  }
}
```

**Response:** `200 OK`
```json
{
  "job_id": "PJ-uuid",
  "message": "receipt printed successfully"
}
```

**Error Response:** `500 Internal Server Error`
```json
{
  "error": "printer disconnected"
}
```

### GET /status

Check print agent and printer status.

**Response:** `200 OK`
```json
{
  "status": "ready",
  "printer": {
    "connected": true,
    "device": "/dev/usb/lp0",
    "paper": "ok"
  },
  "queue": {
    "pending": 0,
    "processing": 0
  }
}
```

### GET /health

Health check endpoint for monitoring.

**Response:** `200 OK`
```json
{
  "status": "healthy"
}
```

## ESC/POS Implementation

### Supported Commands

- **Initialize**: `ESC @` - Initialize printer
- **Text formatting**: Bold, underline, double height/width
- **Alignment**: Left, center, right
- **Barcode printing**: EAN-13, Code 128
- **Cut paper**: Full cut, partial cut
- **Feed control**: Line feeds, form feeds

### Receipt Layout

```
┌─────────────────────────────┐
│       TOKO KASA             │
│   Jl. Contoh No. 123        │
│   081234567890              │
├─────────────────────────────┤
│ INV-001                     │
│ 2026-08-28 12:00:00        │
│ Kasir: Kasir                │
├─────────────────────────────┤
│ Item           Qty   Price   │
│─────────────────────────────│
│ Kopi Susu      2    15.000  │
│               Sub   30.000   │
├─────────────────────────────┤
│ Subtotal               30.000│
│ Discount                  -0│
│ Tax                       +0│
│ Service Charge            +0│
│─────────────────────────────│
│ TOTAL                  30.000│
│ CASH                   30.000│
│ CHANGE                   -0│
├─────────────────────────────┤
│ Payment: CASH              │
├─────────────────────────────┤
│      Terima kasih          │
│   2026-08-28 12:00:00      │
└─────────────────────────────┘
```

## Testing

### Unit Tests

```bash
# Run printer unit tests
go test ./internal/printer/...

# Run specific printer tests
go test ./internal/printer/escpos
go test ./internal/printer/receipt
go test ./internal/printer/transport
```

### Integration Tests

```bash
# Run print agent integration tests
go test ./experiments/printer/print-agent

# Test with real printer (requires hardware)
PRINTER_DEVICE=/dev/usb/lp0 go test ./internal/printer/transport
```

### Manual Testing

```bash
# Test print agent health
curl http://localhost:8081/health

# Test print agent status
curl http://localhost:8081/status

# Test print with sample receipt
curl -X POST http://localhost:8081/print \
  -H "Content-Type: application/json" \
  -H "Idempotency-Key: test-001" \
  -d @sample-receipt.json
```

## Failure Simulation

### Simulating Printer Disconnection

```bash
# Unplug printer USB cable
# Send print request
curl -X POST http://localhost:8081/print \
  -H "Content-Type: application/json" \
  -d @sample-receipt.json

# Expected: Error response
# System behavior: Transaction succeeds, print job fails
```

### Simulating Print Agent Failure

```bash
# Stop print agent
pkill -f print-agent

# Send print request from API
curl -X POST http://localhost:8080/checkout \
  -H "Content-Type: application/json" \
  -d '{
    "items": [{"product_id": "uuid", "quantity": 1}],
    "paid_amount": 10000,
    "payment_method": "CASH"
  }'

# Expected: Transaction succeeds, print job status = FAILED
```

### Simulating Paper Out

```bash
# Remove paper from printer
# Send print request
curl -X POST http://localhost:8081/print \
  -H "Content-Type: application/json" \
  -d @sample-receipt.json

# Expected: Error response indicating paper out
```

## Recovery Procedures

### Print Agent Recovery

**1. Check Print Agent Status**
```bash
curl http://localhost:8081/health
curl http://localhost:8081/status
```

**2. Restart Print Agent**
```bash
# Stop current process
pkill -f print-agent

# Start again
go run ./experiments/printer/print-agent
```

**3. Verify Printer Connection**
```bash
# Check USB device
ls -la /dev/usb/lp0

# Test printer communication
echo "Test" > /dev/usb/lp0
```

**4. Reprint Failed Receipts**
```bash
# Reprint via API
curl -X POST http://localhost:8080/transactions/:id/reprint
```

### Printer Hardware Recovery

**1. Check Physical Connections**
- Verify USB cable is connected
- Check printer power is on
- Ensure paper is loaded

**2. Reset Printer**
```bash
# Power cycle the printer
# Unplug USB cable, wait 10 seconds, reconnect
```

**3. Check System Permissions**
```bash
# Verify device permissions
ls -la /dev/usb/lp0

# Fix permissions if needed
sudo chmod 666 /dev/usb/lp0
```

**4. Test Direct Communication**
```bash
# Send direct ESC/POS command
echo -e "\x1B\x40" > /dev/usb/lp0  # Initialize
echo "Test Print" > /dev/usb/lp0
echo -e "\x0D" > /dev/usb/lp0    # Print and feed
```

## Troubleshooting

### Common Issues

**Printer not detected**
```bash
# Check USB devices
lsusb | grep Blueprint

# Check device permissions
ls -la /dev/usb/

# Check kernel messages
dmesg | grep usb
```

**Permission denied**
```bash
# Add user to lp group
sudo usermod -a -G lp $USER

# Or use chmod (temporary)
sudo chmod 666 /dev/usb/lp0
```

**Print agent not starting**
```bash
# Check if port is already in use
lsof -i :8081

# Check logs for errors
# Run with verbose logging
```

**Print job hanging**
```bash
# Check print agent status
curl http://localhost:8081/status

# Restart print agent
pkill -f print-agent
go run ./experiments/printer/print-agent
```

**Receipt printing garbage characters**
```bash
# Check encoding settings
# Verify ESC/POS command sequence
# Test with simple text first
```

## Monitoring

### Health Checks

```bash
# Continuous health monitoring
watch -n 5 'curl -s http://localhost:8081/health'
```

### Log Monitoring

```bash
# Monitor print agent logs
journalctl -u print-agent -f

# Or if running manually
# Monitor stdout/stderr of the process
```

### Metrics (Future)

Future releases will include:
- Print job success/failure rates
- Average print duration
- Queue depth monitoring
- Error rate tracking

## Performance Optimization

### Batch Printing

For high-volume scenarios, implement batch printing:

```go
// Queue multiple print jobs
jobs := []PrintJob{job1, job2, job3}

// Process sequentially
for _, job := range jobs {
    printAgent.Print(ctx, job.Receipt, job.ID)
}
```

### Connection Pooling

Print agent uses HTTP keep-alive for efficient connection reuse.

### Asynchronous Processing

Consider implementing async print job processing for better performance:

```go
// Queue print job
go func() {
    printAgent.Print(ctx, receipt, idempotencyKey)
}()
```

## Security Considerations

### Network Security

- Run print agent on localhost in production
- Use firewall rules to restrict access
- Implement API authentication (future)

### Device Security

- Restrict USB device access
- Monitor for unauthorized device connections
- Secure physical access to printer

## Best Practices

1. **Always use idempotency keys** for print requests
2. **Implement retry logic** for transient failures
3. **Monitor print agent health** in production
4. **Test print jobs** after printer maintenance
5. **Keep firmware updated** for reliability
6. **Use quality paper** to prevent jams
7. **Regular cleaning** of print head
8. **Spare printer** for business continuity

## References

- [ESC/POS Command Reference](escpos-research.md)
- [Receipt Data Model](receipt-data-model.md)
- [Receipt Layout Specification](receipt-layout.md)
- [USB Research](usb-research.md)
- [Test Results](hasil-pengujian.md)