# ADR-007: Failure Injection Testing Strategy

**Status:** Accepted
**Date:** 2026-08-28

## Context

KASA POS system handles critical business operations including inventory management, financial transactions, and physical receipt printing. System failures in production can have significant business impact:

- Lost sales due to system unavailability
- Inventory inconsistencies from failed transactions
- Financial discrepancies from duplicate transactions
- Customer experience issues from failed prints

Traditional testing focuses on happy paths and basic error handling, but doesn't adequately verify system behavior under realistic failure conditions. We need a comprehensive failure injection strategy to ensure system resilience.

## Decision

Implement a comprehensive failure injection testing framework that simulates realistic production failures and verifies system behavior under stress conditions.

### Failure Scenarios

We implement tests for the following failure scenarios:

1. **Hardware Failures**
   - Printer disconnection during printing
   - USB communication failures
   - Printer power loss

2. **Service Failures**
   - Print agent service unavailability
   - Print agent crashes
   - Network connectivity issues

3. **Database Failures**
   - PostgreSQL connection failures
   - Database query timeouts
   - Transaction deadlocks

4. **Business Logic Failures**
   - Insufficient stock scenarios
   - Invalid payment amounts
   - Concurrent access conflicts

5. **Network Failures**
   - Network timeouts
   - Connection refused
   - DNS resolution failures

### Testing Infrastructure

```go
// Mock components for failure simulation
type disconnectedPrinter struct{}
type failurePrintAgent struct{}
type brokenDB struct{}
type serialStockRepo struct{}
```

### Expected Behaviors

**Critical Failures (Database, Stock):**
- ✅ Fail fast with clear error messages
- ✅ No state changes occur
- ✅ System remains consistent
- ✅ Appropriate error logging

**Non-Critical Failures (Print):**
- ✅ Core business logic succeeds
- ✅ Transaction completes normally
- ✅ Print job marked as failed
- ✅ System remains operational
- ✅ Recovery mechanisms available

### Test Implementation

```go
func TestFailure_PrinterDisconnected(t *testing.T) {
    // Verify transaction succeeds despite printer failure
    // Verify print job status is FAILED
    // Verify system remains operational
}

func TestFailure_ConcurrentCheckout(t *testing.T) {
    // Verify exactly-once semantics
    // Verify no race conditions
    // Verify stock consistency
}
```

## Consequences

### Positive

- **Improved Resilience**: System behavior under failure conditions is well-understood
- **Better Error Handling**: Clear error messages and graceful degradation
- **Confidence**: Comprehensive failure coverage increases deployment confidence
- **Documentation**: Expected behaviors are clearly documented
- **Monitoring**: Tests inform monitoring and alerting strategies

### Negative

- **Test Complexity**: Failure injection tests add complexity to the test suite
- **Maintenance**: Mock components require maintenance as system evolves
- **Execution Time**: Comprehensive failure testing increases test execution time
- **False Confidence**: Tests simulate but cannot perfectly replicate production failures

### Trade-offs

- We prioritize testing critical failures (database, stock) over non-critical ones (print)
- We use realistic mocks rather than actual hardware failure simulation
- We focus on behavior verification rather than performance under failure
- We accept that some edge cases may not be covered by automated tests

## Related Decisions

- [ADR-003](ADR-003-print-job-lifecycle.md) - Print Job as State Machine (informs print failure handling)
- [ADR-006](ADR-006-api-error-model.md) - API Error Model (informs error response structure)

## Implementation Status

- ✅ Failure injection test framework implemented
- ✅ Core failure scenarios covered
- ✅ Expected behaviors documented
- ✅ Integration with CI/CD pipeline recommended
- ⚠️ Production monitoring and alerting based on test scenarios (future work)