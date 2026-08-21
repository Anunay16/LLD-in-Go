# Singleton Design Pattern in Go

The **Singleton Pattern** is a creational design pattern that ensures a struct/class has only **one instance** throughout the application lifetime and provides a **global point of access** to that instance.

---

## 📌 Problem & Intent

Certain shared resources in an application—such as database connections, loggers, or configuration settings—should only have a single instance. Creating multiple instances can lead to resource exhaustion, inconsistent state, or unexpected race conditions.

### The Singleton Solution in Go
In Go, the idiomatic and thread-safe way to implement a Singleton is using the standard library's `sync.Once`.

`sync.Once` guarantees that a function execution occurs **exactly once**, even across multiple goroutines executing concurrently.

---

## 🏗️ Architecture & Structure

```
                             ┌───────────────────┐
                             │ Database (Interface)│
                             ├───────────────────┤
                             │ + GetUserName()   │
                             └─────────▲─────────┘
                                       │
                                       │ Implements
                       ┌───────────────┴───────────────┐
                       │   singletonDatabase (Private) │
                       ├───────────────────────────────┤
                       │ - db map[int]string           │
                       └───────────────▲───────────────┘
                                       │
                        Returned by    │ Thread-safe initialization
            GetSingletonDatabaseInstance() via sync.Once
```

1. **Database Interface**: Exposes operations without revealing the underlying singleton implementation.
2. **`singletonDatabase` (Unexported Struct)**: Encapsulates state and data. Kept private to prevent direct external instantiation.
3. **`sync.Once`**: Guarantees thread-safe initialization on the first call to `GetSingletonDatabaseInstance()`.

---

## 💡 Real-World Use Cases

1. **Database Connection Pools**: Sharing a single thread-safe connection pool across web server request handlers.
2. **Application Logger**: Writing logs from multiple goroutines to a single centralized log handler/stream.
3. **Configuration Manager**: Loading application settings (`config.yaml` / environment variables) once at boot time and sharing read access globally.
4. **Hardware Interface / Device Drivers**: Managing access to a shared physical hardware resource (e.g., printer, serial port).

---

## 💻 Go Implementation Example

The example in this directory models a **Thread-Safe Singleton Database Connection**.

### Project Structure
```text
singleton_pattern/
├── go.mod
├── main.go                       # Entry point accessing Singleton instance
└── database/
    └── singleton_database.go     # Singleton Database implementation using sync.Once
```

### Code Highlights

#### 1. Thread-Safe Initialization with `sync.Once` (`database/singleton_database.go`)
```go
package database

import "sync"

type Database interface {
	GetUserName(id int) string
}

type singletonDatabase struct {
	db map[int]string
}

var (
	once     sync.Once
	instance Database
)

// GetSingletonDatabaseInstance provides a thread-safe global access point.
func GetSingletonDatabaseInstance() Database {
	once.Do(func() {
		instance = &singletonDatabase{
			db: map[int]string{
				1: "alice",
				2: "bob",
			},
		}
	})
	return instance
}

func (s *singletonDatabase) GetUserName(id int) string {
	return s.db[id]
}
```

#### 2. Usage (`main.go`)
```go
package main

import (
	"fmt"
	"singleton_pattern/database"
)

func main() {
	dbInstance := database.GetSingletonDatabaseInstance()
	name := dbInstance.GetUserName(1)
	fmt.Print(name)
}
```

---

## 🚀 How to Run

Navigate to this directory and run:

```bash
go run main.go
```

### Output
```text
alice
```

---

## ⚖️ Trade-offs

| Advantages | Disadvantages |
| :--- | :--- |
| **Controlled Access**: Guarantees only one instance exists across the application. | **Global State**: Can introduce hidden dependencies and make code harder to reason about. |
| **Lazy Initialization**: Resource is allocated only when accessed for the first time. | **Testing Challenges**: Unit testing can be harder because state persists between test cases if not carefully reset or mocked via interfaces. |
| **Thread Safety**: Handled cleanly in Go using `sync.Once`. | |
