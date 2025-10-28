# Order Processing Architecture Comparison

## Architecture I - Synchronous Order Processing
**Design**: Customer → API → Payment (3s) → Response  
Each request blocks for the full payment processing time, creating a bottleneck at 0.33 orders/second maximum throughput.

| Users | Median (ms) | 95%ile (ms) | 99%ile (ms) | Min (ms) | Max (ms) |
|-------|-------------|-------------|-------------|----------|----------|
| 5     | 14,873      | 15,000      | 15,000      | 3,006    | 14,873   | 
| 20    | 30,000      | 59,000      | 59,000      | 3,006    | 59,038   | 

## Architecture II - Async with SNS/SQS
**Design**: Customer → API → SNS → SQS → ECS Workers → Payment (3s)  
API returns immediately (<100ms), orders queue in SQS for background processing by configurable worker goroutines.

### Load Test Results - 20 Users with Varying Goroutines
| Users | Goroutines | Median (ms) | 95%ile (ms) | 99%ile (ms) | Min (ms) | Max (ms) | Avg (ms) | 
|-------|------------|----------|-------------|-------------|-------------|----------|----------|
| 20    | 1          | 15          | 26          | 72          | 11       | 158      | 16.93    | 
| 20    | 5          | 15          | 25          | 71          | 11       | 176      | 16.6     | 
| 20    | 20         | 15          | 24          | 69          | 12       | 198      | 17.06    | 
| 20    | 100        | 14          | 27          | 77          | 11       | 152      | 16.64    |

### SQS Queue Performance - 20 Users Load Test
| Goroutines | Peak Queue Depth (messages) | Time to Clear Queue|
|------------|----------------------------|---------------------|
| 1          | 4.52K                      | ~3.75 hours        | 
| 5          | 3.94K                      | ~40 minutes        | 
| 20         | 1.19K                      | ~3 minutes         | 
| 100        | 900                        | <1 minute          | 

## Architecture III - Serverless with Lambda
**Design**: Customer → API → SNS → Lambda → Payment (3s)  
Lambda auto-scales from 0 to thousands of concurrent executions, no queue management needed.

### Analysis

1. **How often did cold starts occur?**  
   First request only, then after ~5-15 minutes of inactivity. Active use kept Lambda warm.

2. **Is the cost advantage compelling?**  
   Yes. Lambda: $0 for 10K orders/month. ECS: $17/month. Break-even at 1.7M requests/month.

3. **Can you accept losing SQS guarantees?**  
   Yes. SNS gives 2 retries which is sufficient. Lost: queue visibility, custom retry logic.

4. **Scale consideration:**  
   Free until 267K orders/month. No capacity planning needed with auto-scaling.

### Recommendation

**Should your startup switch to Lambda?**  
Yes. Cold starts (70ms on 3s process = 2.3%) are negligible and occur rarely. The cost savings ($0 vs $17/month), elimination of queue management, and automatic scaling far outweigh the minor loss of queue control. For a startup under 267K orders/month, Lambda's operational simplicity lets you focus on product instead of infrastructure.
