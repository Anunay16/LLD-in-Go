package repository

import (
	"fmt"
	"sync"

	"pharmacy-order-management/models"
)

type InventoryRepository struct {
	mu    sync.RWMutex
	stock map[string]int // MedicineID -> Quantity
}

func NewInventoryRepository() *InventoryRepository {
	return &InventoryRepository{
		stock: make(map[string]int),
	}
}

func (r *InventoryRepository) AddStock(medicineID string, quantity int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.stock[medicineID] += quantity
}

func (r *InventoryRepository) GetStock(medicineID string) int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.stock[medicineID]
}

func (r *InventoryRepository) DeductStock(items []models.OrderItem) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, item := range items {
		if r.stock[item.MedicineID] < item.Quantity {
			return fmt.Errorf("%w: medicine %s (available: %d, requested: %d)",
				models.ErrInsufficientStock, item.MedicineName, r.stock[item.MedicineID], item.Quantity)
		}
	}

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
