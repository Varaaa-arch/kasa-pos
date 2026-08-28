# Failure Injection Tests Documentation

This document describes the failure injection tests implemented for the POS system checkout process, including the expected behaviors for each failure scenario.

## Overview

Failure injection tests simulate various production failure scenarios to verify that the system handles errors gracefully and maintains data consistency. These tests ensure that the POS system can recover from failures without compromising data integrity.

## Test Scenarios

### 1. Printer Disconnected (`TestFailure_PrinterDisconnected`)

**Scenario**: Physical printer is disconnected during receipt printing.

**Expected Behavior**:
- ✅ Checkout transaction **succeeds** (transaction is created and committed)
- ✅ Stock is **reduced** normally
- ✅ Print job status is **FAILED**
- ✅ Error message clearly indicates printer disconnection
- ✅ Transaction remains **valid** despite print failure
- ✅ System can continue accepting new orders

**Rationale**: Print failures are non-critical for business operations. The transaction is already valid and committed, so print failure should not roll back the sale.

---

### 2. Print Agent Down (`TestFailure_PrintAgentDown`)

**Scenario**: Print agent service is unreachable or has crashed.

**Expected Behavior**:
- ✅ Atomic checkout **succeeds** (transaction created)
- ✅ Stock is **reduced** normally
- ✅ Print job status is **FAILED**
- ✅ Error message indicates print agent unreachability
- ✅ Transaction status is **COMPLETED**
- ✅ System remains operational for new orders

**Rationale**: Similar to printer disconnection, print agent failures should not block the core checkout process. The transaction is valid even if receipt printing fails.

---

### 3. Print Agent Network Timeout (`TestFailure_PrintAgentNetworkTimeout`)

**Scenario**: Network issues cause print agent requests to timeout.

**Expected Behavior**:
- ✅ System handles timeout gracefully
- ✅ Print job marked as **FAILED** with timeout error
- ✅ Transaction remains **valid**
- ✅ No panic or unhandled exceptions
- ✅ Appropriate timeout error message

**Rationale**: Network timeouts are transient failures. The system should handle them without crashing and allow retry mechanisms.

---

### 4. PostgreSQL Down (`TestFailure_PostgreSQLDown`)

**Scenario**: PostgreSQL database is unreachable or connection fails.

**Expected Behavior**:
- ✅ Checkout **fails immediately** with clear error
- ✅ **No transaction** is created
- ✅ **Stock remains unchanged**
- ✅ Error message indicates database connection issue
- ✅ System state remains **consistent**

**Rationale**: Database failures are critical. The system should fail fast and prevent any state changes when the database is unavailable.

---

### 5. Insufficient Stock (`TestFailure_InsufficientStock`)

**Scenario**: Product stock is insufficient to fulfill the order.

**Expected Behavior**:
- ✅ Checkout **fails** with `ErrInsufficientStock`
- ✅ **No transaction** is committed (transaction may be created in TX but rolled back)
- ✅ **Stock remains unchanged**
- ✅ Error message clearly indicates insufficient stock
- ✅ System prevents overselling

**Rationale**: Stock validation is critical for inventory management. The system must prevent selling items that aren't available.

---

### 6. Concurrent Checkout (`TestFailure_ConcurrentCheckout`)

**Scenario**: Multiple checkout requests attempt to purchase the same limited stock item simultaneously.

**Expected Behavior**:
- ✅ **Exactly one** checkout succeeds (exactly-once semantics)
- ✅ All other checkouts **fail** with `ErrInsufficientStock`
- ✅ Final stock is **consistent** (never negative)
- ✅ **No race conditions** or lost updates
- ✅ Database row locking prevents concurrent modifications

**Rationale**: Race conditions on stock reduction must be prevented. The system uses `SELECT FOR UPDATE` to ensure serializable access to product stock.

---

### 7. Print Agent Idempotency (`TestFailure_PrintAgentIdempotency`)

**Scenario**: Retry mechanism attempts to print the same receipt multiple times due to transient failures.

**Expected Behavior**:
- ✅ Retry with same idempotency key **does not duplicate** print
- ✅ Print agent executes print **exactly once** per idempotency key
- ✅ Returns success on retry without actual printing
- ✅ Prevents duplicate receipts

**Rationale**: Idempotency is crucial for handling transient failures. Retry mechanisms should not cause duplicate operations.

---

### 8. Cascading Failures (`TestFailure_CascadingFailures`)

**Scenario**: Multiple different failure scenarios occur in sequence.

**Expected Behavior**:
- ✅ Each failure is handled with **appropriate error handling**
- ✅ **No panic** or unhandled errors
- ✅ **System state remains consistent** after each failure
- ✅ Error messages are specific to each failure type

**Rationale**: The system should handle multiple failure types gracefully without cascading into system-wide failures.

---

## Test Infrastructure

### Mock Components

The tests use several mock components to simulate failures:

- **`disconnectedPrinter`**: Simulates a physically disconnected printer
- **`failurePrintAgent`**: Simulates print agent failures
- **`brokenDB`**: Simulates database connection failures
- **`serialStockRepo`**: Simulates database row locking with mutex for concurrent tests

### Test Utilities

- **`openSQLiteDB(t)`**: Creates in-memory SQLite database for transaction testing
- **`sampleReceipt()`**: Provides sample receipt data for print tests
- **`serialStockRepo`**: Mutex-based stock repository for simulating `SELECT FOR UPDATE`

---

## Running the Tests

```bash
# Run all failure injection tests
go test -v ./internal/service/checkout -run TestFailure

# Run specific failure test
go test -v ./internal/service/checkout -run TestFailure_PrinterDisconnected

# Run with race detection
go test -race -v ./internal/service/checkout -run TestFailure
```

---

## Key Principles

1. **Fail Fast**: Critical failures (database, stock) should fail immediately
2. **Non-Critical Operations**: Print failures should not block core business logic
3. **Data Consistency**: Stock and transaction data must remain consistent
4. **Error Clarity**: Error messages should clearly indicate the failure type
5. **Graceful Degradation**: System should remain operational despite non-critical failures
6. **Idempotency**: Retry mechanisms should not cause duplicate operations
7. **Concurrency Safety**: Race conditions must be prevented with proper locking

---

## Production Considerations

When deploying to production, consider:

1. **Monitoring**: Set up alerts for print failures to detect hardware issues
2. **Retry Logic**: Implement exponential backoff for transient failures
3. **Circuit Breakers**: Prevent cascading failures when services are down
4. **Health Checks**: Regular health checks for database and print agent
5. **Manual Recovery**: Processes for handling stuck print jobs
6. **Audit Logs**: Log all failures for troubleshooting and analysis

---

## Test Coverage

The failure injection tests cover the following failure domains:

- ✅ **Hardware Failures**: Printer disconnection
- ✅ **Service Failures**: Print agent down
- ✅ **Network Failures**: Timeouts, connection issues
- ✅ **Database Failures**: Connection failures, transaction failures
- ✅ **Business Logic Failures**: Insufficient stock
- ✅ **Concurrency Issues**: Race conditions, simultaneous access
- ✅ **Retry Scenarios**: Idempotency, duplicate prevention

These tests ensure the POS system is resilient to common production failures while maintaining data integrity and business continuity.