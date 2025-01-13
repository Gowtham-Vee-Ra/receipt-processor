package service

import (
	"math"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"receipt-processor/internal/models"

	"github.com/stretchr/testify/assert"
)

func TestCalculatePoints(t *testing.T) {
	svc := NewReceiptService()

	tests := []struct {
		name     string
		receipt  models.Receipt
		expected int64
	}{
		{
			name: "target example",
			receipt: models.Receipt{
				Retailer:     "Target",
				PurchaseDate: "2022-01-01",
				PurchaseTime: "13:01",
				Items: []models.Item{
					{ShortDescription: "Mountain Dew 12PK", Price: "6.49"},
					{ShortDescription: "Emils Cheese Pizza", Price: "12.25"},
					{ShortDescription: "Knorr Creamy Chicken", Price: "1.26"},
					{ShortDescription: "Doritos Nacho Cheese", Price: "3.35"},
					{ShortDescription: "Klarbrunn 12-PK 12 FL OZ", Price: "12.00"},
				},
				Total: "35.35",
			},
			expected: 28,
		},
		{
			name: "m&m corner market example",
			receipt: models.Receipt{
				Retailer:     "M&M Corner Market",
				PurchaseDate: "2022-03-20",
				PurchaseTime: "14:33",
				Items: []models.Item{
					{ShortDescription: "Gatorade", Price: "2.25"},
					{ShortDescription: "Gatorade", Price: "2.25"},
					{ShortDescription: "Gatorade", Price: "2.25"},
					{ShortDescription: "Gatorade", Price: "2.25"},
				},
				Total: "9.00",
			},
			expected: 109,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			points := svc.calculatePoints(tt.receipt)
			if points != tt.expected {
				// Print detailed point calculation
				alphanumeric := len(regexp.MustCompile(`[a-zA-Z0-9]`).FindAllString(tt.receipt.Retailer, -1))
				t.Logf("- Retailer name points: %d (one point per alphanumeric char in '%s')", alphanumeric, tt.receipt.Retailer)

				total, _ := strconv.ParseFloat(tt.receipt.Total, 64)
				if math.Mod(total, 1.0) == 0 {
					t.Log("- Round dollar amount: 50")
				}

				if math.Mod(total*100, 25) == 0 && math.Mod(total, 1.0) != 0 {
					t.Log("- Multiple of 0.25: 25")
				}

				itemPairs := len(tt.receipt.Items) / 2
				t.Logf("- Item pairs: %d (5 points per pair)", itemPairs*5)

				for _, item := range tt.receipt.Items {
					trimmedLen := len(strings.TrimSpace(item.ShortDescription))
					if trimmedLen%3 == 0 {
						price, _ := strconv.ParseFloat(item.Price, 64)
						itemPoints := int64(math.Ceil(price * 0.2))
						t.Logf("- Item description multiple of 3: %d points for '%s'", itemPoints, item.ShortDescription)
					}
				}

				purchaseDate, _ := time.Parse("2006-01-02", tt.receipt.PurchaseDate)
				if purchaseDate.Day()%2 == 1 && tt.receipt.Retailer != "Target" {
					t.Log("- Odd day: 6")
				}

				purchaseTime, _ := time.Parse("15:04", tt.receipt.PurchaseTime)
				hour := purchaseTime.Hour()
				if hour >= 14 && hour < 16 {
					t.Log("- Between 2:00 PM and 4:00 PM: 10")
				}
			}
			assert.Equal(t, tt.expected, points)
		})
	}
}

func TestProcessReceipt(t *testing.T) {
	service := NewReceiptService()

	receipt := models.Receipt{
		Retailer:     "Target",
		PurchaseDate: "2022-01-01",
		PurchaseTime: "13:01",
		Items: []models.Item{
			{ShortDescription: "Mountain Dew 12PK", Price: "6.49"},
		},
		Total: "6.49",
	}

	// Test successful processing
	id, err := service.ProcessReceipt(receipt)
	assert.NoError(t, err)
	assert.NotEmpty(t, id)

	// Test points retrieval
	points, err := service.GetPoints(id)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, points, int64(0))

	// Test invalid receipt ID
	_, err = service.GetPoints("invalid-id")
	assert.Error(t, err)
}

func TestValidateReceipt(t *testing.T) {
	tests := []struct {
		name    string
		receipt models.Receipt
		wantErr bool
	}{
		{
			name: "valid receipt",
			receipt: models.Receipt{
				Retailer:     "Target",
				PurchaseDate: "2022-01-01",
				PurchaseTime: "13:01",
				Items: []models.Item{
					{ShortDescription: "Mountain Dew 12PK", Price: "6.49"},
				},
				Total: "6.49",
			},
			wantErr: false,
		},
		{
			name: "invalid retailer",
			receipt: models.Receipt{
				Retailer:     "Target@123",
				PurchaseDate: "2022-01-01",
				PurchaseTime: "13:01",
				Items: []models.Item{
					{ShortDescription: "Mountain Dew 12PK", Price: "6.49"},
				},
				Total: "6.49",
			},
			wantErr: true,
		},
		{
			name: "invalid date",
			receipt: models.Receipt{
				Retailer:     "Target",
				PurchaseDate: "2022-13-01", // invalid month
				PurchaseTime: "13:01",
				Items: []models.Item{
					{ShortDescription: "Mountain Dew 12PK", Price: "6.49"},
				},
				Total: "6.49",
			},
			wantErr: true,
		},
		{
			name: "invalid price format",
			receipt: models.Receipt{
				Retailer:     "Target",
				PurchaseDate: "2022-01-01",
				PurchaseTime: "13:01",
				Items: []models.Item{
					{ShortDescription: "Mountain Dew 12PK", Price: "6.4"}, // missing cent
				},
				Total: "6.49",
			},
			wantErr: true,
		},
	}

	svc := NewReceiptService()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.ProcessReceipt(tt.receipt)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
