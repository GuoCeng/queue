package queue

import (
	"testing"
)

func TestPriorityQueue_Pop(t *testing.T) {
	// Some items and their priorities.
	items := map[string]int64{
		"banana": 10, "apple": 1, "pear": 12, "banana1": 3, "apple1": 24, "pear1": 15, "banana2": 6, "apple2": 7, "pear2": 8,
	}

	// Create a priority queue, put the items in it, and
	// establish the priority queue (heap) invariants.
	pq := NewPriority()
	source := make([]*Item, len(items))
	i := 0
	for value, priority := range items {
		source[i] = &Item{
			Value:    value,
			Priority: priority,
		}
		i++
	}
	pq.InitFromSlice(source)
	//pq.Clear()

	// Insert a new item and then modify its priority.
	item := &Item{
		Value:    "orange",
		Priority: 1,
	}
	pq.Push(item)
	pq.Update(item, item.Value, 9)

	item = pq.Peek().(*Item)
	t.Logf("11111111111111111   %.2d:%s ", item.Priority, item.Value)
	item = pq.Pop().(*Item)
	t.Logf("22222222222222222   %.2d:%s ", item.Priority, item.Value)
	// Take the items out; they arrive in decreasing priority order.
	for pq.Len() > 0 {
		item := pq.Pop().(*Item)
		t.Logf("%.2d:%s ", item.Priority, item.Value)
	}

	if item, ok := pq.Pop().(*Item); ok {
		t.Logf("%.2d:%s ", item.Priority, item.Value)
	}
	// Output:
	// 05:orange 04:pear 03:banana 02:apple
}
