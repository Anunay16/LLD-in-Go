package services

import (
	"fmt"
	"sync"
	"uber_ride_hailing_lld/internal/models"
)

type UserService struct {
	riders  map[string]*models.Rider
	drivers map[string]*models.Driver
	mu      sync.RWMutex
}

func NewUserService() *UserService {
	return &UserService{
		riders:  make(map[string]*models.Rider),
		drivers: make(map[string]*models.Driver),
	}
}

func (s *UserService) RegisterRider(id, name, phone string, initialLoc models.Location) *models.Rider {
	s.mu.Lock()
	defer s.mu.Unlock()

	rider := models.NewRider(id, name, phone, initialLoc)
	s.riders[id] = rider
	return rider
}

func (s *UserService) RegisterDriver(id, name, phone string, vehicle *models.Vehicle, initialLoc models.Location) *models.Driver {
	s.mu.Lock()
	defer s.mu.Unlock()

	driver := models.NewDriver(id, name, phone, vehicle, initialLoc)
	s.drivers[id] = driver
	return driver
}

func (s *UserService) GetRider(id string) (*models.Rider, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rider, exists := s.riders[id]
	if !exists {
		return nil, fmt.Errorf("rider %s not found", id)
	}
	return rider, nil
}

func (s *UserService) GetDriver(id string) (*models.Driver, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	driver, exists := s.drivers[id]
	if !exists {
		return nil, fmt.Errorf("driver %s not found", id)
	}
	return driver, nil
}

func (s *UserService) GetAvailableDrivers() []*models.Driver {
	s.mu.RLock()
	defer s.mu.RUnlock()

	available := make([]*models.Driver, 0)
	for _, driver := range s.drivers {
		if driver.IsAvailable() {
			available = append(available, driver)
		}
	}
	return available
}
