package queue

type Node struct {
	next  *Node
	Value any
}

type Queue struct {
	currentNode *Node
	anchorNode  *Node
}

func New() *Queue {
	node := &Node{}
	node.next = node
	return &Queue{
		currentNode: node,
		anchorNode:  node,
	}
}

func (q *Queue) IsEmpty() bool {
	return q.currentNode.next == q.anchorNode
}

func (q *Queue) Push(val any) {
	node := &Node{
		Value: val,
	}
	node.next = node
	q.currentNode.next.next = node
	q.currentNode = node
}

func (q *Queue) Pop() (val any, ok bool) {
	node := q.anchorNode.next
	ok = q.anchorNode != node.next
	q.anchorNode.next = node.next
	node.next = q.anchorNode

	return node.Value, ok
}
