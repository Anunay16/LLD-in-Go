# Pharmacy / Medicine Order Management System (LLD in Go)

A clean, modular Low-Level Design (LLD) in Go for an online Pharmacy Order Management System using the **`services` – `repository` – `models`** structure with separate dedicated repositories.

---

## 1. Project Structure

```
inventory-service/
├── cmd/
│   └── main.go                         # Executable demo running complete end-to-end scenarios
├── models/
│   └── models.go                       # Medicine, Cart, Order, and Payment structs
├── repository/                         # Separate, dedicated repository files
│   ├── medicine_repository.go          # In-memory storage for Medicine Catalog (Product Master)
│   ├── inventory_repository.go          # In-memory storage for Real-time Stock Levels
│   ├── cart_repository.go              # In-memory storage for User Carts
│   └── order_repository.go             # In-memory storage for Orders
├── services/                           # Business logic layer
│   ├── inventory_service.go            # Inventory management service
│   ├── cart_service.go                 # Cart operations service
│   ├── payment.go                      # Strategy pattern (UPI, CreditCard, COD)
│   └── order_service.go                # Order placement, checkout, and cancellation
└── README.md
```

---

## 2. End-to-End System Flow

```
[User]
  │
  ├── 1. AddItem(MedicineID, Quantity)
  │      └── CartService validates medicine exists in MedicineRepository
  │      └── CartService verifies available stock in InventoryService
  │      └── Saves Cart Item in CartRepository
  │
  ├── 2. PlaceOrder(PaymentDetails)
  │      └── OrderService calculates total amount from Cart
  │      └── InventoryService atomically checks & deducts stock for all items
  │      └── PaymentStrategy executes payment (UPI / Credit Card / Cash on Delivery)
  │          ├── [SUCCESS]: Order saved with Status PLACED & Cart cleared
  │          └── [FAILURE]: Compensating Rollback restores stock to Inventory
  │
  └── 3. CancelOrder(OrderID)
         └── OrderService verifies order is PLACED
         └── InventoryService restores all ordered quantities back to stock
         └── Updates Order status to CANCELLED
```

### Step-by-Step Flow Breakdown:

1. **Adding Items to Cart (`services/cart_service.go`)**:
   - User adds a medicine with a quantity.
   - `CartService` checks `MedicineRepository` to ensure the medicine exists.
   - `CartService` checks `InventoryService.GetStock()` to ensure stock is available.
   - If stock is sufficient, the item is added to `CartRepository`.

2. **Placing the Order / Checkout (`services/order_service.go`)**:
   - User provides payment details (e.g. UPI, Credit Card, COD).
   - `OrderService` calculates the total price from the cart items.
   - **Atomic Stock Deduction**: `InventoryService.DeductStock()` checks all items atomically and deducts stock under a mutex lock to prevent overselling.
   - **Payment Execution**: `PaymentStrategy.Pay()` processes the transaction.
   - **Success**: If payment succeeds, an `Order` is saved with status `PLACED`, and the cart is cleared.
   - **Compensating Rollback**: If payment fails (e.g. invalid CVV or bank failure), `InventoryService.RestoreStock()` immediately restores all deducted quantities back to inventory with **zero stock leak**.

3. **Order Cancellation (`services/order_service.go`)**:
   - User requests to cancel an order.
   - `OrderService` validates that the order is currently `PLACED`.
   - `InventoryService.RestoreStock()` adds all items back into the inventory stock.
   - Order status is updated to `CANCELLED`.

---

## 3. Code Snapshots of Key Functions

### 3.1 Order Placement & Checkout Orchestration (`services/order_service.go`)

```go
func (s *OrderService) PlaceOrder(userID string, paymentDetails models.PaymentDetails) (*models.Order, error) {
    cart := s.cartService.GetOrCreateCart(userID)
    if len(cart.Items) == 0 {
        return nil, models.ErrCartEmpty
    }

    // 1. Prepare Order Items & Calculate Total
    var orderItems []models.OrderItem
    var totalAmount float64
    for _, item := range cart.Items {
        subtotal := item.Subtotal()
        totalAmount += subtotal
        orderItems = append(orderItems, models.OrderItem{
            MedicineID:   item.Medicine.ID,
            MedicineName: item.Medicine.Name,
            Quantity:     item.Quantity,
            UnitPrice:    item.Medicine.Price,
            TotalPrice:   subtotal,
        })
    }

    // 2. Atomically Check and Deduct Inventory
    if err := s.invService.DeductStock(orderItems); err != nil {
        return nil, err
    }

    // 3. Process Payment via selected strategy
    paymentStrategy, err := GetPaymentStrategy(paymentDetails.Method)
    if err != nil {
        s.invService.RestoreStock(orderItems) // Rollback stock
        return nil, err
    }

    if err := paymentStrategy.Pay(totalAmount, paymentDetails); err != nil {
        s.invService.RestoreStock(orderItems) // Compensate on payment failure
        return nil, fmt.Errorf("payment failed: %w", err)
    }

    // 4. Create and Save Order
    order := &models.Order{
        ID:            fmt.Sprintf("ORD-%d", time.Now().UnixNano()),
        UserID:        userID,
        Items:         orderItems,
        TotalAmount:   totalAmount,
        PaymentMethod: paymentDetails.Method,
        Status:        models.OrderStatusPlaced,
        CreatedAt:     time.Now(),
    }
    s.orderRepo.Save(order)

    // 5. Clear Cart on Success
    s.cartService.ClearCart(userID)

    return order, nil
}
```

---

### 3.2 Concurrency-Safe Stock Deduction & Rollback (`repository/inventory_repository.go`)

```go
func (r *InventoryRepository) DeductStock(items []models.OrderItem) error {
    r.mu.Lock()
    defer r.mu.Unlock()

    // 1. Atomic Check: Verify availability for ALL items first
    for _, item := range items {
        if r.stock[item.MedicineID] < item.Quantity {
            return fmt.Errorf("%w: medicine %s (available: %d, requested: %d)",
                models.ErrInsufficientStock, item.MedicineName, r.stock[item.MedicineID], item.Quantity)
        }
    }

    // 2. Deduct quantities
    for _, item := range items {
        r.stock[item.MedicineID] -= item.Quantity
    }
    return nil
}

func (r *InventoryRepository) RestoreStock(items []models.OrderItem) {
    r.mu.Lock()
    defer r.mu.Unlock()

    for _, item := range items {
        r.stock[item.MedicineID] += item.Quantity
    }
}
```

---

### 3.3 Pluggable Payment Strategies (`services/payment.go`)

```go
type PaymentStrategy interface {
    Pay(amount float64, details models.PaymentDetails) error
}

type UPIPayment struct{}

func (u *UPIPayment) Pay(amount float64, details models.PaymentDetails) error {
    if details.UPIID == "" || !strings.Contains(details.UPIID, "@") {
        return fmt.Errorf("%w: invalid UPI ID", models.ErrPaymentFailed)
    }
    fmt.Printf("[Payment] Paid $%.2f via UPI (%s)\n", amount, details.UPIID)
    return nil
}

type CreditCardPayment struct{}

func (c *CreditCardPayment) Pay(amount float64, details models.PaymentDetails) error {
    if details.CardNumber == "" || len(details.CVV) < 3 {
        return fmt.Errorf("%w: invalid card details", models.ErrPaymentFailed)
    }
    fmt.Printf("[Payment] Paid $%.2f via Credit Card\n", amount)
    return nil
}

type CashOnDeliveryPayment struct{}

func (cod *CashOnDeliveryPayment) Pay(amount float64, details models.PaymentDetails) error {
    fmt.Printf("[Payment] Order placed with Cash on Delivery for $%.2f\n", amount)
    return nil
}
```

---

### 3.4 Adding Items to Cart with Stock Verification (`services/cart_service.go`)

```go
func (s *CartService) AddItem(userID, medicineID string, quantity int) error {
    med, err := s.medicineRepo.FindByID(medicineID)
    if err != nil {
        return err
    }

    // Guard: ensure inventory has enough units before adding to cart
    if s.invService.GetStock(medicineID) < quantity {
        return models.ErrInsufficientStock
    }

    cart := s.GetOrCreateCart(userID)
    cart.Items[medicineID] = models.CartItem{
        Medicine: med,
        Quantity: quantity,
    }
    s.cartRepo.Save(cart)
    return nil
}
```

---

### 3.5 Order Cancellation & Stock Restoration (`services/order_service.go`)

```go
func (s *OrderService) CancelOrder(orderID string) error {
    order, err := s.orderRepo.FindByID(orderID)
    if err != nil {
        return err
    }

    if order.Status != models.OrderStatusPlaced {
        return fmt.Errorf("order is already %s", order.Status)
    }

    // Restore inventory back to shelf
    s.invService.RestoreStock(order.Items)
    order.Status = models.OrderStatusCancelled
    s.orderRepo.Save(order)
    fmt.Printf("[Order] Order %s has been cancelled and stock restored.\n", order.ID)
    return nil
}
```

---

## 4. Why are `MedicineRepository` and `InventoryRepository` Separate?

| Aspect | `MedicineRepository` (Catalog Master) | `InventoryRepository` (Stock Levels) |
| :--- | :--- | :--- |
| **Data Nature** | **Static / Slowly Changing**: Name, Description, Price, Category. | **Dynamic / High-Velocity**: Quantities, Deductions, Restocks. |
| **Access Pattern** | **Read-Heavy (99% Reads)**: Millions of users browse products without mutating data. | **Write-Heavy / Concurrent**: Stock is rapidly decremented/incremented per checkout. |
| **Locking & Caching** | Highly cacheable (CDN, Redis with long TTL). No locks needed for reads. | Requires transactional mutex locks to prevent race conditions and overselling. |
| **Domain Owner** | Product Catalog Team. | Warehouse & Fulfillment Team. |

---

## 5. Repositories Summary (`repository/`)

1. **[`medicine_repository.go`](repository/medicine_repository.go)**: Stores and lists products.
2. **[`inventory_repository.go`](repository/inventory_repository.go)**: Concurrency-safe stock management (`AddStock`, `GetStock`, `DeductStock`, `RestoreStock`).
3. **[`cart_repository.go`](repository/cart_repository.go)**: Stores user carts.
4. **[`order_repository.go`](repository/order_repository.go)**: Stores placed orders.

---

## 6. How to Run the Demo

```bash
go run cmd/main.go
```
