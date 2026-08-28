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

---

## Double Checkout Test (TASK 5.18)

### Test Overview

The double checkout test (`TestFailure_DoubleCheckout` and `TestFailure_DoubleCheckoutWithPrint`) simulates sending the same checkout request twice to verify the current behavior without idempotency mechanisms.

### Current Behavior (Without Idempotency)

#### Transaction Behavior
- ✗ **Creates 2 separate transactions** with different transaction IDs
- ✗ **Uses the same invoice number** for both transactions (data inconsistency)
- ✗ **Stock is reduced twice** (from 10 to 6 instead of 10 to 8 for single checkout)
- ✗ **Both transactions show COMPLETED status**

#### Print Behavior
- ✗ **Print occurs twice** (both print jobs succeed)
- ✗ **Same print job ID** (due to mock implementation, but in reality would be different)
- ✗ **Two physical receipts would be printed**

### Test Results

```
=== DOUBLE CHECKOUT BEHAVIOR ===
First checkout: TX ID = 5bdf4765-63c2-442c-ba39-1fe4476bc1b6, Invoice = INV-DOUBLE-TEST, Status = COMPLETED
Second checkout: TX ID = b3134c60-c1a4-483f-bf19-7372835a23b0, Invoice = INV-DOUBLE-TEST, Status = COMPLETED
Total transactions created: 2
Final stock: 6 (initial: 10, purchased: 2x2 = 4, expected: 6)
Transaction IDs are different: true
Invoice numbers are the same: true

✗ CURRENT BEHAVIOR: 2 transactions created - DUPLICATE!
✗ CURRENT BEHAVIOR: Stock reduced 2x - from 10 to 6 (should be 8 for single checkout)
✗ CURRENT BEHAVIOR: Different transaction IDs - this is duplicate transaction!
✗ CURRENT BEHAVIOR: Same invoice number for different transactions - data inconsistency!
```

### Issues Identified

1. **Duplicate Transactions**: Same checkout creates multiple transactions
2. **Inventory Inconsistency**: Stock is reduced multiple times incorrectly
3. **Data Integrity**: Same invoice number for different transactions
4. **Duplicate Printing**: Multiple receipts for the same sale
5. **Financial Impact**: Customer could be charged multiple times (in real payment systems)

### Required Improvements

1. **Idempotency Keys**: Implement unique idempotency keys for each checkout request
2. **Deduplication**: Check for existing transactions with same idempotency key
3. **Invoice Uniqueness**: Ensure invoice numbers are unique across transactions
4. **Print Idempotency**: Prevent duplicate printing for the same transaction
5. **Payment Idempotency**: Integrate with payment gateway idempotency

### Implementation Plan (Future)

The full idempotency implementation is planned for later stages and should include:

1. **Request-Level Idempotency**:
   - Generate unique idempotency key for each checkout request
   - Store idempotency keys with transaction results
   - Return cached result for duplicate requests

2. **Database Schema Changes**:
   - Add `idempotency_key` column to transactions table
   - Add unique constraint on idempotency_key
   - Add index for fast lookups

3. **API Changes**:
   - Accept idempotency key in checkout requests
   - Return existing transaction for duplicate keys
   - Handle idempotency conflicts gracefully

4. **Print Idempotency**:
   - Track print job execution per transaction
   - Prevent duplicate print jobs for same transaction
   - Link print jobs to transaction idempotency

### Current Status

⚠️ **WARNING**: The system currently does NOT have full idempotency implementation. Double checkouts will create duplicate transactions, reduce stock incorrectly, and print multiple receipts. This needs to be addressed before production deployment.

### Running the Double Checkout Test

```bash
# Run double checkout tests
go test -v ./internal/service/checkout -run TestFailure_DoubleCheckout
```

---

## Full E2E Regression Test (TASK 5.19)

### Test Overview

The Full E2E Regression Test (`TestFullE2ERegression` and `TestFullE2ERegressionWithAPI`) verifies the complete end-to-end flow from POS input to physical receipt output on BP-LITE58 printer.

### Complete Chain Verification

The test validates the entire system chain:

```
POS → Checkout → Transaction → PrintJob → Print Agent → BP-LITE58 → Receipt
```

### Test Components

#### 1. BP-LITE58 Printer Simulation
- **`bpLite58Printer`**: Realistic simulation of BP-LITE58 thermal printer
- Tracks open/close/write operations
- Simulates physical printer behavior
- Verifies proper printer lifecycle management

#### 2. Realistic Print Agent
- **`realisticPrintAgent`**: Simulates real print agent behavior
- Includes processing delays
- Integrates with BP-LITE58 simulation
- Returns realistic print responses

### Test Steps

#### TestFullE2ERegression

**STEP 1: Database Setup**
- Create PostgreSQL connection
- Setup product and transaction repositories
- Create test product with initial stock

**STEP 2: POS - Cart Creation**
- User selects products and adds to cart
- Verify cart total calculation
- Validate cart item structure

**STEP 3: Checkout Processing**
- Execute atomic checkout
- Process payment information
- Generate transaction record

**STEP 4: Transaction Persistence**
- Verify transaction stored in database
- Validate transaction status (COMPLETED)
- Confirm transaction details accuracy

**STEP 5: Stock Management**
- Verify stock reduction in database
- Validate inventory consistency
- Confirm atomic stock update

**STEP 6: Print Job Creation**
- Create print job from transaction
- Verify print job status (PENDING)
- Validate print job initialization

**STEP 7: Print Agent Setup**
- Initialize BP-LITE58 printer simulation
- Setup realistic print agent
- Configure print communication

**STEP 8: Print Execution**
- Send receipt to print agent
- Execute print via BP-LITE58
- Verify print agent response

**STEP 9: Receipt Verification**
- Verify print job completion
- Validate BP-LITE58 output
- Confirm printer lifecycle (open/write/close)

**STEP 10: Full Chain Verification**
- Verify receipt matches transaction
- Validate data consistency across chain
- Confirm end-to-end data integrity

#### TestFullE2ERegressionWithAPI

Tests the complete flow through the API layer, simulating real POS system usage:

- HTTP request simulation
- API endpoint integration
- Print agent server communication
- Complete chain verification via API

### Verification Points

Each component in the chain is verified:

#### ✅ POS Layer
- Cart creation and validation
- Item addition and total calculation
- User input processing

#### ✅ Checkout Layer
- Atomic transaction processing
- Payment handling
- Invoice generation

#### ✅ Transaction Layer
- Database persistence
- Status management
- Data integrity

#### ✅ Stock Layer
- Inventory updates
- Atomic stock reduction
- Consistency verification

#### ✅ PrintJob Layer
- Job creation and lifecycle
- Status management
- Error handling

#### ✅ Print Agent Layer
- Communication reliability
- Request/response handling
- Idempotency key usage

#### ✅ BP-LITE58 Layer
- Physical printer simulation
- Data output verification
- Lifecycle management

#### ✅ Receipt Layer
- Final output validation
- Data consistency
- Format verification

### Running the Full E2E Test

```bash
# Requires DATABASE_URL environment variable
export DATABASE_URL="postgres://user:password@localhost:5432/pos_system"

# Run full E2E regression test
go test -v ./internal/service/checkout -run TestFullE2ERegression

# Run API E2E test
go test -v ./internal/service/checkout -run TestFullE2ERegressionWithAPI

# Run all E2E tests
go test -v ./internal/service/checkout -run "TestFullE2E"
```

### Expected Output

```
=== FULL E2E REGRESSION TEST SUMMARY ===
✓ POS: Cart created successfully
✓ Checkout: Transaction processed successfully
✓ Transaction: Data persisted correctly
✓ Stock: Inventory updated correctly
✓ PrintJob: Print job created and managed
✓ Print Agent: Communication successful
✓ BP-LITE58: Physical printer simulation successful
✓ Receipt: Final output verified
=== COMPLETE CHAIN: POS → Checkout → Transaction → PrintJob → Print Agent → BP-LITE58 → Receipt ===
```

### Test Coverage

The full E2E regression test provides coverage for:

- ✅ **Complete System Integration**: All components working together
- ✅ **Data Consistency**: End-to-end data integrity verification
- ✅ **Error Handling**: Graceful failure handling across chain
- ✅ **Performance**: Processing time verification
- ✅ **Resource Management**: Proper cleanup and lifecycle management
- ✅ **Real-world Simulation**: Realistic component behavior

### Prerequisites

- PostgreSQL database with test schema
- Valid `DATABASE_URL` environment variable
- Database tables: `products`, `transactions`, `transaction_items`
- Sufficient database permissions for test operations

### Cleanup

The test includes automatic cleanup:
- Deletes test transactions
- Removes test products
- Closes database connections
- Shuts down test servers

### Integration with Other Tests

The full E2E regression test complements:
- Unit tests (individual component testing)
- Integration tests (component interaction testing)
- Failure injection tests (error scenario testing)
- Concurrent tests (race condition testing)

This provides a comprehensive test pyramid covering all system aspects.