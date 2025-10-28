import random
import json
from locust import HttpUser, task, between

class OrderUser(HttpUser):
    # Random wait time between 100-500ms
    wait_time = between(0.1, 0.5)

    def on_start(self):
        """Initialize user with a customer ID"""
        self.customer_id = random.randint(1000, 9999)
        self.order_count = 0

    @task
    def place_sync_order(self):
        self.order_count += 1

        # Generate order data
        order_data = {
            "customer_id": self.customer_id,
            "items": [
                {
                    "product_id": f"PROD-{random.randint(100, 999)}",
                    "quantity": random.randint(1, 5),
                    "price": round(random.uniform(9.99, 99.99), 2)
                }
            ]
        }

        response = self.client.post(
            "/orders/async",
            json=order_data,
            headers={"Content-Type": "application/json"}
        )

        if response.status_code == 200:
            result = response.json()
            print(f"✓ Customer {self.customer_id} - Order completed in {result.get('processing_time', 0):.2f}s")
