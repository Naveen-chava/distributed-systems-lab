package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"redis-streams-go/internal/shared"
	"github.com/redis/go-redis/v9"
)

var ctx = context.Background()

func main() {
	rdb := shared.InitRedis(ctx)

	// Create Consumer Group
	err := rdb.XGroupCreateMkStream(ctx, shared.StreamName, shared.GroupName, "0").Err()
	if err != nil {
		fmt.Printf("Consumer group setup: %v (Normal if it already exists)\n", err)
	}

	for i := 1; i <= 20; i++ {
		message := map[string]interface{}{
			"event_id":   i,
			"event_type": "user_signup",
			"timestamp":  time.Now().Unix(),
		}

		err := rdb.XAdd(ctx, &redis.XAddArgs{
			Stream: shared.StreamName,
			Values: message,
		}).Err()

		if err != nil {
			log.Printf("Producer error: %v\n", err)
		} else {
			fmt.Printf("[PRODUCER] Sent msg %d\n", i)
		}

		time.Sleep(1 * time.Second)
	}
	fmt.Println("[PRODUCER] Finished sending 20 messages. Exiting.")
}
