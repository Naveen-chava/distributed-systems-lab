# Redis Streams Distributed Example (Go)

This is a simple demonstration of **Redis Streams** using Go and Docker. it showcases how to build a distributed system where multiple consumers cooperatively process a stream of messages with built-in fault tolerance.

## Architecture

- **`internal/shared`**: Shared configuration and Redis initialization.
- **`cmd/producer`**: A standalone service that generates event data and pushes it to the Redis Stream.
- **`cmd/consumer`**: A standalone service that processes events from a **Consumer Group**.
- **Redis**: The message broker providing the stream infrastructure.

## Key Redis Streams Concepts

### 1. Consumer Groups
The `docker-compose.yml` spins up two independent consumers (`consumer-1` and `consumer-2`) belonging to the same group. Redis automatically load-balances the messages between them.

### 2. Acknowledgment (`XACK`)
Messages are only removed from a consumer's "pending" list once they are explicitly acknowledged. If a consumer crashes before acknowledgment, the message is preserved.

### 3. Recovery
The consumer is built with a **Two-Phase Startup**:
1.  **Recovery Phase**: On startup, it first checks its **Pending Entries List (PEL)** using the Special ID `0`. It processes any messages that were left unacknowledged from a previous run (e.g., after a crash).
2.  **Normal Phase**: Once the PEL is empty, it starts listening for **New** messages using the Special ID `>`.


### Build and Start
```bash
docker compose up --build
```

### Observe the Flow
- You will see the **Producer** sending 20 messages.
- **Consumer-1** and **Consumer-2** will alternate processing the messages.
- We have configured **Consumer-1** to simulate a crash after every 5 messages (`CRASH_AFTER_MESSAGES=5`).
- You will see Docker restart `consumer-1`, which will then proactively **RECOVER** the message it was working on before the crash.
