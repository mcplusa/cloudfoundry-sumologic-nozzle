package eventQueue

import (
	. "bitbucket.org/mcplusa-ondemand/firehose-to-sumologic/events"
)

//Node to put in queue
type Node struct {
	event Event
}

// Queue is a basic FIFO queue based on a circular list that resizes as needed.
type Queue struct {
	nodes []*Node
	head  int
	tail  int
	count int
}

func NewNode(event Event) *Node {
	return &Node{
		event: event,
	}
}

func NewQueue(n []*Node) *Queue {
	return &Queue{
		nodes: n,
	}
}

func (q *Queue) GetNode() []*Node {
	return q.nodes
}

func (n *Queue) GetCount() int {
	return n.count
}

func (n *Node) GetNodeEvent() Event {
	return n.event
}

// Push adds a node to the queue.
func (q *Queue) Push(n *Node) {
	if q.head == q.tail && q.count > 0 {
		nodes := make([]*Node, len(q.nodes)*2)
		copy(nodes, q.nodes[q.head:])
		copy(nodes[len(q.nodes)-q.head:], q.nodes[:q.head])
		q.head = 0
		q.tail = len(q.nodes)
		q.nodes = nodes
	}
	q.nodes[q.tail] = n
	q.tail = (q.tail + 1) % len(q.nodes)
	q.count++
}

// Pop removes and returns a node from the queue in first to last order.
func (q *Queue) Pop() *Node {
	if q.count == 0 {
		return nil
	}
	node := q.nodes[q.head]
	q.head = (q.head + 1) % len(q.nodes)
	q.count--
	return node
}
