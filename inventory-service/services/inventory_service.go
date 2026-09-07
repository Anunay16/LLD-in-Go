package services

import (
	"pharmacy-order-management/models"
	"pharmacy-order-management/repository"
)

type InventoryService struct {
	invRepo *repository.InventoryRepository
}

func NewInventoryService(invRepo *repository.InventoryRepository) *InventoryService {
	return &InventoryService{
		invRepo: invRepo,
	}
}

func (s *InventoryService) AddStock(medicineID string, quantity int) {
	s.invRepo.AddStock(medicineID, quantity)
}

func (s *InventoryService) GetStock(medicineID string) int {
	return s.invRepo.GetStock(medicineID)
}

func (s *InventoryService) DeductStock(items []models.OrderItem) error {
	return s.invRepo.DeductStock(items)
}

func (s *InventoryService) RestoreStock(items []models.OrderItem) {
	s.invRepo.RestoreStock(items)
}
