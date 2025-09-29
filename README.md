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
- Go in memory storage of order data per runtime instance is used purely to answer fetch requests
- No session handling or caching layer will be considered for this task
- Timestamp is passed through as string. We will not parse it and consider further for any schema validation
- Invalid orders are ignored (could alternatively reject whole batch but we choose to skip invalid and just log errors)
- Schema validation will be considered strictly based upon the key name and associated data type
- Pagination has not been considered while drafting this solution for the sake of simlicity

## Components

### 1. HTTP Layer
- Endpoints:
  - POST /orders/info - accept batch of orders in the provided format. Returns 200OK with customer items OR errors.
  - GET /customers/{id}/items - return list of all items purchased by an individual customer.
  - GET /customers/summary - list of summaries including all customers.

### 2. Validation & Parsing
- Strict JSON decoder to ensure no unexpected keys.

### 3. Transformation Service(Conceptual at the time of drafting this markdown)
- Input: []Order
- Output:
  - []CustomerItem
  - []CustomerSummary aggregated per customer.

### 4. Error Handling Strategy
- Collect per-order validation error.
- If ALL orders invalid -> return 400.

A detailed Order Management Subsystem following a DDD (Domain-Driven Design) approach is documented [here](./resources/docs/DDD.md). Please read it at your own leisure or ignore.
