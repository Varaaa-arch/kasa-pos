# API Documentation

KASA POS API provides RESTful endpoints for managing products, processing checkouts, and handling transactions. The API is built with Go and follows JSON format for requests and responses.

## Base URL

```
Development: http://localhost:8080
Production: https://api.kasa.pos.com
```

## Authentication

⚠️ **Currently Not Implemented**

The API currently does not require authentication. This is a known limitation that will be addressed in future releases.

### Future Authentication Plan
- JWT-based authentication
- API key support
- Role-based access control
- Rate limiting per user

## Common Headers

```http
Content-Type: application/json
Accept: application/json
```

## Common Response Format

### Success Response
```json
{
  "data": { ... },
  "meta": {
    "timestamp": "2026-08-28T12:00:00Z",
    "request_id": "uuid"
  }
}
```

### Error Response
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid request parameters",
    "details": [
      {
        "field": "quantity",
        "message": "must be greater than 0"
      }
    ],
    "request_id": "uuid",
    "timestamp": "2026-08-28T12:00:00Z"
  }
}
```

## Endpoints

### Products

#### List All Products
```http
GET /products
```

**Response:**
```json
{
  "data": [
    {
      "id": "uuid",
      "sku": "PROD-001",
      "name": "Kopi Susu",
      "price": 15000,
      "stock": 10,
      "created_at": "2026-08-28T10:00:00Z",
      "updated_at": "2026-08-28T10:00:00Z"
    }
  ]
}
```

#### Create Product
```http
POST /products
Content-Type: application/json

{
  "sku": "PROD-002",
  "name": "Teh Manis",
  "price": 8000,
  "stock": 20
}
```

**Response:** `201 Created`
```json
{
  "data": {
    "id": "uuid",
    "sku": "PROD-002",
    "name": "Teh Manis",
    "price": 8000,
    "stock": 20,
    "created_at": "2026-08-28T10:00:00Z",
    "updated_at": "2026-08-28T10:00:00Z"
  }
}
```

#### Get Product by ID
```http
GET /products/:id
```

**Response:** `200 OK`
```json
{
  "data": {
    "id": "uuid",
    "sku": "PROD-001",
    "name": "Kopi Susu",
    "price": 15000,
    "stock": 10,
    "created_at": "2026-08-28T10:00:00Z",
    "updated_at": "2026-08-28T10:00:00Z"
  }
}
```

#### Update Product
```http
PUT /products/:id
Content-Type: application/json

{
  "sku": "PROD-001",
  "name": "Kopi Susu Large",
  "price": 18000,
  "stock": 15
}
```

**Response:** `200 OK`

#### Delete Product
```http
DELETE /products/:id
```

**Response:** `204 No Content`

### Checkout

#### Process Checkout
```http
POST /checkout
Content-Type: application/json

{
  "items": [
    {
      "product_id": "uuid",
      "quantity": 2
    }
  ],
  "paid_amount": 30000,
  "payment_method": "CASH",
  "invoice_number": "INV-001"
}
```

**Response:** `201 Created`
```json
{
  "data": {
    "transaction_id": "uuid",
    "invoice_number": "INV-001",
    "status": "COMPLETED",
    "subtotal": 30000,
    "discount": 0,
    "tax": 0,
    "service_charge": 0,
    "total": 30000,
    "paid_amount": 30000,
    "change": 0,
    "payment_method": "CASH",
    "items": [
      {
        "id": "uuid",
        "product_id": "uuid",
        "sku": "PROD-001",
        "name": "Kopi Susu",
        "quantity": 2,
        "unit_price": 15000,
        "subtotal": 30000
      }
    ],
    "created_at": "2026-08-28T12:00:00Z",
    "print_job": {
      "id": "PJ-uuid",
      "status": "COMPLETED",
      "error": null
    }
  }
}
```

**Error Responses:**

- `400 Bad Request` - Invalid request parameters
- `404 Not Found` - Product not found
- `409 Conflict` - Insufficient stock
- `500 Internal Server Error` - Server error

### Transactions

#### List Transactions
```http
GET /transactions
```

**Query Parameters:**
- `limit` - Number of results (default: 50)
- `offset` - Pagination offset (default: 0)
- `status` - Filter by status (COMPLETED, CANCELLED)

**Response:**
```json
{
  "data": [
    {
      "id": "uuid",
      "invoice_number": "INV-001",
      "status": "COMPLETED",
      "total": 30000,
      "created_at": "2026-08-28T12:00:00Z"
    }
  ],
  "meta": {
    "total": 100,
    "limit": 50,
    "offset": 0
  }
}
```

#### Get Transaction by ID
```http
GET /transactions/:id
```

**Response:** `200 OK`
```json
{
  "data": {
    "id": "uuid",
    "invoice_number": "INV-001",
    "status": "COMPLETED",
    "subtotal": 30000,
    "discount": 0,
    "tax": 0,
    "service_charge": 0,
    "total": 30000,
    "paid_amount": 30000,
    "change": 0,
    "payment_method": "CASH",
    "items": [ ... ],
    "created_at": "2026-08-28T12:00:00Z"
  }
}
```

#### Reprint Receipt
```http
POST /transactions/:id/reprint
```

**Response:** `200 OK`
```json
{
  "data": {
    "transaction_id": "uuid",
    "print_job": {
      "id": "PJ-uuid",
      "status": "COMPLETED",
      "error": null
    }
  }
}
```

## Error Codes

| Code | HTTP Status | Description |
|------|-------------|-------------|
| `VALIDATION_ERROR` | 400 | Request validation failed |
| `PRODUCT_NOT_FOUND` | 404 | Product not found |
| `TRANSACTION_NOT_FOUND` | 404 | Transaction not found |
| `INSUFFICIENT_STOCK` | 409 | Not enough stock for checkout |
| `EMPTY_CART` | 400 | Cannot checkout with empty cart |
| `INSUFFICIENT_CASH` | 400 | Payment amount less than total |
| `INTERNAL_ERROR` | 500 | Internal server error |
| `DATABASE_ERROR` | 500 | Database operation failed |
| `PRINT_AGENT_ERROR` | 500 | Print agent communication failed |

## Rate Limiting

⚠️ **Currently Not Implemented**

Rate limiting will be implemented in future releases to prevent API abuse.

## CORS

The API supports CORS for cross-origin requests from the frontend.

**Allowed Origins:**
- Development: `http://localhost:3000`
- Production: `https://kasa.pos.com`

## Testing the API

### Using cURL

```bash
# List products
curl http://localhost:8080/products

# Create product
curl -X POST http://localhost:8080/products \
  -H "Content-Type: application/json" \
  -d '{
    "sku": "TEST-001",
    "name": "Test Product",
    "price": 10000,
    "stock": 5
  }'

# Process checkout
curl -X POST http://localhost:8080/checkout \
  -H "Content-Type: application/json" \
  -d '{
    "items": [{"product_id": "uuid", "quantity": 1}],
    "paid_amount": 10000,
    "payment_method": "CASH"
  }'
```

### Using HTTPie

```bash
# List products
http GET localhost:8080/products

# Create product
http POST localhost:8080/products \
  sku=TEST-001 \
  name="Test Product" \
  price:=10000 \
  stock:=5

# Process checkout
http POST localhost:8080/checkout \
  items:='[{"product_id": "uuid", "quantity": 1}]' \
  paid_amount:=10000 \
  payment_method=CASH
```

## Webhook Integration

⚠️ **Currently Not Implemented**

Future releases will support webhooks for:
- Transaction completion events
- Low stock alerts
- System health notifications

## Versioning

The API currently uses version 1. Future breaking changes will be released under new version paths:

- Current: `http://localhost:8080/`
- Future: `http://localhost:8080/v2/`

## SDKs

⚠️ **Currently Not Available**

Official SDKs will be provided for:
- JavaScript/TypeScript
- Go
- Python

## Support

For API issues and questions:
- Check the [Architecture Documentation](../architecture/)
- Review [Error Handling](../architecture/ADR-006-api-error-model.md)
- Consult [Testing Guide](../../README.md#testing)