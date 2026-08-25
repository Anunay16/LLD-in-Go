package main

import (
	"fmt"

	"github.com/lld/cache/core"
	"github.com/lld/cache/eviction"
)

func main() {
	fmt.Println("=== 1. LRU Cache Demo ===")
	demoLRU()

	fmt.Println("\n=== 2. LFU Cache Demo ===")
	demoLFU()
}

func demoLRU() {
	// Capacity = 3, Eviction Strategy = LRU
	lruPolicy := eviction.NewLRUPolicy()
	cache := core.NewCache(3, lruPolicy, nil)

	cache.Put("A", 100)
	cache.Put("B", 200)
	cache.Put("C", 300)

	// Access A -> Recency Order: A (MRU), C, B (LRU)
	cache.Get("A")

	// Put D -> Evicts B (least recently used)
	cache.Put("D", 400)

	_, foundB := cache.Get("B")
	valA, foundA := cache.Get("A")
	valD, foundD := cache.Get("D")

	fmt.Printf("Get 'B' (evicted): found=%v\n", foundB)
	fmt.Printf("Get 'A': val=%v, found=%v\n", valA, foundA)
	fmt.Printf("Get 'D': val=%v, found=%v\n", valD, foundD)
	fmt.Printf("Current Cache Size: %d / %d\n", cache.Len(), cache.Capacity())
}

func demoLFU() {
	// Capacity = 3, Eviction Strategy = LFU
	lfuPolicy := eviction.NewLFUPolicy()
	cache := core.NewCache(3, lfuPolicy, nil)

	cache.Put("apple", "fruit")
	cache.Put("carrot", "veg")
	cache.Put("banana", "fruit")

	// Access "apple" 3 times (freq = 4)
	cache.Get("apple")
	cache.Get("apple")
	cache.Get("apple")

	// Access "carrot" 1 time (freq = 2)
	cache.Get("carrot")

	// "banana" was only inserted once (freq = 1)

	// Put "date" -> Evicts "banana" (least frequently used)
	cache.Put("date", "fruit")

	_, foundBanana := cache.Get("banana")
	valApple, foundApple := cache.Get("apple")
	valDate, foundDate := cache.Get("date")

	fmt.Printf("Get 'banana' (evicted): found=%v\n", foundBanana)
	fmt.Printf("Get 'apple': val=%v, found=%v\n", valApple, foundApple)
	fmt.Printf("Get 'date': val=%v, found=%v\n", valDate, foundDate)
	fmt.Printf("Current Cache Size: %d / %d\n", cache.Len(), cache.Capacity())
}
