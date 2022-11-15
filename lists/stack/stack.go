package stack

type Node struct {
	prev  *Node
	Value any
}

type Stack struct {
	node *Node
}

func New() *Stack {
	node := &Node{}
	node.prev = node
	return &Stack{
		node: node,
	}
}

func (obj *Stack) IsEmpty() bool {
	return obj.node == obj.node.prev
}

func (obj *Stack) Push(val any) {
	obj.node = &Node{
		prev:  obj.node,
		Value: val,
	}
}

func (obj *Stack) Pop() (val any, ok bool) {
	node := obj.node
	obj.node = node.prev
	return node.Value, node != obj.node
}
