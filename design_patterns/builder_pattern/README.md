# Builder Design Pattern in Go

The **Builder Pattern** is a creational design pattern that allows step-by-step construction of complex objects. It enables you to produce different types and representations of an object using the same construction code.

---

## 📌 Problem & Intent

When creating complex objects with many optional fields or configuration options, constructors can quickly become messy and hard to manage.

### The Naive Approach (Telescoping Constructor Anti-Pattern)
In languages with constructors (or factory functions in Go), you might end up creating functions with long argument lists:

```go
// Telescoping constructor anti-pattern
func NewCar(engineCapacity int, meterConsole string, color string, seats int, hasSunRoof bool, hasGPS bool) *Car {
    // ...
}
```

### Why Naive Approach Fails
- **Unreadable Call Sites**: `NewCar(2000, "Digital", "Red", 4, true, false)` is hard to read and easy to pass parameters in the wrong order.
- **Inflexible Construction**: If only a subset of parameters is needed, callers are forced to pass `nil`, `0`, or default values for unneeded fields.

### The Builder Pattern Solution
Extract object construction code out of its own class and move it to separate objects called **Builders**. Construction is broken down into steps (`SetColor()`, `SetEngineCapacity()`, etc.) executed sequentially via method chaining.

---

## 🏗️ Architecture & Structure

```
                         ┌───────────────────────┐
                         │  CarBuilder Interface │
                         ├───────────────────────┤
                         │ + SetEngineCapacity() │
                         │ + SetMeterConsole()   │
                         │ + SetColor()          │
                         │ + Build()             │
                         └───────────▲───────────┘
                                     │
                                     │ Implements
                         ┌───────────┴───────────┐
                         │   NormalCarBuilder    │
                         ├───────────────────────┤
                         │ - c Car               │
                         └───────────┬───────────┘
                                     │ Builds
                                     ▼
                               ┌───────────┐
                               │    Car    │
                               └───────────┘
```

1. **Product (`Car`)**: The complex object under construction.
2. **Builder Interface (`CarBuilder`)**: Declares step-by-step construction methods shared by all builders.
3. **Concrete Builder (`NormalCarBuilder`)**: Implements the builder interface and maintains the state of the product being built.

---

## 💡 Real-World Use Cases

1. **HTTP Client & Request Configuration**: Configuring headers, timeouts, query params, body, and authentication before executing an HTTP request.
2. **SQL Query Builders**: Dynamically building queries (e.g., `db.Select("name").From("users").Where("id = ?", 1).Build()`).
3. **UI / Document Generation**: Constructing complex HTML, PDF, or Markdown documents step-by-step.
4. **GUI Component Assembly**: Building complex UI dialogs or windows with customizable buttons, toolbars, and layouts.

---

## 💻 Go Implementation Example

The example in this directory demonstrates building a **Car** object using fluent method chaining.

### Project Structure
```text
builder_pattern/
├── go.mod
├── main.go               # Entry point executing the Builder
└── car/
    └── car.go            # Car product, CarBuilder interface & NormalCarBuilder
```

### Code Highlights

#### 1. Builder Interface & Product (`car/car.go`)
```go
package car

type Car struct {
	engineCapacity int
	meterConsole   string
	color          string
}

type CarBuilder interface {
	SetEngineCapacity(capacity int) CarBuilder
	SetMeterConsole(consoleType string) CarBuilder
	SetColor(color string) CarBuilder
	Build() Car
}
```

#### 2. Fluent Method Chaining (`car/car.go`)
```go
func (cb *NormalCarBuilder) SetColor(color string) CarBuilder {
	cb.c.color = color
	return cb // returns builder for method chaining
}

func (cb *NormalCarBuilder) Build() Car {
	return cb.c
}
```

#### 3. Usage (`main.go`)
```go
package main

import (
	"builder_pattern/car"
	"fmt"
)

func main() {
	builder := car.NewNormalCarBuilder()

	myCar := builder.
		SetColor("Red").
		SetEngineCapacity(2000).
		SetMeterConsole("Digital").
		Build()

	fmt.Println(myCar)
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
Car created with color: Red, engine capacity: 2000 and console type: Digital
```

---

## ⚖️ Trade-offs

| Advantages | Disadvantages |
| :--- | :--- |
| **Fluent Interface**: Method chaining makes code highly readable and self-documenting. | **Increased Code Complexity**: Requires creating new builder interfaces and concrete structs. |
| **Immutable Products**: Can construct fully initialized, immutable objects cleanly. | **Duplicate State**: Builder often duplicates fields present in the target object. |
| **Single Responsibility**: Separates construction logic from business logic. | |
