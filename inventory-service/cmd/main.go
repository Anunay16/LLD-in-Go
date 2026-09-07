package main

import (
	"fmt"

	"pharmacy-order-management/models"
	"pharmacy-order-management/repository"
	"pharmacy-order-management/services"
)

func main() {
	fmt.Println("==================================================================")
	fmt.Println("       PHARMACY / MEDICINE ORDER MANAGEMENT SYSTEM")
	fmt.Println("==================================================================")

	// 1. Initialize Individual Repositories
	medRepo := repository.NewMedicineRepository()
	cartRepo := repository.NewCartRepository()
	orderRepo := repository.NewOrderRepository()
	invRepo := repository.NewInventoryRepository()

	// 2. Initialize Services (injecting repositories)
	invService := services.NewInventoryService(invRepo)
	cartService := services.NewCartService(cartRepo, medRepo, invService)
	orderService := services.NewOrderService(orderRepo, cartService, invService)

	// 3. Setup Medicine Catalog and Initial Stock
	med1 := models.Medicine{ID: "M1", Name: "Paracetamol 500mg", Price: 5.0}
	med2 := models.Medicine{ID: "M2", Name: "Amoxicillin 250mg", Price: 15.0}
	med3 := models.Medicine{ID: "M3", Name: "Vitamin C 1000mg", Price: 8.0}

	medRepo.Save(med1)
	medRepo.Save(med2)
	medRepo.Save(med3)

	invService.AddStock(med1.ID, 50) // 50 Paracetamol
	invService.AddStock(med2.ID, 10) // 10 Amoxicillin
	invService.AddStock(med3.ID, 20) // 20 Vitamin C

	fmt.Println("\n[1] Medicine Catalog & Stock Loaded:")
	for _, m := range medRepo.List() {
		fmt.Printf("    • %-20s | Price: $%-5.2f | Stock: %d\n", m.Name, m.Price, invService.GetStock(m.ID))
	}

	// 4. User Alice adds medicines to Cart
	userID := "alice"
	fmt.Printf("\n[2] User '%s' adds items to cart:\n", userID)
	_ = cartService.AddItem(userID, med1.ID, 2) // 2 x $5 = $10
	_ = cartService.AddItem(userID, med2.ID, 1) // 1 x $15 = $15

	cart := cartService.GetOrCreateCart(userID)
	for _, item := range cart.Items {
		fmt.Printf("    • %-20s x %d = $%.2f\n", item.Medicine.Name, item.Quantity, item.Subtotal())
	}

	// 5. Place Order with UPI Payment
	fmt.Println("\n[3] Placing Order via UPI...")
	paymentDetails := models.PaymentDetails{
		Method: models.PaymentMethodUPI,
		UPIID:  "alice@bank",
	}

	order, err := orderService.PlaceOrder(userID, paymentDetails)
	if err != nil {
		fmt.Printf("    ❌ Order failed: %v\n", err)
	} else {
		fmt.Printf("    ✔ Order Placed Successfully! OrderID: %s | Total: $%.2f | Status: %s\n",
			order.ID, order.TotalAmount, order.Status)
	}

	fmt.Println("\n[4] Updated Stock after Alice's Order:")
	fmt.Printf("    • Paracetamol Remaining: %d (was 50)\n", invService.GetStock(med1.ID))
	fmt.Printf("    • Amoxicillin Remaining: %d (was 10)\n", invService.GetStock(med2.ID))

	// 6. Payment Failure Simulation (Compensating Stock Rollback)
	userBob := "bob"
	fmt.Printf("\n[5] User '%s' tries to place order with invalid payment details:\n", userBob)
	_ = cartService.AddItem(userBob, med2.ID, 2)

	stockBefore := invService.GetStock(med2.ID)
	fmt.Printf("    • Amoxicillin stock before Bob's order attempt: %d\n", stockBefore)

	_, failErr := orderService.PlaceOrder(userBob, models.PaymentDetails{
		Method: models.PaymentMethodCreditCard,
		CVV:    "1", // Invalid CVV
	})

	if failErr != nil {
		fmt.Printf("    ✔ Expected Error Caught: %v\n", failErr)
		fmt.Printf("    ✔ Stock restored automatically: Amoxicillin stock is still %d\n", invService.GetStock(med2.ID))
	}

	// 7. Insufficient Stock Scenario
	userCharlie := "charlie"
	fmt.Printf("\n[6] User '%s' tries to order 20 units of Amoxicillin (Only 9 available):\n", userCharlie)
	errStock := cartService.AddItem(userCharlie, med2.ID, 20)
	if errStock != nil {
		fmt.Printf("    ✔ Stock Guard: %v\n", errStock)
	}

	// 8. Order Cancellation Demo
	fmt.Println("\n[7] Cancelling Alice's Order:")
	_ = orderService.CancelOrder(order.ID)
	fmt.Printf("    • Paracetamol Stock after cancellation: %d (Restored)\n", invService.GetStock(med1.ID))
	fmt.Printf("    • Amoxicillin Stock after cancellation: %d (Restored)\n", invService.GetStock(med2.ID))

	fmt.Println("\n==================================================================")
	fmt.Println("             ALL SIMPLE SCENARIOS COMPLETED")
	fmt.Println("==================================================================")
}
