package list

// Node represents an element in the DoublyLinkedList storing a string key.
type Node struct {
	Value      string
	Prev, Next *Node
}

// DoublyLinkedList is a sentinel-node doubly linked list with O(1) operations.
type DoublyLinkedList struct {
	head *Node // sentinel head
	tail *Node // sentinel tail
	size int
}

// New creates and initializes a new DoublyLinkedList.
func New() *DoublyLinkedList {
	head := &Node{}
	tail := &Node{}
	head.Next = tail
	tail.Prev = head

	return &DoublyLinkedList{
		head: head,
		tail: tail,
		size: 0,
	}
}

// PushFront inserts a new node with value at the front of the list.
func (l *DoublyLinkedList) PushFront(val string) *Node {
	node := &Node{Value: val}
	l.insertAfter(node, l.head)
	return node
}

// PushBack inserts a new node with value at the back of the list.
func (l *DoublyLinkedList) PushBack(val string) *Node {
	node := &Node{Value: val}
	l.insertBefore(node, l.tail)
	return node
}

// MoveToFront moves an existing node to the front of the list.
func (l *DoublyLinkedList) MoveToFront(node *Node) {
	l.removeNode(node)
	l.insertAfter(node, l.head)
}

// Remove removes the given node from the list.
func (l *DoublyLinkedList) Remove(node *Node) {
	l.removeNode(node)
}

// PopBack removes and returns the last element, or nil if empty.
func (l *DoublyLinkedList) PopBack() *Node {
	if l.size == 0 {
		return nil
	}
	node := l.tail.Prev
	l.removeNode(node)
	return node
}

// PopFront removes and returns the first element, or nil if empty.
func (l *DoublyLinkedList) PopFront() *Node {
	if l.size == 0 {
		return nil
	}
	node := l.head.Next
	l.removeNode(node)
	return node
}

// Len returns the number of elements in the list.
func (l *DoublyLinkedList) Len() int {
	return l.size
}

func (l *DoublyLinkedList) insertAfter(node, at *Node) {
	node.Prev = at
	node.Next = at.Next
	at.Next.Prev = node
	at.Next = node
	l.size++
}

func (l *DoublyLinkedList) insertBefore(node, at *Node) {
	node.Next = at
	node.Prev = at.Prev
	at.Prev.Next = node
	at.Prev = node
	l.size++
}

func (l *DoublyLinkedList) removeNode(node *Node) {
	node.Prev.Next = node.Next
	node.Next.Prev = node.Prev
	node.Next = nil
	node.Prev = nil
	l.size--
}
