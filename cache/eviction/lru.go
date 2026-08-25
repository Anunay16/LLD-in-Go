package eviction

import (
	"github.com/lld/cache/internal/list"
)

// LRUPolicy implements the Least Recently Used cache eviction strategy.
type LRUPolicy struct {
	dll   *list.DoublyLinkedList
	nodes map[string]*list.Node
}

// NewLRUPolicy creates a new LRU eviction policy instance.
func NewLRUPolicy() *LRUPolicy {
	return &LRUPolicy{
		dll:   list.New(),
		nodes: make(map[string]*list.Node),
	}
}

// KeyAccessed marks key as most recently used by moving it to the front.
func (l *LRUPolicy) KeyAccessed(key string) {
	if node, exists := l.nodes[key]; exists {
		l.dll.MoveToFront(node)
	}
}

// KeyAdded registers a new key as the most recently used element.
func (l *LRUPolicy) KeyAdded(key string) {
	if node, exists := l.nodes[key]; exists {
		l.dll.MoveToFront(node)
		return
	}
	node := l.dll.PushFront(key)
	l.nodes[key] = node
}

// KeyRemoved removes a key from LRU tracking.
func (l *LRUPolicy) KeyRemoved(key string) {
	if node, exists := l.nodes[key]; exists {
		l.dll.Remove(node)
		delete(l.nodes, key)
	}
}

// Evict finds and removes the least recently used key (from the back).
func (l *LRUPolicy) Evict() (string, error) {
	if l.dll.Len() == 0 {
		return "", ErrEmptyPolicy
	}

	node := l.dll.PopBack()
	delete(l.nodes, node.Value)
	return node.Value, nil
}
