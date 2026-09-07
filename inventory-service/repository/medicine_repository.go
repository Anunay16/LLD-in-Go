package repository

import (
	"sync"

	"pharmacy-order-management/models"
)

type MedicineRepository struct {
	mu        sync.RWMutex
	medicines map[string]models.Medicine
}

func NewMedicineRepository() *MedicineRepository {
	return &MedicineRepository{
		medicines: make(map[string]models.Medicine),
	}
}

func (r *MedicineRepository) Save(med models.Medicine) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.medicines[med.ID] = med
}

func (r *MedicineRepository) FindByID(id string) (models.Medicine, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	med, exists := r.medicines[id]
	if !exists {
		return models.Medicine{}, models.ErrMedicineNotFound
	}
	return med, nil
}

func (r *MedicineRepository) List() []models.Medicine {
	r.mu.RLock()
	defer r.mu.RUnlock()
	list := make([]models.Medicine, 0, len(r.medicines))
	for _, med := range r.medicines {
		list = append(list, med)
	}
	return list
}
