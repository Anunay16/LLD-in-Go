package storage

// Storage defines the interface for key-value data storage.
type Storage interface {
	Get(key string) (any, bool)
	Set(key string, value any)
	Delete(key string) bool
	Len() int
	Clear()
}

// MemoryStorage is an in-memory hashmap implementation of Storage.
type MemoryStorage struct {
	data map[string]any
}

// NewMemoryStorage creates a new MemoryStorage instance.
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		data: make(map[string]any),
	}
}

func (m *MemoryStorage) Get(key string) (any, bool) {
	val, exists := m.data[key]
	return val, exists
}

func (m *MemoryStorage) Set(key string, value any) {
	m.data[key] = value
}

func (m *MemoryStorage) Delete(key string) bool {
	if _, exists := m.data[key]; exists {
		delete(m.data, key)
		return true
	}
	return false
}

func (m *MemoryStorage) Len() int {
	return len(m.data)
}

func (m *MemoryStorage) Clear() {
	m.data = make(map[string]any)
}
