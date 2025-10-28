package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

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
	// Buffered channel to limit concurrent payment processing
	// Size of 1 ensures only one payment processes at a time
	semaphore chan struct{}
}

// NewPaymentProcessor creates a payment processor with limited throughput
func NewPaymentProcessor() *PaymentProcessor {
	return &PaymentProcessor{
		semaphore: make(chan struct{}, 1),
	}
}

// ProcessPayment simulates payment verification with 3 second delay
func (p *PaymentProcessor) ProcessPayment(orderID string) error {
	// Acquire semaphore (blocks if another payment is processing)
	p.semaphore <- struct{}{}
	defer func() { <-p.semaphore }()

	log.Printf("Processing payment for order %s...", orderID)

	// Simulate the 3-second payment verification delay
	time.Sleep(3 * time.Second)

	log.Printf("Payment processed for order %s", orderID)
	return nil
}

// OrderHandler handles order requests
type OrderHandler struct {
	paymentProcessor *PaymentProcessor
	stats            *Stats
}

// Stats tracks system metrics
type Stats struct {
	mu               sync.Mutex
	totalRequests    int
	successfulOrders int
	failedOrders     int
}

// NewOrderHandler creates a new order handler
func NewOrderHandler(processor *PaymentProcessor) *OrderHandler {
	return &OrderHandler{
		paymentProcessor: processor,
		stats:            &Stats{},
	}
}

// HandleSyncOrder processes orders synchronously
func (h *OrderHandler) HandleSyncOrder(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()

	// Increment request counter
	h.stats.mu.Lock()
	h.stats.totalRequests++
	h.stats.mu.Unlock()

	// Parse order from request body
	var order Order
	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		h.stats.mu.Lock()
		h.stats.failedOrders++
		h.stats.mu.Unlock()

		http.Error(w, "Invalid order format", http.StatusBadRequest)
		return
	}

	// Generate order ID if not provided
	if order.OrderID == "" {
		order.OrderID = uuid.New().String()
	}
	order.CreatedAt = time.Now()
	order.Status = "processing"

	// Process payment (this blocks for 3 seconds)
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

	// Payment successful
	order.Status = "completed"

	h.stats.mu.Lock()
	h.stats.successfulOrders++
	h.stats.mu.Unlock()

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]interface{}{
		"order_id":        order.OrderID,
		"status":          order.Status,
		"message":         "Order processed successfully",
		"processing_time": time.Since(startTime).Seconds(),
	}
	json.NewEncoder(w).Encode(response)

	log.Printf("Order %s completed in %.2f seconds", order.OrderID, time.Since(startTime).Seconds())
}

// HandleHealth returns health status
func (h *OrderHandler) HandleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "healthy",
		"service": "order-processor-sync",
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
		"successful_orders": h.stats.successfulOrders,
		"failed_orders":     h.stats.failedOrders,
		"success_rate":      float64(h.stats.successfulOrders) / float64(h.stats.totalRequests) * 100,
	})
}

func main() {
	// Create payment processor with bottleneck
	paymentProcessor := NewPaymentProcessor()

	// Create order handler
	orderHandler := NewOrderHandler(paymentProcessor)

	// Setup routes
	router := mux.NewRouter()
	router.HandleFunc("/orders/sync", orderHandler.HandleSyncOrder).Methods("POST")
	router.HandleFunc("/health", orderHandler.HandleHealth).Methods("GET")
	router.HandleFunc("/stats", orderHandler.HandleStats).Methods("GET")

	// Start server
	port := ":8080"
	log.Printf("Starting synchronous order processor on port %s", port)
	log.Printf("Payment processing bottleneck: 3 seconds per order")
	log.Printf("Expected behavior under load: requests will queue and timeout")

	if err := http.ListenAndServe(port, router); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
