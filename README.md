# Flash Sale Order Processing System: From Synchronous to Serverless 

## Overview

This project simulates an **e-commerce flash sale** scenario to explore how system design impacts scalability and customer experience.

We start with a **synchronous order processing system** (3 orders/sec) and evolve it into an **event-driven asynchronous architecture** using **AWS SNS, SQS, and ECS**, before finally migrating to a **serverless design** powered by **AWS Lambda**.

---

## Architecture Evolution

### Phase 1: Synchronous Processing (Baseline)
POST /orders/sync → Verify Payment (3s delay) → Return 200 OK
- Simulates traditional request-response processing.
- Bottleneck: Each order verification takes 3 seconds → limited throughput.
- Tech: Go service deployed on AWS ECS behind an ALB.

---
#### Folder Structure

```
current-system/
├── Terraform/
│   ├── modules/
│   │   ├── ecr/
│   │   ├── ecs/
│   │   ├── logging/
│   │   └── network/
│   ├── main.tf
│   ├── outputs.tf
│   ├── provider.tf
│   └── variables.tf
└── src/
    ├── Dockerfile
    └── main.go
```
### Phase 2: Asynchronous Processing (SNS + SQS + ECS)


    Sync:   Customer → API → Payment (3s) → Response
    Async:  Customer → API → Queue → Response (<100ms)
                               ↓
                       Background Workers → Payment (3s)

- Orders acknowledged instantly (202 Accepted).
- Payment processing happens in the background.
- Scalable via additional ECS workers (goroutines).

**Tech:**
- SNS topic: `order-processing-events`
- SQS queue: `order-processing-queue`
- ECS services:
    - `order-receiver` (API service)
    - `order-processor` (background worker)

---

#### Folder Structure

```
Async-Solution/
├── Terraform/
│   ├── modules/
│   │   ├── ecr/
│   │   ├── ecs/
│   │   ├── lambda/
│   │   ├── logging/
│   │   ├── network/
│   │   └── sns/
│   │   └── sqs/
│   ├── main.tf
│   ├── outputs.tf
│   ├── provider.tf
│   └── variables.tf
└── src/
    ├── order-processor/
    │   ├── Dockerfile
    │   └── main.go
    └── order-receiver/
        ├── Dockerfile
        └── main.go
```

### Phase 3: Serverless Processing (SNS → Lambda)
Client → /orders/async → SNS → Lambda

- Eliminates queues and ECS worker management.
- AWS Lambda automatically scales on demand.
- Pay-per-use model with free tier coverage up to ~267K orders/month.

#### Folder Structure

```
Serverless/
├── Terraform/
│   ├── modules/
│   │   ├── ecr/
│   │   ├── ecs/
│   │   ├── lambda/
│   │   ├── logging/
│   │   ├── network/
│   │   └── sns/
│   ├── main.tf
│   ├── outputs.tf
│   ├── provider.tf
│   └── variables.tf
└── src/
    ├── order-processor-lambda/
    │   └── main.go
    └── order-receiver/
        ├── Dockerfile
        └── main.go
```
---

## Infrastructure Setup

**Provisioned via Terraform:**
- VPC (10.0.0.0/16)
- 2 Public Subnets (for ALB)
- 2 Private Subnets (for ECS tasks)
- Application Load Balancer (ALB)
- ECS Cluster + Services (Receiver, Processor)
- SNS Topic + SQS Queue
- Lambda Function (Phase 3)

---

## Application Endpoints

| Endpoint       | Method | Description                         |
|----------------|--------|-------------------------------------|
| `/orders/sync` | POST   | Synchronous order with 3s payment delay |
| `/orders/async`| POST   | Publishes order to SNS, returns immediately |

---

## Load Testing with Locust

| Scenario     | Users | Duration | Expected Result                     |
|-------------|-------|----------|------------------------------------|
| Normal      | 5     | 30s      | 100% success                        |
| Flash Sale  | 20    | 60s      | Sync: Failures, Async: 100% accepted |

**Locust Config:**
- Spawn rate: 1 (normal), 10 (flash)
- Wait time: 100–500ms between requests
- Target endpoint: `/orders/sync` or `/orders/async`

---

## CloudWatch Monitoring

**Metrics observed:**
- `ApproximateNumberOfMessagesVisible` (SQS queue depth)
- `NumberOfMessagesDeleted` (processed count)
- `Lambda Duration` and `Init Duration` (for cold start analysis)

**Captured graphs:**
1. Queue spike during flash sale
2. Gradual drain post-load
3. Lambda cold start overhead (~70ms once per idle period)

---

## Serverless Evaluation (Lambda Phase)

- **Cold Starts:** Occur after ~5 minutes idle
- **Avg Init Duration:** ~73ms
- **Overhead:** ~2.4% on a 3-second task
- Negligible for long-running (3s) tasks

---

## How to Run

### 1. Configure AWS CLI
```bash
aws configure
```
Set your AWS Access Key, Secret Key, and Region  
Optionally, set a session token if using temporary credentials

### 2. Initialize Terraform
```bash
terraform init
```

### 3. Plan Terraform Deployment
```bash
terraform plan
```
Review resources that will be created

### 4. Apply Terraform Deployment
```bash
terraform apply
```
Confirm apply with `yes` or use `-auto-approve`
##  Deliverables

-  Terraform Infrastructure (VPC, ALB, ECS, SNS, SQS, Lambda)
-  Go Services (`/orders/sync`, `/orders/async`)
-  Locust Load Tests
-  CloudWatch Metrics (SQS Queue Depth, Lambda Logs)

---

## Key Takeaways

- Synchronous systems collapse under high load.
- Event-driven design ensures resilience via decoupling.
- Queues add reliability but require tuning.
- Serverless simplifies everything—ideal for startups scaling fast.
