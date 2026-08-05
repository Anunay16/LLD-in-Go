# Concurrency Seat Booking System in Go 🎟️

A production-grade, interview-ready reference implementation in Go for solving high-concurrency seat booking problems.

This project demonstrates how to prevent **double-booking**, eliminate **deadlocks** during multi-seat reservations, ensure **atomic state rollbacks**, and implement a **Two-Phase Reservation Pattern (Hold $\rightarrow$ Confirm)**.

---

## 📖 Table of Contents

1. [Key Concurrency Challenges](#-key-concurrency-challenges)
2. [Architectural Overview & Design Patterns](#-architectural-overview--design-patterns)
3. [Project Directory Structure](#-project-directory-structure)
4. [Step-by-Step Tutorial & Code Walkthrough](#-step-by-step-tutorial--code-walkthrough)
   - [Phase 1: Concurrency Lock Abstraction](#phase-1-concurrency-lock-abstraction)
   - [Phase 2: Domain Entities & State Machine](#phase-2-domain-entities--state-machine)
   - [Phase 3: Thread-Safe Data Repository](#phase-3-thread-safe-data-repository)
   - [Phase 4: Two-Phase Booking Service](#phase-4-two-phase-booking-service)
   - [Phase 5: Extracted Test Scenarios](#phase-5-extracted-test-scenarios)
5. [How to Run Scenarios](#-how-to-run-scenarios)
6. [System Design Interview Deep Dive](#-system-design-interview-deep-dive)

---

## 🚨 Key Concurrency Challenges

In high-concurrency ticketing systems (e.g., BookMyShow, Ticketmaster, airline reservations):

```
                  +-----------------------------------+
                  |   User 1 wants Seats [A1, A2]     |
                  |   User 2 wants Seats [A2, A3]     |
                  +-----------------------------------+
                                    |
          +-------------------------+-------------------------+
          |                                                   |
          v                                                   v
   [Race Condition]                                    [Deadlock Risk]
Two users inspect Seat A2 at                     User 1 locks A1 -> waits for A2
same time -> both see "AVAILABLE"                User 2 locks A2 -> waits for A1
-> both book same seat!                          --> SYSTEM DEADLOCK!
```

1. **Race Conditions / Double Booking**: Multiple incoming requests inspect seat availability concurrently, see `AVAILABLE`, and simultaneously mark the seat as `BOOKED`.
2. **Cyclic Deadlocks**: User 1 tries to reserve `[A1, A2]` (locks `A1` first, requests `A2`). User 2 tries to reserve `[A2, A1]` (locks `A2` first, requests `A1`). Without resource ordering, both block indefinitely.
3. **Partial Allocation Failure**: If a user requests seats `[A1, A2, A3]`, and `A1` & `A2` lock successfully but `A3` fails, `A1` and `A2` **must be released immediately** (atomic rollback) so other users are not blocked.
4. **Temporary Reservation Expiration**: Seats cannot go directly to `BOOKED` state without payment. They must transition to a temporary `HELD` state with a Time-To-Live (TTL) expiration window (e.g., 5 minutes for payment completion).

---

## 🏗️ Architectural Overview & Design Patterns

### 1. Canonical Resource Ordering (Deadlock Prevention)
Before locking multiple resources (`[A2, A1, A3]`), keys are **lexicographically sorted** into a deterministic sequence (`[A1, A2, A3]`). Because all threads lock resources in the exact same order, cyclic wait conditions are mathematically impossible.

### 2. Atomic Multi-Acquire with Instant Rollback
`TryAcquireMulti` iterates through canonically sorted keys. If acquiring any individual lock fails, it automatically releases all previously acquired locks in reverse order before returning an error.

### 3. Decoupled Lock Abstraction (`LockProvider`)
Business logic relies on a generic `LockProvider` interface rather than an in-memory lock directly. This makes it trivial to swap `InMemoryLockProvider` with `RedisLockProvider` or `EtcdLockProvider` for distributed systems.

### 4. Two-Phase Reservation State Machine
```
               +-------------------+
               |     AVAILABLE     |
               +-------------------+
                         |
                         | HoldSeats()
                         v
               +-------------------+
               |       HELD        |
               +-------------------+
              /                     \
ConfirmBooking()                     CancelHold() OR Timeout (Lazy Eviction)
            /                         \
           v                           v
  +-----------------+         +-------------------+
  |     BOOKED      |         |     AVAILABLE     |
  +-----------------+         +-------------------+
```

---

## 📁 Project Directory Structure

```
booking_service_lld/
├── README.md                 # Project Documentation & Tutorial
├── go.mod                    # Go module definition
├── main.go                   # Clean, concise application entry point
├── scenarios_test.go         # Comprehensive test suite covering all scenarios
└── internal/
    ├── lock/
    │   ├── errors.go         # Common lock error definitions
    │   ├── provider.go       # Lock & LockProvider interfaces
    │   ├── in_memory.go      # Deadlock-free In-Memory Lock Provider
    │   └── redis.go          # Distributed Redis Lock Provider
    └── booking/
        ├── entity.go         # Domain models (Seat, Hold, Show) & enums
        ├── repository.go     # Thread-safe repository with defensive deep cloning
        └── service.go        # Business logic for HoldSeats, ConfirmBooking, CancelHold
```

---

## 🛠️ Step-by-Step Tutorial & Code Walkthrough

### Phase 1: Concurrency Lock Abstraction

Defined in [`internal/lock/provider.go`](file:///Users/anunay/Developer/go_coding/booking_service_lld/internal/lock/provider.go) and implemented in [`internal/lock/in_memory.go`](file:///Users/anunay/Developer/go_coding/booking_service_lld/internal/lock/in_memory.go).

```go
type LockProvider interface {
	TryAcquire(ctx context.Context, key string, owner string, ttl time.Duration) (Lock, error)
	Acquire(ctx context.Context, key string, owner string, ttl time.Duration) (Lock, error)
	TryAcquireMulti(ctx context.Context, keys []string, owner string, ttl time.Duration) ([]Lock, error)
	IsLockedBy(ctx context.Context, key string, owner string) bool
}
```

#### Deadlock Prevention via Canonical Sorting:
```go
func (p *InMemoryLockProvider) TryAcquireMulti(ctx context.Context, keys []string, owner string, ttl time.Duration) ([]Lock, error) {
	if len(keys) == 0 {
		return nil, nil
	}

	// 1. CANONICAL SORTING: Sort keys alphabetically to enforce strict global lock order
	sortedKeys := make([]string, len(keys))
	copy(sortedKeys, keys)
	sort.Strings(sortedKeys)
	sortedKeys = deduplicate(sortedKeys)

	acquired := make([]Lock, 0, len(sortedKeys))

	// 2. ATOMIC ACQUIRE & ROLLBACK
	for _, key := range sortedKeys {
		l, err := p.TryAcquire(ctx, key, owner, ttl)
		if err != nil {
			// Rollback acquired locks in reverse order if any key fails
			for i := len(acquired) - 1; i >= 0; i-- {
				_ = acquired[i].Unlock(ctx)
			}
			return nil, fmt.Errorf("%w: failed on resource '%s'", ErrLockNotAcquired, key)
		}
		acquired = append(acquired, l)
	}

	return acquired, nil
}
```

---

### Phase 2: Domain Entities & State Machine

Defined in [`internal/booking/entity.go`](file:///Users/anunay/Developer/go_coding/booking_service_lld/internal/booking/entity.go).

```go
type SeatStatus string

const (
	SeatStatusAvailable SeatStatus = "AVAILABLE"
	SeatStatusHeld      SeatStatus = "HELD"
	SeatStatusBooked   SeatStatus = "BOOKED"
)

type Seat struct {
	ID        string     `json:"id"`
	Number    string     `json:"number"`
	Status    SeatStatus `json:"status"`
	HeldBy    string     `json:"held_by,omitempty"`
	HoldID    string     `json:"hold_id,omitempty"`
	HeldUntil time.Time  `json:"held_until,omitempty"`
}

// IsAvailable performs state check with lazy eviction for expired holds
func (s *Seat) IsAvailable(now time.Time) bool {
	if s.Status == SeatStatusAvailable {
		return true
	}
	if s.Status == SeatStatusHeld && now.After(s.HeldUntil) {
		return true // Expired hold behaves as available
	}
	return false
}
```

---

### Phase 3: Thread-Safe Data Repository

Defined in [`internal/booking/repository.go`](file:///Users/anunay/Developer/go_coding/booking_service_lld/internal/booking/repository.go).

To prevent subtle Go data races where caller goroutines modify returned struct pointers outside mutex locks, `InMemoryRepository` uses **defensive deep copying**:

```go
func (r *InMemoryRepository) GetShow(ctx context.Context, id string) (*Show, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	show, exists := r.shows[id]
	if !exists {
		return nil, ErrShowNotFound
	}
	return cloneShow(show), nil // Returns a deep copy
}
```

---

### Phase 4: Two-Phase Booking Service

Defined in [`internal/booking/service.go`](file:///Users/anunay/Developer/go_coding/booking_service_lld/internal/booking/service.go).

#### 1. Holding Seats (`HoldSeats`)
- Formats lock keys (`show:<showID>:seat:<seatID>`).
- Calls `TryAcquireMulti` (locks seats canonically and atomically).
- Defers unlocking.
- Reads show state, verifies availability.
- Updates seat state to `HELD` with expiration timestamp.
- Saves `Hold` record and updated `Show`.

#### 2. Confirming Booking (`ConfirmBooking`)
- Retrieves `Hold` record using `holdID`.
- Validates user identity and hold active/expiration status.
- Locks seats, verifies seat status matches `holdID`.
- Updates seat status to `BOOKED` and hold status to `CONFIRMED`.

---

### Phase 5: Extracted Test Scenarios

To keep [`main.go`](file:///Users/anunay/Developer/go_coding/booking_service_lld/main.go) decluttered, all test scenarios are extracted into [`scenarios_test.go`](file:///Users/anunay/Developer/go_coding/booking_service_lld/scenarios_test.go):

1. **`TestScenario_HighConcurrencyBooking`**: Simulates Alice `[A1, A2]`, Bob `[A2, A3]`, and Charlie `[A2, A1]` competing concurrently for overlapping seats. Proves deadlock prevention via canonical sorting.
2. **`TestScenario_HoldAndConfirmBooking`**: Happy path of holding seats and confirming booking upon payment.
3. **`TestScenario_HoldExpiration`**: Demonstrates seat hold expiration and subsequent re-booking after TTL.
4. **`TestScenario_CancelHold`**: Early hold cancellation when user abandons reservation.
5. **`TestScenario_ValidationAndErrorHandling`**: Edge cases (non-existent seat/show, hold user mismatch, holding already booked seats).

---

## 🚀 How to Run Scenarios

Make sure you have Go installed (1.20+). You can run all test scenarios or run them 1-by-1:

### Run All Scenarios
```bash
go test -v ./...
```

### Run Scenarios 1 by 1
```bash
# 1. High Concurrency Booking & Deadlock Prevention
go test -v -run TestScenario_HighConcurrencyBooking

# 2. Hold and Confirm Booking Workflow
go test -v -run TestScenario_HoldAndConfirmBooking

# 3. Hold Expiration and Re-Booking
go test -v -run TestScenario_HoldExpiration

# 4. Early Hold Cancellation
go test -v -run TestScenario_CancelHold

# 5. Validation and Error Handling Edge Cases
go test -v -run TestScenario_ValidationAndErrorHandling
```

---

## 💡 System Design Interview Deep Dive

### 1. How to Scale from In-Memory to Distributed Systems?
Replace `InMemoryLockProvider` with `RedisLockProvider` using `SET key owner NX PX ttl`.
- **Lock Key**: `lock:show:<show_id>:seat:<seat_id>`
- **Value**: Random UUID (`owner`)
- **TTL**: Safety fallback (e.g. 10-30 seconds) to prevent abandoned locks if worker crashes.

### 2. Application-Level Lock vs Database Lock

| Strategy | Pros | Cons |
| :--- | :--- | :--- |
| **Distributed/Application Lock (Redis/In-Memory)** | Ultra-fast, offloads DB contention, enables custom hold logic & instant dead-letter rollback | Requires redis cluster management / lock TTL tuning |
| **Pessimistic DB Lock (`SELECT ... FOR UPDATE`)** | Strong transactional guarantees inside RDBMS | DB connection pool bottleneck, high lock contention under traffic spikes |
| **Optimistic DB Lock (`UPDATE seats SET status='HELD', version=version+1 WHERE id=? AND version=?`)** | High throughput for low contention | High retry rate and wasted requests under high contention |

### 3. Expiration Strategy: Active vs Lazy Eviction
- **Lazy Eviction** (Implemented in `Seat.IsAvailable`): Checks `now.After(seat.HeldUntil)` during read operations. Avoids background sweep overhead.
- **Active Expiration** (Cron / Worker): Background worker periodically finds `Hold` records where `status = ACTIVE AND expires_at < NOW()` and releases seat holds to update stats/UI in real-time.
