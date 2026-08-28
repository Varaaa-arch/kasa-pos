# System Overview

## Architecture

KASA POS system follows a layered architecture with clear separation of concerns:

```
┌─────────────────────────────────────────────────────────────┐
│                    Presentation Layer                         │
│  ┌─────────────┐     ┌─────────────┐     ┌─────────────┐  │
│  │ POS Web     │────▶│ POS API     │────▶│ Print Agent │  │
│  │ (Next.js)   │     │ (Go HTTP)   │     │ (Go HTTP)   │  │
│  │ :3000       │     │ :8080       │     │ :8081       │  │
│  └─────────────┘     └─────────────┘     └─────────────┘  │
└─────────────────────────────────────────────────────────────┘
                           │                    │
                           ▼                    ▼
┌─────────────────────────────────────────────────────────────┐
│                    Business Logic Layer                       │
│  ┌─────────────┐     ┌─────────────┐     ┌─────────────┐  │
│  │ Checkout    │     │ Transaction │     │ Receipt     │  │
│  │ Service     │     │ Service     │     │ Service     │  │
│  └─────────────┘     └─────────────┘     └─────────────┘  │
│  ┌─────────────┐     ┌─────────────┐                        │
│  │ Product     │     │ Print       │                        │
│  │ Service     │     │ Service     │                        │
│  └─────────────┘     └─────────────┘                        │
└─────────────────────────────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│                    Data Access Layer                          │
│  ┌─────────────┐     ┌─────────────┐                        │
│  │ Product     │     │ Transaction │                        │
│  │ Repository  │     │ Repository  │                        │
│  └─────────────┘     └─────────────┘                        │
└─────────────────────────────────────────────────────────────┘
                           │
                           ▼
┌─────────────────────────────────────────────────────────────┐
│                    Infrastructure Layer                        │
│  ┌─────────────┐     ┌─────────────┐     ┌─────────────┐  │
│  │ PostgreSQL  │     │ USB Printer │     │ ESC/POS     │  │
│  │ Database    │     │ (BP-LITE58) │     │ Protocol    │  │
│  └─────────────┘     └─────────────┘     └─────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

## Core Components

### 1. POS Web (Frontend)
- **Technology**: Next.js, TypeScript, Tailwind CSS
- **Port**: 3000
- **Responsibilities**: 
  - User interface for cashiers
  - Cart management
  - Product browsing
  - Checkout initiation

### 2. POS API (Backend)
- **Technology**: Go, net/http
- **Port**: 8080
- **Responsibilities**:
  - REST API endpoints
  - Request validation
  - Business logic orchestration
  - Error handling

### 3. Print Agent (Printing Service)
- **Technology**: Go, net/http
- **Port**: 8081
- **Responsibilities**:
  - Receipt rendering
  - Printer communication
  - Print job management
  - Idempotency handling

### 4. PostgreSQL Database
- **Technology**: PostgreSQL 17
- **Port**: 5434 (default)
- **Responsibilities**:
  - Data persistence
  - Transaction management
  - Stock management
  - Data consistency

### 5. BP-LITE58 Printer
- **Technology**: Thermal printer, USB interface
- **Protocol**: ESC/POS
- **Responsibilities**:
  - Physical receipt printing
  - Paper cutting
  - Status reporting

## Data Flow

### Checkout Flow

```
1. User Action: Add products to cart (POS Web)
2. API Request: POST /checkout (POS API)
3. Transaction Processing:
   - Begin database transaction
   - Create transaction record
   - Reduce stock with row locking
   - Commit transaction
4. Receipt Generation:
   - Create receipt from transaction
   - Generate print job
5. Print Execution:
   - Send receipt to Print Agent
   - Render ESC/POS commands
   - Send to BP-LITE58 printer
6. Response: Return transaction + print job status
```

### Error Handling Flow

```
1. Critical Error (Database, Stock):
   - Rollback transaction
   - Return error response
   - No state changes
   - Log error details

2. Non-Critical Error (Print):
   - Complete transaction
   - Mark print job as failed
   - Return success with print failure
   - Log for manual recovery
```

## Key Design Principles

### 1. Separation of Concerns
- Clear boundaries between layers
- Each component has single responsibility
- Minimal coupling between components

### 2. Fault Tolerance
- Print failures don't block transactions
- Database failures fail fast
- Graceful degradation where possible

### 3. Data Consistency
- ACID transactions for critical operations
- Row-level locking for stock management
- Idempotency for print operations

### 4. Testability
- Dependency injection for mocking
- Clear interfaces for components
- Comprehensive test coverage

## Technology Rationale

### Go for Backend
- Strong typing and compilation
- Excellent concurrency support
- Built-in HTTP server
- Easy deployment (single binary)

### PostgreSQL for Database
- ACID compliance
- Row-level locking
- JSON support
- Mature ecosystem

### Next.js for Frontend
- React-based with SSR
- TypeScript support
- Built-in routing
- Excellent developer experience

### ESC/POS for Printing
- Industry standard protocol
- Wide printer compatibility
- Simple command structure
- Thermal printer optimization

## Deployment Architecture

### Development Environment
```
localhost:3000 - POS Web
localhost:8080 - POS API  
localhost:8081 - Print Agent
localhost:5434 - PostgreSQL
```

### Production Environment (Recommended)
```
Load Balancer → POS Web (multiple instances)
              → POS API (multiple instances)
              → Print Agent (per printer location)
PostgreSQL (HA setup with replication)
```

## Security Considerations

### Current Implementation
- Basic input validation
- SQL injection prevention (parameterized queries)
- CORS configuration

### Future Enhancements
- Authentication and authorization
- API rate limiting
- Request signing
- Audit logging
- Encryption at rest

## Monitoring and Observability

### Current Implementation
- Structured logging
- Error tracking
- Health check endpoints

### Future Enhancements
- Metrics collection
- Distributed tracing
- Performance monitoring
- Alerting system

## Scalability Considerations

### Horizontal Scaling
- POS API: Stateless, can be scaled horizontally
- POS Web: Stateless, can be scaled horizontally
- Print Agent: Per-location deployment
- Database: Read replicas for scaling reads

### Performance Optimization
- Database connection pooling
- Caching for frequently accessed data
- Async print job processing
- CDN for static assets