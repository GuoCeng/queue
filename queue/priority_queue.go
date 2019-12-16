package queue

import (
	"container/heap"
	"sync"
)

// An Item is something we manage in a priority queue.
type Item struct {
	Value    interface{} // The value of the item; arbitrary.
	Priority int64       // The priority of the item in the queue.
	index    int         // The index of the item in the heap. The index is needed by update and is maintained by the heap.Interface methods.
}

func (i *Item) Index() int {
	return i.index
}

// A priorityQueue implements heap.Interface and holds Items.
type priorityQueue struct {
	items    []*Item
	newItems []*Item
	delItem  *Item
}

func (pq priorityQueue) Len() int { return len(pq.items) }

func (pq priorityQueue) Less(i, j int) bool {
	// We want Pop to give us the highest, not lowest, priority so we use greater than here.
	// 这边跟heap包中的不同，这里希望返回Priority最小的
	return pq.items[i].Priority < pq.items[j].Priority
}

func (pq priorityQueue) Swap(i, j int) {
	if pq.Len() > 1 {
		pq.items[i], pq.items[j] = pq.items[j], pq.items[i]
		pq.items[i].index = i
		pq.items[j].index = j
	}
}

func (pq *priorityQueue) Push(x interface{}) {
	n := len(pq.items)
	item := x.(*Item)
	item.index = n
	pq.items = append(pq.items, item)
}

func (pq *priorityQueue) Pop() interface{} {
	if len(pq.items) == 0 {
		return nil
	}
	old := pq.items
	n := len(old)
	item := old[n-1]
	pq.delItem = item
	pq.newItems = old[0 : n-1]
	return item
}

func (pq *priorityQueue) remove() {
	if pq.delItem != nil && pq.newItems != nil {
		old := pq.items
		n := len(old)
		item := old[n-1]
		if pq.delItem != item {
			return
		}
		old[n-1] = nil  // avoid memory leak
		item.index = -1 // for safety
		pq.delItem = nil
		pq.items = pq.newItems
		pq.newItems = nil
	}
}

// update modifies the priority and value of an Item in the queue.
func (pq *priorityQueue) update(item *Item, value interface{}, priority int64) {
	item.Value = value
	item.Priority = priority
	heap.Fix(pq, item.index)
}

func NewPriority() *PriorityQueue {
	return &PriorityQueue{}
}

func NewPriorityFromSlice(s []*Item) *PriorityQueue {
	pq := NewPriority()
	pq.InitFromSlice(s)
	return pq
}

type PriorityQueue struct {
	mu    sync.Mutex
	queue priorityQueue
}

func (pq *PriorityQueue) InitFromSlice(s []*Item) {
	pq.queue.items = append(pq.queue.items, s...)
	pq.init()
}

func (pq *PriorityQueue) init() {
	heap.Init(&pq.queue)
}

func (pq *PriorityQueue) Len() int {
	return len(pq.queue.items)
}

func (pq *PriorityQueue) Push(x interface{}) {
	heap.Push(&pq.queue, x)
}

// Pop 获取第一个元素，并删除
func (pq *PriorityQueue) Pop() interface{} {
	pq.mu.Lock()
	defer pq.mu.Unlock()
	if pq.Len() == 0 {
		return nil
	}
	item := heap.Pop(&pq.queue)
	pq.queue.remove()
	return item
}

//Peek 获取第一个元素，但不删除
func (pq *PriorityQueue) Peek() interface{} {
	pq.mu.Lock()
	defer pq.mu.Unlock()
	item := heap.Pop(&pq.queue)
	pq.queue.delItem = nil
	pq.queue.newItems = nil
	pq.queue.Swap(0, pq.Len()-1)
	return item
}

func (pq *PriorityQueue) Update(item *Item, value interface{}, priority int64) {
	pq.queue.update(item, value, priority)
}

func (pq *PriorityQueue) Clear() {
	pq.queue.items = nil
}
