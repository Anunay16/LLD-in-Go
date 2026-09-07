package services

import (
	"fmt"
	"time"

	"pharmacy-order-management/models"
	"pharmacy-order-management/repository"
)

type OrderService struct {
	orderRepo   *repository.OrderRepository
	cartService *CartService
	invService  *InventoryService
}

func NewOrderService(orderRepo *repository.OrderRepository, cartService *CartService, invService *InventoryService) *OrderService {
	return &OrderService{
		orderRepo:   orderRepo,
		cartService: cartService,
		invService:  invService,
	}
}

// PlaceOrder converts the user's cart into an order, checks & deducts inventory, and processes payment
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
		s.invService.RestoreStock(orderItems) // Compensate / Rollback stock
		return nil, err
	}

	if err := paymentStrategy.Pay(totalAmount, paymentDetails); err != nil {
		s.invService.RestoreStock(orderItems) // Compensate / Rollback stock
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

func (s *OrderService) CancelOrder(orderID string) error {
	order, err := s.orderRepo.FindByID(orderID)
	if err != nil {
		return err
	}

	if order.Status != models.OrderStatusPlaced {
		return fmt.Errorf("order is already %s", order.Status)
	}

	// Restore inventory
	s.invService.RestoreStock(order.Items)
	order.Status = models.OrderStatusCancelled
	s.orderRepo.Save(order)
	fmt.Printf("[Order] Order %s has been cancelled and stock restored.\n", order.ID)
	return nil
}
