package eviction

import (
	"github.com/lld/cache/internal/list"
)

// lfuNode holds metadata for an individual key in LFU tracking.
type lfuNode struct {
	key      string
	freq     int
	listNode *list.Node
}

// LFUPolicy implements the Least Frequently Used cache eviction strategy with O(1) operations.
type LFUPolicy struct {
	minFreq    int
	keyToNode  map[string]*lfuNode
	freqToList map[int]*list.DoublyLinkedList
}

// NewLFUPolicy creates a new LFU eviction policy instance.
func NewLFUPolicy() *LFUPolicy {
	return &LFUPolicy{
		minFreq:    0,
		keyToNode:  make(map[string]*lfuNode),
		freqToList: make(map[int]*list.DoublyLinkedList),
	}
}

// KeyAccessed increments the frequency counter for the given key in O(1).
func (l *LFUPolicy) KeyAccessed(key string) {
	node, exists := l.keyToNode[key]
	if !exists {
		return
	}

	oldFreq := node.freq
	oldList := l.freqToList[oldFreq]
	if oldList != nil {
		oldList.Remove(node.listNode)
		if oldList.Len() == 0 {
			delete(l.freqToList, oldFreq)
			if l.minFreq == oldFreq {
				l.minFreq++
			}
		}
	}

	node.freq++
	newList, exists := l.freqToList[node.freq]
	if !exists {
		newList = list.New()
		l.freqToList[node.freq] = newList
	}

	node.listNode = newList.PushFront(key)
}

// KeyAdded registers a new key with frequency 1 and resets minFreq to 1.
func (l *LFUPolicy) KeyAdded(key string) {
	if _, exists := l.keyToNode[key]; exists {
		l.KeyAccessed(key)
		return
	}

	node := &lfuNode{
		key:  key,
		freq: 1,
	}

	freq1List, exists := l.freqToList[1]
	if !exists {
		freq1List = list.New()
		l.freqToList[1] = freq1List
	}

	node.listNode = freq1List.PushFront(key)
	l.keyToNode[key] = node
	l.minFreq = 1
}

// KeyRemoved removes a key from LFU tracking.
func (l *LFUPolicy) KeyRemoved(key string) {
	node, exists := l.keyToNode[key]
	if !exists {
		return
	}

	freqList := l.freqToList[node.freq]
	if freqList != nil {
		freqList.Remove(node.listNode)
		if freqList.Len() == 0 {
			delete(l.freqToList, node.freq)
		}
	}

	delete(l.keyToNode, key)
	if len(l.keyToNode) == 0 {
		l.minFreq = 0
	}
}

// Evict selects and removes the key with the lowest frequency.
func (l *LFUPolicy) Evict() (string, error) {
	if len(l.keyToNode) == 0 {
		return "", ErrEmptyPolicy
	}

	minList := l.freqToList[l.minFreq]
	if minList == nil || minList.Len() == 0 {
		return "", ErrEmptyPolicy
	}

	victimNode := minList.PopBack()
	victimKey := victimNode.Value

	if minList.Len() == 0 {
		delete(l.freqToList, l.minFreq)
	}

	delete(l.keyToNode, victimKey)
	if len(l.keyToNode) == 0 {
		l.minFreq = 0
	}

	return victimKey, nil
}
