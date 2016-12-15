package eventQueue

import . "bitbucket.org/mcplusa-ondemand/firehose-to-sumologic/events"

//Node to put in queue
type Node struct {
	Event Event
}

// Queue is a basic FIFO queue based on a circular list that resizes as needed.
type Queue struct {
	Nodes []*Node
	head  int
	tail  int
	count int
}

func NewNode(event Event) *Node {
	return &Node{
		Event: event,
	}
}

func NewQueue(n []*Node) Queue {
	return Queue{
		Nodes: n,
	}
}

func (q *Queue) GetNode() []*Node {
	return q.Nodes
}

func (n *Queue) GetCount() int {
	return n.count
}

func (n *Node) GetNodeEvent() Event {
	return n.Event
}

// Push adds a node to the queue.
func (q *Queue) Push(n *Node) {
	if q.head == q.tail && q.count > 0 {
		nodes := make([]*Node, len(q.Nodes)*2)
		copy(nodes, q.Nodes[q.head:])
		copy(nodes[len(q.Nodes)-q.head:], q.Nodes[:q.head])
		q.head = 0
		q.tail = len(q.Nodes)
		q.Nodes = nodes
	}
	q.Nodes[q.tail] = n
	q.tail = (q.tail + 1) % len(q.Nodes)
	q.count++
}

// Pop removes and returns a node from the queue in first to last order.
func (q *Queue) Pop() *Node {
	if q.count == 0 {
		return nil
	}
	node := q.Nodes[q.head]
	q.head = (q.head + 1) % len(q.Nodes)
	q.count--
	return node
}
