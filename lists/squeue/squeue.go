package squeue

import (
	"sync"
	"sync/atomic"
	"unsafe"
)

type Node struct {
	next  *Node
	Value any
}

type SQueue struct {
	currentNode *Node
	anchorNode  *Node

	mutex sync.Mutex
}

func New() *SQueue {
	node := &Node{}
	node.next = node
	return &SQueue{
		currentNode: node,
		anchorNode:  node,
		mutex:       sync.Mutex{},
	}
}

func (q *SQueue) IsEmpty() bool {
	return q.currentNode.next == q.anchorNode
}

func (q *SQueue) Push(val any) {
	node := &Node{
		Value: val,
	}
	node.next = node

	(*Node)(atomic.SwapPointer(
		(*unsafe.Pointer)(unsafe.Pointer(&q.currentNode)),
		unsafe.Pointer(node),
	)).next.next = node
}

func (q *SQueue) Pop() (val any, ok bool) {
	anchor := q.anchorNode

	node := anchor.next
	ok = anchor != node.next
	anchor.next = node.next
	node.next = anchor

	return node.Value, ok
}
