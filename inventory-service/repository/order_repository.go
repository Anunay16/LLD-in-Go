package repository

import (
	"sync"

	"pharmacy-order-management/models"
)

type OrderRepository struct {
	mu     sync.RWMutex
	orders map[string]*models.Order
}

func NewOrderRepository() *OrderRepository {
	return &OrderRepository{
		orders: make(map[string]*models.Order),
	}
}

func (r *OrderRepository) Save(order *models.Order) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.orders[order.ID] = order
}

func (r *OrderRepository) FindByID(id string) (*models.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	order, exists := r.orders[id]
	if !exists {
		return nil, models.ErrOrderNotFound
	}
	return order, nil
}
