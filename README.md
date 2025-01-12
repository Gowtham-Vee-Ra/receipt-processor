# Receipt Processor

A REST API service that processes receipts and awards points based on defined rules.

## Requirements

- Go 1.21 or higher
- Docker (optional)

## Running the Service

### Using Go

```bash
# Run directly
go run cmd/server/main.go
```

### Using Docker

```bash
# Build and run with Docker
docker build -t receipt-processor .
docker run -p 8080:8080 receipt-processor
```

## API Endpoints

### Process Receipt
POST `/receipts/process`
- Processes a receipt and returns an ID
- Request body: Receipt JSON
- Response: `{ "id": "uuid-string" }`

### Get Points
GET `/receipts/{id}/points`
- Returns points awarded for a receipt
- Response: `{ "points": 100 }`

## Points Rules

1. One point for every alphanumeric character in the retailer name
2. 50 points if the total is a round dollar amount with no cents
3. 25 points if the total is a multiple of 0.25
4. 5 points for every two items on the receipt
5. Multiply the price by 0.2 and round up if the item description length is a multiple of 3
6. 6 points if the day in the purchase date is odd
7. 10 points if the time of purchase is between 2:00pm and 4:00pm

## Testing

```bash
go test ./...
```