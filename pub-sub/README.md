# Kafka-like Pub-Sub System LLD in Golang

This project implements a Low-Level Design (LLD) of a distributed Publish-Subscribe messaging system similar to Apache Kafka, written in Golang. 

## 1. Functional Requirements

*   **Topics:** The system must allow the creation of topics, which represent logical streams of data.
*   **Partitions:** Topics must be divided into partitions to allow for scalability and parallel processing. Partitions maintain a strict ordering of messages.
*   **Producers:** Applications must be able to publish messages to a specific topic.
*   **Message Routing (Partitioning):**
    *   If a message has a key, all messages with the same key must be routed to the same partition (Hash-based partitioning) to guarantee ordering for that key.
    *   If a message has no key, the system should distribute it (e.g., via Round-Robin or Timestamp-based routing) across available partitions for load balancing.
*   **Consumers and Consumer Groups:**
    *   Consumers should be able to subscribe to a topic and read messages.
    *   Consumers should be organized into Consumer Groups.
    *   Each partition in a topic must be assigned to exactly one consumer within a Consumer Group. (This prevents duplicate processing within the group and ensures ordered consumption per partition).
*   **Offset Management:** The system must keep track of the offsets (the position of the message in a partition) that a consumer group has successfully processed, allowing consumers to resume from where they left off in case of failures.

## 2. Non-Functional Requirements

*   **Scalability:** The system should be able to handle an increasing volume of messages by adding more partitions to a topic and more consumers to a group.
*   **High Throughput:** The design should support high message ingestion and consumption rates (achieved via concurrent processing on partitions).
*   **Availability/Fault Tolerance:** (In this LLD, simulated via concurrent access safety) The system should allow multiple concurrent reads and writes without corrupting the state. In a full system, this implies data replication across multiple brokers.
*   **Durability:** Messages appended to a partition should be persistent. (In this in-memory LLD, it's simulated by retaining messages in memory slices; a real system would write to disk).
*   **Concurrency:** The system must be thread-safe. Multiple producers and consumers will be accessing the broker, topics, and partitions simultaneously.

## 3. SOLID Principles & Architecture

The architecture relies heavily on interfaces defined in `core/interfaces.go` to enforce SOLID principles, specifically **Dependency Inversion** and **Open/Closed**:

### Dependency Inversion Principle (DIP)
*   High-level modules (Clients, Producers, ConsumerGroups) **do not depend on concrete implementations** of the Broker, Topic, or Partition.
*   Instead, they depend on abstractions (`core.Broker`, `core.Topic`, `core.Partition`). 
*   This makes the system highly testable (e.g., we can easily inject a mock Broker into the Producer for unit testing).

### Open/Closed Principle (OCP)
*   The `core.Partitioner` interface allows you to define new routing strategies (e.g., `RoundRobinPartitioner`, `ConsistentHashPartitioner`) without ever modifying the `broker.DefaultTopic` code. You simply inject a different `Partitioner` into the topic.
*   The `core.Partition` interface allows extending the system with a `DiskPartition` instead of a `MemoryPartition` without changing how Topics or Consumers interact with it.

## 4. Core Components

*   `core.Broker`, `core.Topic`, `core.Partition`, `core.Partitioner`: The abstract interfaces defining the contract of the system.
*   `models.Message`: Encapsulates the data being transmitted (Key, Value, Offset, Timestamp).
*   `broker.MemoryPartition`: The fundamental unit of storage. An append-only sequence of messages.
*   `broker.DefaultTopic`: A collection of partitions. Handles routing logic by delegating to a `Partitioner`.
*   `broker.MemoryBroker`: The central registry that manages topics.
*   `client.Producer`: Provides the API for applications to send messages to the broker.
*   `client.ConsumerGroup`: Manages consumer instances, assigns partitions to them, and tracks committed offsets.

## 5. Design Patterns Used

### 5.1. Strategy Pattern (Partitioning Strategy)
*   **Where:** `core.Partitioner` and `DefaultTopic.Publish()`
*   **Why:** We need different algorithms to decide which partition a message goes to, and we want to change them dynamically or extend them without modifying the Topic class (OCP).

### 5.2. Singleton Pattern
*   **Where:** `broker.GetBroker()`
*   **Why:** The broker acts as the central coordinator. We only want one instance of the Broker to manage the state of all topics and partitions to avoid inconsistencies.

### 5.3. Observer Pattern (Pub-Sub variant)
*   **Where:** `ConsumerGroup.consumePartition()` (polling loop)
*   **Why:** Consumers need to be notified of or retrieve new messages as they arrive in a partition.
*   **Implementation:** Kafka uses a "pull" model. The consumer continuously polls (`Read(offset)`) the partition. If a message is found, it "observes" it; otherwise, it yields/sleeps.

### 5.4. Concurrency Patterns (Worker Pool / Goroutines)
*   **Where:** `ConsumerGroup.Subscribe()`
*   **Why:** To achieve high throughput, partitions must be processed in parallel.
*   **Implementation:** The Consumer Group launches a distinct Goroutine (worker) for each partition it is assigned to, coupled with `sync.RWMutex` for memory safety.

## How to Run

1. Make sure you have Go installed.
2. Clone/navigate to the directory.
3. Run the demonstration:
   ```bash
   go run main.go
   ```
