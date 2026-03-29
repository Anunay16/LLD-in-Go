package notificationservice

import (
	"container/heap"
	"sync"
	"time"
)

type Task struct {
	executeAt   time.Time
	executeFunc func()
	index       int
}

type PriorityQueue []*Task

func (pq PriorityQueue) Len() int {
	return len(pq)
}

func (pq PriorityQueue) Less(i, j int) bool {
	return pq[i].executeAt.Before(pq[j].executeAt)
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *PriorityQueue) Push(x interface{}) {
	task := x.(*Task)
	*pq = append(*pq, task)
}

func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	task := old[n-1]
	*pq = old[0 : n-1]
	return task
}

type Scheduler struct {
	mu   sync.Mutex
	pq   PriorityQueue
	wake chan struct{}
	stop chan struct{}
}

func NewScheduler() *Scheduler {
	s := &Scheduler{
		pq:   make(PriorityQueue, 0),
		wake: make(chan struct{}, 1),
		stop: make(chan struct{}),
	}
	heap.Init(&s.pq)
	go s.run()
	return s
}

func (s *Scheduler) Schedule(executeAt time.Time, f func()) {
	s.mu.Lock()
	defer s.mu.Unlock()

	task := &Task{
		executeAt:   executeAt,
		executeFunc: f,
	}

	heap.Push(&s.pq, task)

	// Wake up worker if needed
	select {
	case s.wake <- struct{}{}:
	default:
	}
}

func (s *Scheduler) run() {
	/*
	   LOOP:
	     lock
	     if no work → unlock + wait
	     peek next task
	     compute wait time
	     unlock
	     wait OR get interrupted
	     lock again
	     re-check state
	     execute task

	   	“Worker sleeps until there’s work or time arrives.
	   	Before sleeping, it leaves the key (lock).
	   	When it wakes up, it checks if work still exists.”
	*/
	for {
		s.mu.Lock()

		if len(s.pq) == 0 {
			// no task is pending, go to sleep, wake up if any new task is added
			s.mu.Unlock()
			select {
			case <-s.wake:
				continue
			case <-s.stop:
				return
			}
		}

		nextTask := s.pq[0]
		now := time.Now()
		wait := nextTask.executeAt.Sub(now)

		// Never hold a lock while waiting
		s.mu.Unlock()

		if wait > 0 {
			// check if any other task with lower executeAt comes or
			// execute if the wait time becomes 0
			select {
			case <-time.After(wait):
			case <-s.wake:
				continue
			case <-s.stop:
				return
			}
		}

		s.mu.Lock()
		if len(s.pq) == 0 {
			s.mu.Unlock()
			continue
		}

		task := heap.Pop(&s.pq).(*Task)
		s.mu.Unlock()

		// Execute in separate goroutine (non-blocking)
		go task.executeFunc()
	}
}
