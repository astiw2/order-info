# Order Information Processing System - High Level Architecture

## Goals
- Accept batched order submissions via REST (single request with 1..N orders for 1..M customers)
- Validate order schema strictly (reject if unexpected/missing keys)
- Transform orders to two views:
  1. Per-customer flat list of purchased items
  2. Global customer summaries (item count, total amount in EUR)

## High level diagram

<img src="resources/images/batch-processor.png" alt="Batch Processor" />

## Constraints / Assumptions
- No session handling or caching layer will be considered for this task
- Timestamp is passed through as string. We will not parse it and consider further for any schema validation
- Schema validation will be considered strictly based upon the key name and associated data type
- Pagination has not been considered for the sake of simlicity
- Currency handling is limited to whole euros for the sake of simplicity

## Components

### 1. HTTP Layer
- Endpoints:
  - POST /orders/info - accept batch of orders in the provided format. Returns 200OK with customer items and validation errors for erroneous orders if found.

### 2. Validation & Parsing
- Strict JSON decoder to ensure no unexpected keys.

### 3. Transformation
- **Input**: `[]Order` - batch of customer orders
- **Processing**: 
  - Validates each order using `ValidateOrder` function
  - Transforms valid orders into flattened customer items
  - Aggregates data per customer for summary generation
- **Output**:
  - `[]CustomerItem` - flat list of all purchased items with customer ID
  - `[]CustomerSummary` - aggregated per-customer data
  - `[]ValidationError` - detailed errors for invalid orders

### 4. Error Handling Strategy
- Collect per-order validation error.

## Development

### Running the Server

To start the order processing server, use the provided Makefile:

```bash
make dev
```

This command will:
1. Clean any previous build artifacts
2. Build the application binary
3. Start the server on `http://localhost:8080` provided the port is not already binded with any other application

### Testing the API

Once the server is running, you can test the POST endpoint using curl. 
Install [jq](https://jqlang.org/) for formatted response.

```bash
curl -X POST http://localhost:8080/orders/info \
  -H 'Content-Type: application/json' \
  -d '[
    {
      "customerId": "C1",
      "orderId": "O1", 
      "timestamp": "1730419200000",
      "items": [
        { "itemId": "I100", "costEur": 5 },
        { "itemId": "I101", "costEur": 3 }
      ]
    },
    {
      "customerId": "10",
      "orderId": "O2",
      "timestamp": "1730419205000", 
      "items": [
        { "itemid": "I200", "costEur": 7 }
      ]
    }
  ]' | jq .
```

A detailed Order Management Subsystem following a DDD (Domain-Driven Design) approach is documented [here](./resources/docs/DDD.md). Please read it at your own leisure or ignore.
