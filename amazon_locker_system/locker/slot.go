package locker

import (
	"strconv"
	"sync/atomic"
)

type SlotSize string

const (
	LargeSlotSize  SlotSize = "large"
	MediumSlotSize SlotSize = "medium"
	SmallSlotSize  SlotSize = "small"
)

type Slot struct {
	id        int64
	available atomic.Bool
	size      SlotSize
}

var slotIDCounter atomic.Int64

func NewSlot(size SlotSize) *Slot {
	id := slotIDCounter.Add(1)
	slot := &Slot{
		id:   id,
		size: size,
	}
	slot.available.Store(true)
	return slot
}

func (ss *Slot) GetSize() SlotSize {
	return ss.size
}

func (ss *Slot) Acquire() bool {
	return ss.available.CompareAndSwap(true, false)
}

func (ss *Slot) IsAvailable() bool {
	return ss.available.Load()
}

func (ss *Slot) GetID() string {
	return strconv.FormatInt(ss.id, 10)
}
