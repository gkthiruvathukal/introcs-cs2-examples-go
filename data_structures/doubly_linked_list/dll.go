package main

import "fmt"

type Node struct {
	data   any
	nodeID int
	prev   *Node
	next   *Node
}

type DoublyLinkedList struct {
	head        *Node
	tail        *Node
	size        int
	nextNodeID  int
}

func NewDoublyLinkedList() *DoublyLinkedList {
	return &DoublyLinkedList{nextNodeID: 1}
}

func (l *DoublyLinkedList) Insert(data any, nodeID ...int) {
	id := l.nextNodeID
	if len(nodeID) > 0 && nodeID[0] > 0 {
		id = nodeID[0]
		if id >= l.nextNodeID {
			l.nextNodeID = id + 1
		}
	} else {
		l.nextNodeID++
	}
	n := &Node{data: data, nodeID: id}
	if l.head == nil {
		l.head = n
		l.tail = n
	} else {
		n.prev = l.tail
		l.tail.next = n
		l.tail = n
	}
	l.size++
}

func (l *DoublyLinkedList) unlink(n *Node) {
	if n.prev != nil {
		n.prev.next = n.next
	} else {
		l.head = n.next
	}
	if n.next != nil {
		n.next.prev = n.prev
	} else {
		l.tail = n.prev
	}
	l.size--
}

func (l *DoublyLinkedList) nodeAt(index int) (*Node, error) {
	if index < 0 || index >= l.size {
		return nil, fmt.Errorf("index %d out of range for list of size %d", index, l.size)
	}
	cur := l.head
	for i := 0; i < index; i++ {
		cur = cur.next
	}
	return cur, nil
}

func (l *DoublyLinkedList) insertBeforeNode(target *Node, data any, nodeID ...int) {
	id := l.nextNodeID
	if len(nodeID) > 0 && nodeID[0] > 0 {
		id = nodeID[0]
		if id >= l.nextNodeID {
			l.nextNodeID = id + 1
		}
	} else {
		l.nextNodeID++
	}
	n := &Node{data: data, nodeID: id, next: target, prev: target.prev}
	if target.prev != nil {
		target.prev.next = n
	} else {
		l.head = n
	}
	target.prev = n
	l.size++
}

func (l *DoublyLinkedList) Remove(data any) error {
	cur := l.head
	for cur != nil {
		if fmt.Sprintf("%v", cur.data) == fmt.Sprintf("%v", data) {
			l.unlink(cur)
			return nil
		}
		cur = cur.next
	}
	return fmt.Errorf("value not found in the list")
}

func (l *DoublyLinkedList) RemoveAt(index int) error {
	n, err := l.nodeAt(index)
	if err != nil {
		return err
	}
	l.unlink(n)
	return nil
}

func (l *DoublyLinkedList) RemoveByNodeID(nodeID int) error {
	cur := l.head
	for cur != nil {
		if cur.nodeID == nodeID {
			l.unlink(cur)
			return nil
		}
		cur = cur.next
	}
	return fmt.Errorf("no node with id %d", nodeID)
}

func (l *DoublyLinkedList) InsertAt(index int, data any, nodeID ...int) error {
	if index < 0 || index > l.size {
		return fmt.Errorf("insert index %d out of range for list of size %d", index, l.size)
	}
	if index == l.size {
		l.Insert(data, nodeID...)
		return nil
	}
	target, err := l.nodeAt(index)
	if err != nil {
		return err
	}
	l.insertBeforeNode(target, data, nodeID...)
	return nil
}

func (l *DoublyLinkedList) InsertAfter(nodeID int, data any, newNodeID ...int) error {
	cur := l.head
	for cur != nil {
		if cur.nodeID == nodeID {
			if cur == l.tail {
				l.Insert(data, newNodeID...)
			} else {
				l.insertBeforeNode(cur.next, data, newNodeID...)
			}
			return nil
		}
		cur = cur.next
	}
	return fmt.Errorf("no node with id %d", nodeID)
}

func (l *DoublyLinkedList) InsertBefore(nodeID int, data any, newNodeID ...int) error {
	cur := l.head
	for cur != nil {
		if cur.nodeID == nodeID {
			l.insertBeforeNode(cur, data, newNodeID...)
			return nil
		}
		cur = cur.next
	}
	return fmt.Errorf("no node with id %d", nodeID)
}

type FindResult struct {
	Index  int
	NodeID int
}

func (l *DoublyLinkedList) Find(data any) *FindResult {
	cur := l.head
	idx := 0
	for cur != nil {
		if fmt.Sprintf("%v", cur.data) == fmt.Sprintf("%v", data) {
			return &FindResult{Index: idx, NodeID: cur.nodeID}
		}
		cur = cur.next
		idx++
	}
	return nil
}

func (l *DoublyLinkedList) ToList() []any {
	var items []any
	cur := l.head
	for cur != nil {
		items = append(items, cur.data)
		cur = cur.next
	}
	return items
}

func (l *DoublyLinkedList) ToListBackward() []any {
	var items []any
	cur := l.tail
	for cur != nil {
		items = append(items, cur.data)
		cur = cur.prev
	}
	return items
}

func (l *DoublyLinkedList) Clear() {
	l.head = nil
	l.tail = nil
	l.size = 0
	l.nextNodeID = 1
}

type NodeSnapshot struct {
	NodeID    int
	NodeLabel string
	Data      any
	PrevLabel string
	NextLabel string
}

func (l *DoublyLinkedList) NodeSnapshots() []NodeSnapshot {
	var snaps []NodeSnapshot
	cur := l.head
	for cur != nil {
		prevLabel := "/"
		if cur.prev != nil {
			prevLabel = fmt.Sprintf("node-%d", cur.prev.nodeID)
		}
		nextLabel := "/"
		if cur.next != nil {
			nextLabel = fmt.Sprintf("node-%d", cur.next.nodeID)
		}
		snaps = append(snaps, NodeSnapshot{
			NodeID:    cur.nodeID,
			NodeLabel: fmt.Sprintf("node-%d", cur.nodeID),
			Data:      cur.data,
			PrevLabel: prevLabel,
			NextLabel: nextLabel,
		})
		cur = cur.next
	}
	return snaps
}

func (l *DoublyLinkedList) RestoreFromSnapshots(snaps []NodeSnapshot) {
	l.Clear()
	for _, s := range snaps {
		l.Insert(s.Data, s.NodeID)
	}
}

func (l *DoublyLinkedList) At(index int) (any, error) {
	n, err := l.nodeAt(index)
	if err != nil {
		return nil, err
	}
	return n.data, nil
}
