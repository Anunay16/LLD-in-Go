package repository

import (
	"sync"

	"pharmacy-order-management/models"
)

type CartRepository struct {
	mu    sync.RWMutex
	carts map[string]*models.Cart
}

func NewCartRepository() *CartRepository {
	return &CartRepository{
		carts: make(map[string]*models.Cart),
	}
}

func (r *CartRepository) FindByUserID(userID string) *models.Cart {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.carts[userID]
}

func (r *CartRepository) Save(cart *models.Cart) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.carts[cart.UserID] = cart
}

func (r *CartRepository) Delete(userID string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.carts, userID)
}
