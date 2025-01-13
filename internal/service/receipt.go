package service

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"receipt-processor/internal/models"

	"github.com/google/uuid"
)

type ReceiptService struct {
	receipts sync.Map
}

func NewReceiptService() *ReceiptService {
	return &ReceiptService{}
}

func (s *ReceiptService) ProcessReceipt(receipt models.Receipt) (string, error) {
	if err := validateReceipt(receipt); err != nil {
		return "", fmt.Errorf("The receipt is invalid: %w", err)
	}

	id := uuid.New().String()
	points := s.calculatePoints(receipt)
	s.receipts.Store(id, points)

	return id, nil
}

func (s *ReceiptService) GetPoints(id string) (int64, error) {
	points, exists := s.receipts.Load(id)
	if !exists {
		return 0, fmt.Errorf("no receipt found for that id")
	}
	return points.(int64), nil
}

func validateReceipt(receipt models.Receipt) error {
	if !regexp.MustCompile(`^[\w\s\-&]+$`).MatchString(receipt.Retailer) {
		return fmt.Errorf("invalid retailer name")
	}

	if _, err := time.Parse("2006-01-02", receipt.PurchaseDate); err != nil {
		return fmt.Errorf("invalid purchase date")
	}

	if _, err := time.Parse("15:04", receipt.PurchaseTime); err != nil {
		return fmt.Errorf("invalid purchase time")
	}

	if !regexp.MustCompile(`^\d+\.\d{2}$`).MatchString(receipt.Total) {
		return fmt.Errorf("invalid total")
	}

	if len(receipt.Items) == 0 {
		return fmt.Errorf("no items in receipt")
	}

	for _, item := range receipt.Items {
		if !regexp.MustCompile(`^[\w\s\-]+$`).MatchString(item.ShortDescription) {
			return fmt.Errorf("invalid item description")
		}
		if !regexp.MustCompile(`^\d+\.\d{2}$`).MatchString(item.Price) {
			return fmt.Errorf("invalid item price")
		}
	}

	return nil
}

func (s *ReceiptService) calculatePoints(receipt models.Receipt) int64 {
	var points int64

	// Rule 1: One point for every alphanumeric character in the retailer name
	alphanumeric := regexp.MustCompile(`[a-zA-Z0-9]`)
	points += int64(len(alphanumeric.FindAllString(receipt.Retailer, -1)))

	// Rule 2: 50 points if the total is a round dollar amount
	total, _ := strconv.ParseFloat(receipt.Total, 64)
	if math.Mod(total, 1.0) == 0 {
		points += 50
	}

	// Rule 3: 25 points if the total is a multiple of 0.25
	if math.Mod(total*100, 25) == 0 {
		points += 25
	}

	// Rule 4: 5 points for every two items
	points += int64(len(receipt.Items) / 2 * 5)

	// Rule 5: If the trimmed length of the item description is a multiple of 3,
	// multiply the price by 0.2 and round up to the nearest integer
	for _, item := range receipt.Items {
		trimmedDesc := strings.TrimSpace(item.ShortDescription)
		if len(trimmedDesc)%3 == 0 {
			price, _ := strconv.ParseFloat(item.Price, 64)
			points += int64(math.Ceil(price * 0.2))
		}
	}

	// Rule 6: 6 points if the day in the purchase date is odd
	purchaseDate, _ := time.Parse("2006-01-02", receipt.PurchaseDate)
	if purchaseDate.Day()%2 == 1 {
		points += 6
	}

	// Rule 7: 10 points if the time of purchase is between 2:00pm and 4:00pm
	purchaseTime, _ := time.Parse("15:04", receipt.PurchaseTime)
	hour := purchaseTime.Hour()
	if hour >= 14 && hour < 16 {
		points += 10
	}

	return points
}
