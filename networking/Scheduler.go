package scheduler

import (
	"container/heap"
	"sync"
	"time"
)

type Event struct {
	Time time.Time
}

// have to implement 5 methods for the priority queue
type PriorityQueue []*Event

func (pq PriorityQueue) Len() int            { return len(pq) }
func (pq PriorityQueue) Less(i, j int) bool  { return pq[i].Time.Before(pq[j].Time) }
func (pq PriorityQueue) Swap(i, j int)       { pq[i], pq[j] = pq[j], pq[i] }
func (pq *PriorityQueue) Push(x interface{}) { *pq = append(*pq, x.(*Event)) }
func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	*pq = old[:n-1]
	return item
}

type Scheduler struct {
	mutex  sync.Mutex
	pq     PriorityQueue
	reinit chan struct{}
	kill   chan struct{}

	timer       *time.Timer
	startTime   time.Time
	currentTime time.Time
	increment   time.Time

	// should add an end game channel
}

func CreateScheduler() {
	scheduler := &Scheduler{}
	heap.Init(&scheduler.pq)
	go scheduler.loop()
}

func (scheduler *Scheduler) AddEvent(t time.Time) {

	scheduler.mutex.Lock()
	scheduler.mutex.Unlock()

}

func (scheduler *Scheduler) loop() {

	for {

		scheduler.mutex.Lock()
		if scheduler.pq.Len() > 0 {
			next := scheduler.pq[0]
			timeUntilNext := time.Until(next.Time)

			if scheduler.timer == nil {
				scheduler.timer = time.NewTimer(timeUntilNext)
			} else {
				// drain timer and reset
				if !scheduler.timer.Stop() {
					select {
					case <-scheduler.timer.C:
					default:
					}
				}
				scheduler.timer.Reset(timeUntilNext)
			}
		}
		scheduler.mutex.Unlock()

		select {
		case <-scheduler.timer.C:
			// timer fired; loop will handle due items
		case <-scheduler.reinit:
			// earlier job inserted; loop recalculates
		case <-scheduler.kill:
			return
		}

	}

}
