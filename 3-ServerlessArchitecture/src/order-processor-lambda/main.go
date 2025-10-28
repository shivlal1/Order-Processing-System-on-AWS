package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

// Order represents an e-commerce order
type Order struct {
	OrderID    string    `json:"order_id"`
	CustomerID int       `json:"customer_id"`
	Status     string    `json:"status"`
	Items      []Item    `json:"items"`
	CreatedAt  time.Time `json:"created_at"`
}

// Item represents an item in an order
type Item struct {
	ProductID string  `json:"product_id"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
}

// ProcessPayment simulates payment processing with 3-second delay
func ProcessPayment(orderID string) error {
	log.Printf("Processing payment for order %s...", orderID)
	time.Sleep(3 * time.Second)
	log.Printf("Payment processed for order %s", orderID)
	return nil
}

// HandleRequest processes SNS events containing orders
func HandleRequest(ctx context.Context, snsEvent events.SNSEvent) error {
	log.Printf("Received %d records", len(snsEvent.Records))

	for _, record := range snsEvent.Records {
		// Extract the order from the SNS message
		var order Order
		if err := json.Unmarshal([]byte(record.SNS.Message), &order); err != nil {
			log.Printf("Failed to parse order: %v", err)
			return fmt.Errorf("failed to parse order: %w", err)
		}

		log.Printf("Processing order %s from customer %d", order.OrderID, order.CustomerID)

		// Process payment (3-second delay)
		startTime := time.Now()
		if err := ProcessPayment(order.OrderID); err != nil {
			log.Printf("Payment processing failed for order %s: %v", order.OrderID, err)
			return fmt.Errorf("payment processing failed: %w", err)
		}

		processingTime := time.Since(startTime)
		log.Printf("Order %s completed in %.2f seconds", order.OrderID, processingTime.Seconds())

		// Log order details for monitoring
		itemCount := len(order.Items)
		var totalValue float64
		for _, item := range order.Items {
			totalValue += item.Price * float64(item.Quantity)
		}

		log.Printf("Order summary - ID: %s, Items: %d, Total: $%.2f",
			order.OrderID, itemCount, totalValue)
	}

	return nil
}

func main() {
	// Start the Lambda handler
	lambda.Start(HandleRequest)
}
