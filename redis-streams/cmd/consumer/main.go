package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"redis-streams-go/internal/shared"
	"github.com/redis/go-redis/v9"
	"strconv"
)

var ctx = context.Background()

func main() {
	consumerName := os.Getenv("CONSUMER_NAME")
	if consumerName == "" {
		consumerName = "default-consumer"
	}

	crashAfter := 0
	if val := os.Getenv("CRASH_AFTER_MESSAGES"); val != "" {
		crashAfter, _ = strconv.Atoi(val)
	}

	rdb := shared.InitRedis(ctx)

	// Ensure Consumer Group exists
	// We do this in both producer and consumer to avoid race conditions on startup.
	err := rdb.XGroupCreateMkStream(ctx, shared.StreamName, shared.GroupName, "0").Err()
	if err != nil {
		// We can ignore the error if the group already exists
		fmt.Printf("[%s] Consumer group setup: %v (Normal if it already exists)\n", consumerName, err)
	}

	fmt.Printf("[%s] Started consumer (Crash after: %d)...\n", consumerName, crashAfter)

	msgCount := 0

	// 1. Recovery Phase: Process pending messages first (ID "0")
	// This ensures that if the consumer crashed, it picks up where it left off.
	fmt.Printf("[%s] Checking for pending messages...\n", consumerName)
	for {
		entries, err := rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    shared.GroupName,
			Consumer: consumerName,
			Streams:  []string{shared.StreamName, "0"}, // "0" means pending messages for THIS consumer
			Count:    1,
			Block:    1 * time.Second,
		}).Result()

		if err != nil {
			if err == redis.Nil {
				break // No more pending messages
			}
			log.Printf("[%s] Error reading pending: %v\n", consumerName, err)
			break
		}

		if len(entries) == 0 || len(entries[0].Messages) == 0 {
			break
		}

		fmt.Printf("[%s] RECOVERING pending message...\n", consumerName)
		processMessages(rdb, consumerName, entries, &msgCount, 0) // Don't crash during recovery for simplicity
	}

	// 2. Normal Processing Phase: Process NEW messages (ID ">")
	fmt.Printf("[%s] Entering normal processing loop...\n", consumerName)
	for {
		entries, err := rdb.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    shared.GroupName,
			Consumer: consumerName,
			Streams:  []string{shared.StreamName, ">"}, // ">" means new messages never delivered to anyone
			Count:    1,
			Block:    0,
		}).Result()

		if err != nil {
			if err == redis.Nil {
				continue
			}
			log.Printf("[%s] Error reading group: %v\n", consumerName, err)
			time.Sleep(1 * time.Second)
			continue
		}

		processMessages(rdb, consumerName, entries, &msgCount, crashAfter)
	}
}

func processMessages(rdb *redis.Client, consumerName string, entries []redis.XStream, msgCount *int, crashAfter int) {
	for _, stream := range entries {
		for _, msg := range stream.Messages {
			fmt.Printf("[%s] Processing msg %s: %v\n", consumerName, msg.ID, msg.Values)

			*msgCount++
			if crashAfter > 0 && *msgCount >= crashAfter {
				fmt.Printf("[%s] !!! SIMULATING CRASH after %d messages !!!\n", consumerName, *msgCount)
				os.Exit(1)
			}

			// Simulate work
			time.Sleep(500 * time.Millisecond)

			err := rdb.XAck(ctx, shared.StreamName, shared.GroupName, msg.ID).Err()
			if err != nil {
				log.Printf("[%s] XACK error: %v\n", consumerName, err)
			} else {
				fmt.Printf("[%s] Acknowledged msg %s\n", consumerName, msg.ID)
			}
		}
	}
}
