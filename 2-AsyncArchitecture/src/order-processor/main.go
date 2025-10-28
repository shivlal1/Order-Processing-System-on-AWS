package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
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

// ProcessorStats tracks processing metrics
type ProcessorStats struct {
	mu                sync.Mutex
	messagesReceived  int64
	messagesProcessed int64
	messagesFailed    int64
	startTime         time.Time
}

// OrderProcessor handles SQS messages and payment processing
type OrderProcessor struct {
	sqsClient     *sqs.Client
	queueURL      string
	workerCount   int
	stats         *ProcessorStats
	activeWorkers int32
}

func NewOrderProcessor(sqsClient *sqs.Client, queueURL string, workerCount int) *OrderProcessor {
	return &OrderProcessor{
		sqsClient:   sqsClient,
		queueURL:    queueURL,
		workerCount: workerCount,
		stats: &ProcessorStats{
			startTime: time.Now(),
		},
	}
}

// ProcessPayment simulates payment processing with 3-second delay
func (p *OrderProcessor) ProcessPayment(orderID string) error {
	log.Printf("Worker processing payment for order %s...", orderID)
	time.Sleep(3 * time.Second)
	log.Printf("Payment processed for order %s", orderID)
	return nil
}

// ProcessMessage handles a single SQS message
func (p *OrderProcessor) ProcessMessage(ctx context.Context, message types.Message) {
	atomic.AddInt64(&p.stats.messagesReceived, 1)

	// Extract order from SNS message wrapper
	var snsMessage struct {
		Message string `json:"Message"`
	}

	if err := json.Unmarshal([]byte(*message.Body), &snsMessage); err != nil {
		log.Printf("Failed to parse SNS message: %v", err)
		atomic.AddInt64(&p.stats.messagesFailed, 1)
		return
	}

	// Parse the actual order
	var order Order
	if err := json.Unmarshal([]byte(snsMessage.Message), &order); err != nil {
		log.Printf("Failed to parse order: %v", err)
		atomic.AddInt64(&p.stats.messagesFailed, 1)
		return
	}

	// Process payment
	startTime := time.Now()
	if err := p.ProcessPayment(order.OrderID); err != nil {
		log.Printf("Payment processing failed for order %s: %v", order.OrderID, err)
		atomic.AddInt64(&p.stats.messagesFailed, 1)
		return
	}

	// Delete message from queue after successful processing
	deleteInput := &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(p.queueURL),
		ReceiptHandle: message.ReceiptHandle,
	}

	if _, err := p.sqsClient.DeleteMessage(ctx, deleteInput); err != nil {
		log.Printf("Failed to delete message: %v", err)
		// Message will become visible again after visibility timeout
	}

	atomic.AddInt64(&p.stats.messagesProcessed, 1)
	log.Printf("Order %s processed in %.2f seconds", order.OrderID, time.Since(startTime).Seconds())
}

// Worker polls SQS and processes messages
func (p *OrderProcessor) Worker(ctx context.Context, workerID int) {
	atomic.AddInt32(&p.activeWorkers, 1)
	defer atomic.AddInt32(&p.activeWorkers, -1)

	log.Printf("Worker %d started", workerID)

	for {
		select {
		case <-ctx.Done():
			log.Printf("Worker %d stopping", workerID)
			return
		default:
			// Poll SQS for messages
			receiveInput := &sqs.ReceiveMessageInput{
				QueueUrl:            aws.String(p.queueURL),
				MaxNumberOfMessages: 10,
				WaitTimeSeconds:     20, // Long polling
				VisibilityTimeout:   30,
			}

			result, err := p.sqsClient.ReceiveMessage(ctx, receiveInput)
			if err != nil {
				log.Printf("Worker %d: Failed to receive messages: %v", workerID, err)
				time.Sleep(5 * time.Second)
				continue
			}

			// Process each message
			for _, message := range result.Messages {
				p.ProcessMessage(ctx, message)
			}
		}
	}
}

// Start begins processing with configured number of workers
func (p *OrderProcessor) Start(ctx context.Context) {
	log.Printf("Starting order processor with %d workers", p.workerCount)

	var wg sync.WaitGroup

	// Start worker goroutines
	for i := 0; i < p.workerCount; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			p.Worker(ctx, workerID)
		}(i)
	}

	// Start stats reporter
	go p.ReportStats(ctx)

	// Wait for all workers to finish
	wg.Wait()
	log.Println("All workers stopped")
}

// ReportStats periodically logs processing statistics
func (p *OrderProcessor) ReportStats(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			p.stats.mu.Lock()
			received := atomic.LoadInt64(&p.stats.messagesReceived)
			processed := atomic.LoadInt64(&p.stats.messagesProcessed)
			failed := atomic.LoadInt64(&p.stats.messagesFailed)
			uptime := time.Since(p.stats.startTime)
			activeWorkers := atomic.LoadInt32(&p.activeWorkers)
			p.stats.mu.Unlock()

			rate := float64(processed) / uptime.Seconds()

			log.Printf("=== PROCESSOR STATS ===")
			log.Printf("Uptime: %.0f seconds", uptime.Seconds())
			log.Printf("Active Workers: %d/%d", activeWorkers, p.workerCount)
			log.Printf("Messages Received: %d", received)
			log.Printf("Messages Processed: %d", processed)
			log.Printf("Messages Failed: %d", failed)
			log.Printf("Processing Rate: %.2f orders/second", rate)
			log.Printf("====================")
		}
	}
}

func main() {
	// Get configuration from environment
	queueURL := os.Getenv("SQS_QUEUE_URL")
	if queueURL == "" {
		log.Fatal("SQS_QUEUE_URL environment variable not set")
	}

	// Get worker count from environment (default to 1)
	workerCount := 1
	if wc := os.Getenv("WORKER_COUNT"); wc != "" {
		if count, err := strconv.Atoi(wc); err == nil && count > 0 {
			workerCount = count
		}
	}

	// Initialize AWS SDK
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatal("Unable to load AWS SDK config:", err)
	}

	sqsClient := sqs.NewFromConfig(cfg)

	// Create processor
	processor := NewOrderProcessor(sqsClient, queueURL, workerCount)

	// Start processing
	ctx := context.Background()
	log.Printf("Starting order processor service")
	log.Printf("SQS Queue: %s", queueURL)
	log.Printf("Worker Count: %d", workerCount)
	log.Printf("Each worker can process 1 order every 3 seconds")
	log.Printf("Maximum throughput: %.2f orders/second", float64(workerCount)/3.0)

	processor.Start(ctx)
}
