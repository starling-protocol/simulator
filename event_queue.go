package simulator

import "time"

// An EventQueue implements heap.Interface and holds Items.
type EventQueue []Event

func (pq EventQueue) Len() int { return len(pq) }

func (pq EventQueue) Less(i, j int) bool {
	if pq[i].Time() == pq[j].Time() {
		return int(pq[i].SequenceNumber()) < int(pq[j].SequenceNumber())
	} else {
		return pq[i].Time() < pq[j].Time()
	}
}

func (pq EventQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}

func (pq *EventQueue) Push(x any) {
	item := x
	*pq = append(*pq, item.(Event))
}

func (pq *EventQueue) Pop() any {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil // avoid memory leak
	*pq = old[0 : n-1]
	return item
}

func (pq *EventQueue) Next() any {
	tmp := *pq
	item := tmp[0]
	return item
}

func (pq *EventQueue) ShouldPop(currentTime time.Duration) bool {
	tmp := *pq
	return tmp[0].Time() <= currentTime
}
