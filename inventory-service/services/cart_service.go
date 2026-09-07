package services

import (
	"pharmacy-order-management/models"
	"pharmacy-order-management/repository"
)

type CartService struct {
	cartRepo     *repository.CartRepository
	medicineRepo *repository.MedicineRepository
	invService   *InventoryService
}

func NewCartService(cartRepo *repository.CartRepository, medicineRepo *repository.MedicineRepository, invService *InventoryService) *CartService {
	return &CartService{
		cartRepo:     cartRepo,
		medicineRepo: medicineRepo,
		invService:   invService,
	}
}

func (s *CartService) GetOrCreateCart(userID string) *models.Cart {
	cart := s.cartRepo.FindByUserID(userID)
	if cart == nil {
		cart = models.NewCart(userID)
		s.cartRepo.Save(cart)
	}
	return cart
}

func (s *CartService) AddItem(userID, medicineID string, quantity int) error {
	med, err := s.medicineRepo.FindByID(medicineID)
	if err != nil {
		return err
	}

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

func (s *CartService) RemoveItem(userID, medicineID string) {
	cart := s.GetOrCreateCart(userID)
	delete(cart.Items, medicineID)
	s.cartRepo.Save(cart)
}

func (s *CartService) ClearCart(userID string) {
	cart := s.GetOrCreateCart(userID)
	cart.Items = make(map[string]models.CartItem)
	s.cartRepo.Save(cart)
}
