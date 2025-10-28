package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sns/types"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

// Order represents an e-commerce order
type Order struct {
	OrderID    string    `json:"order_id"`
	CustomerID int       `json:"customer_id"`
	Status     string    `json:"status"` // pending, processing, completed
	Items      []Item    `json:"items"`
	CreatedAt  time.Time `json:"created_at"`
}

// Item represents an item in an order
type Item struct {
	ProductID string  `json:"product_id"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
}

// PaymentProcessor simulates the bottleneck with limited throughput
type PaymentProcessor struct {
	semaphore chan struct{}
}

func NewPaymentProcessor() *PaymentProcessor {
	return &PaymentProcessor{
		semaphore: make(chan struct{}, 1),
	}
}

func (p *PaymentProcessor) ProcessPayment(orderID string) error {
	p.semaphore <- struct{}{}
	defer func() { <-p.semaphore }()

	log.Printf("Processing payment for order %s...", orderID)
	time.Sleep(3 * time.Second)
	log.Printf("Payment processed for order %s", orderID)
	return nil
}

// OrderHandler handles order requests
type OrderHandler struct {
	paymentProcessor *PaymentProcessor
	snsClient        *sns.Client
	topicArn         string
	stats            *Stats
}

// Stats tracks system metrics
type Stats struct {
	mu               sync.Mutex
	totalRequests    int
	syncOrders       int
	asyncOrders      int
	successfulOrders int
	failedOrders     int
}

func NewOrderHandler(processor *PaymentProcessor, snsClient *sns.Client, topicArn string) *OrderHandler {
	return &OrderHandler{
		paymentProcessor: processor,
		snsClient:        snsClient,
		topicArn:         topicArn,
		stats:            &Stats{},
	}
}

// HandleSyncOrder processes orders synchronously (existing endpoint)
func (h *OrderHandler) HandleSyncOrder(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	h.stats.mu.Lock()
	h.stats.totalRequests++
	h.stats.syncOrders++
	h.stats.mu.Unlock()

	var order Order
	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		h.stats.mu.Lock()
		h.stats.failedOrders++
		h.stats.mu.Unlock()

		http.Error(w, "Invalid order format", http.StatusBadRequest)
		return
	}

	if order.OrderID == "" {
		order.OrderID = uuid.New().String()
	}
	order.CreatedAt = time.Now()
	order.Status = "processing"

	// Process payment synchronously (blocks for 3 seconds)
	if err := h.paymentProcessor.ProcessPayment(order.OrderID); err != nil {
		h.stats.mu.Lock()
		h.stats.failedOrders++
		h.stats.mu.Unlock()

		order.Status = "failed"
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error":    "Payment processing failed",
			"order_id": order.OrderID,
		})
		return
	}

	order.Status = "completed"

	h.stats.mu.Lock()
	h.stats.successfulOrders++
	h.stats.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]interface{}{
		"order_id":        order.OrderID,
		"status":          order.Status,
		"message":         "Order processed successfully",
		"processing_time": time.Since(startTime).Seconds(),
		"processing_mode": "synchronous",
	}
	json.NewEncoder(w).Encode(response)

	log.Printf("Sync order %s completed in %.2f seconds", order.OrderID, time.Since(startTime).Seconds())
}

// HandleAsyncOrder accepts orders and publishes to SNS for async processing
func (h *OrderHandler) HandleAsyncOrder(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	h.stats.mu.Lock()
	h.stats.totalRequests++
	h.stats.asyncOrders++
	h.stats.mu.Unlock()

	var order Order
	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		h.stats.mu.Lock()
		h.stats.failedOrders++
		h.stats.mu.Unlock()

		http.Error(w, "Invalid order format", http.StatusBadRequest)
		return
	}

	if order.OrderID == "" {
		order.OrderID = uuid.New().String()
	}
	order.CreatedAt = time.Now()
	order.Status = "accepted"

	// Publish order to SNS for async processing
	orderJSON, err := json.Marshal(order)
	if err != nil {
		h.stats.mu.Lock()
		h.stats.failedOrders++
		h.stats.mu.Unlock()

		http.Error(w, "Failed to marshal order", http.StatusInternalServerError)
		return
	}

	input := &sns.PublishInput{
		Message:  aws.String(string(orderJSON)),
		TopicArn: aws.String(h.topicArn),
		MessageAttributes: map[string]types.MessageAttributeValue{
			"order_id": {
				DataType:    aws.String("String"),
				StringValue: aws.String(order.OrderID),
			},
		},
	}

	_, err = h.snsClient.Publish(context.TODO(), input)
	if err != nil {
		h.stats.mu.Lock()
		h.stats.failedOrders++
		h.stats.mu.Unlock()

		log.Printf("Failed to publish to SNS: %v", err)
		http.Error(w, "Failed to queue order", http.StatusInternalServerError)
		return
	}

	h.stats.mu.Lock()
	h.stats.successfulOrders++
	h.stats.mu.Unlock()

	// Return immediately with 202 Accepted
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)

	response := map[string]interface{}{
		"order_id":        order.OrderID,
		"status":          "accepted",
		"message":         "Order accepted for processing",
		"processing_time": time.Since(startTime).Seconds(),
		"processing_mode": "asynchronous",
	}
	json.NewEncoder(w).Encode(response)

	log.Printf("Async order %s accepted in %.4f seconds", order.OrderID, time.Since(startTime).Seconds())
}

// HandleHealth returns health status
func (h *OrderHandler) HandleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "healthy",
		"service": "order-receiver",
	})
}

// HandleStats returns system statistics
func (h *OrderHandler) HandleStats(w http.ResponseWriter, r *http.Request) {
	h.stats.mu.Lock()
	defer h.stats.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"total_requests":    h.stats.totalRequests,
		"sync_orders":       h.stats.syncOrders,
		"async_orders":      h.stats.asyncOrders,
		"successful_orders": h.stats.successfulOrders,
		"failed_orders":     h.stats.failedOrders,
		"success_rate":      float64(h.stats.successfulOrders) / float64(h.stats.totalRequests) * 100,
	})
}

func main() {
	// Get SNS topic ARN from environment
	topicArn := os.Getenv("SNS_TOPIC_ARN")
	if topicArn == "" {
		log.Fatal("SNS_TOPIC_ARN environment variable not set")
	}

	// Initialize AWS SDK
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatal("Unable to load AWS SDK config:", err)
	}

	snsClient := sns.NewFromConfig(cfg)

	// Create payment processor with bottleneck
	paymentProcessor := NewPaymentProcessor()

	// Create order handler
	orderHandler := NewOrderHandler(paymentProcessor, snsClient, topicArn)

	// Setup routes
	router := mux.NewRouter()
	router.HandleFunc("/orders/sync", orderHandler.HandleSyncOrder).Methods("POST")
	router.HandleFunc("/orders/async", orderHandler.HandleAsyncOrder).Methods("POST")
	router.HandleFunc("/health", orderHandler.HandleHealth).Methods("GET")
	router.HandleFunc("/stats", orderHandler.HandleStats).Methods("GET")

	// Start server
	port := ":8080"
	log.Printf("Starting order receiver service on port %s", port)
	log.Printf("SNS Topic: %s", topicArn)
	log.Printf("Endpoints: /orders/sync (3s delay) and /orders/async (<100ms)")

	if err := http.ListenAndServe(port, router); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
