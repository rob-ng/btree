package btree

import (
	"errors"
	"fmt"
	"sort"
)

//=============================================================================
//= Types
//=============================================================================

// A BTree represents a B-Tree.
type BTree struct {
	order int   // Maximum number of children each node can have.
	root  *node // Root node of BTree.
}

// An Item is an element which can be compared to another Item.
type Item interface {
	Less(other Item) bool
}

type items []Item

type children []*node

type node struct {
	items    items
	children children
	parent   *node
}

//=============================================================================
//= Methods
//=============================================================================

// find returns the index of the item in items.
// If item does not exist in items, return where it would be located
// (where 0 <= index <= len(array)).
func (its *items) find(it Item) int {
	return sort.Search(len(*its), func(i int) bool { return it.Less((*its)[i]) })
}

// insert inserts the item into items.
// It returns the index where the item was inserted.
func (its *items) insert(it Item) int {
	i := its.find(it)
	*its = append(*its, nil)
	copy((*its)[i+1:], (*its)[i:])
	(*its)[i] = it
	return i
}

func (its *items) truncate(newLen int) {
	for i := newLen; i < len(*its); i++ {
		(*its)[i] = nil
	}
	*its = (*its)[:newLen]
}

func (its *items) delete(index int) {
	copy((*its)[index:], (*its)[index+1:])
	(*its)[len(*its)-1] = nil
	*its = (*its)[:len(*its)-1]
}

func (chi *children) truncate(newLen int) {
	for i := newLen; i < len(*chi); i++ {
		(*chi)[i] = nil
	}
	*chi = (*chi)[:newLen]
}

// split inserts an item into a particular node.
// After inserting the item into the node's 'items' field, the function
// performs a series of checks / operations to ensure that the B-Tree remains
// balanced and its invariants hold.
// Note that this process can be recursive.
func (b *BTree) split(node *node, item Item) {
	node.items.insert(item)
	if len(node.items) < b.order {
		return
	}

	mid := len(node.items) / 2
	midItem := node.items[mid]
	rightNode := newNode(b.order, nil, nil, node.parent)
	rightNode.items = append(rightNode.items, node.items[mid+1:]...)
	node.items.truncate(mid)
	if len(node.children) > 0 {
		rightNode.children = append(rightNode.children, node.children[mid+1:]...)
		node.children.truncate(mid + 1)
		for _, c := range rightNode.children {
			c.parent = rightNode
		}
	}

	if node.parent == nil {
		newRoot := newNode(b.order, items{midItem}, children{node, rightNode}, nil)
		node.parent = newRoot
		rightNode.parent = newRoot
		b.root = newRoot
		return
	}

	i := node.parent.items.find(item)
	node.parent.children = append(node.parent.children, nil)
	copy(node.parent.children[i+1:], node.parent.children[i:])
	node.parent.children[i+1] = rightNode
	b.split(node.parent, midItem)
}

// Insert inserts a new item into the BTree.
func (b *BTree) Insert(item Item) {
	curr := b.root
	// Loop continues until reaching leaf or finding equal item.
	for {
		i := curr.items.find(item)

		if curr.items.match(item, i-1) {
			return
		} else if i >= len(curr.children) || curr.children == nil {
			break
		}

		curr = curr.children[i]
	}

	b.split(curr, item)
}

// search searches for an item in the tree.
// It returns the node containing item and the index of item in the items
// array.
func (b *BTree) search(item Item) (*node, int) {
	curr := b.root
	for {
		i := curr.items.find(item)
		if curr.items.match(item, i-1) {
			return curr, i - 1
		} else if i >= len(curr.children) || curr.children == nil {
			return nil, -1
		}
		curr = curr.children[i]
	}
}

// max returns the rightmost node of a particular subtree.
func (b *BTree) max(root *node) *node {
	curr := root
	for {
		if curr.children == nil {
			return curr
		}
		curr = curr.children[len(curr.children)-1]
	}
}

func (b *BTree) rebalance(node *node) {
	return
}

// Delete deletes an item from the Btree.
func (b *BTree) Delete(item Item) {
	// 1. Search for node containing element
	del, i := b.search(item)
	if i == -1 {
		return
	}
	var affected *node
	if del.children == nil {
		//    A. If element is in leaf, simply delete value from node.items
		del.items.delete(i)
		affected = del
	} else {
		//    B. If element is in internal, we need to find replacement for
		//    deleted item as separation value. To do this, we either find
		//    largest element in left subtree or smallest element in right
		//    subtree.
		// Replacement is max value of items less than deleted value.
		// Recall that for any item at index i, children[i] gives left
		// subtree.
		maxNode := b.max(del.children[i])
		del.items[i] = maxNode.items[len(maxNode.items)-1]
		maxNode.items.delete(len(maxNode.items) - 1)
		affected = maxNode
	}
	// 2. Replace the tree.
	b.rebalance(affected)
}

// match checks if item and given index is equal to given item.
func (its *items) match(item Item, index int) bool {
	if index >= 0 && index < len(*its) &&
		!(item.Less((*its)[index]) || (*its)[index].Less(item)) {
		return true
	}
	return false
}

// Search searches for an item in the Btree.
// On success, returns a pointer to the value and sets error to nil.
// On failure, returns nil and error.
func (b *BTree) Search(item Item) (*Item, error) {
	curr := b.root
	for {
		i := curr.items.find(item)
		if curr.items.match(item, i-1) {
			return &curr.items[i-1], nil
		} else if i >= len(curr.children) || curr.children == nil {
			break
		}

		curr = curr.children[i]
	}

	return nil, errors.New("item not found in BTree")
}

// Bulkload initializes a BTree using an array of items.
func (b *BTree) Bulkload(items items) {

}

//=============================================================================
//= Functions
//=============================================================================

// newNode returns a new node.
// The node's fields are initalized with 0 length and capacity determined by
// the BTree's order.
// Note that by definition a non-leaf node with k children contains k-1 keys.
// Hence the capacity of items is 1 less than that of children.
func newNode(order int, i items, c children, parent *node) *node {
	return &node{
		items:    i,
		children: c,
		parent:   parent,
	}
}

// NewBTree returns a new BTree.
func NewBTree(order int) *BTree {
	return &BTree{
		order: order,
		root:  newNode(order, nil, nil, nil),
	}
}

type walker struct {
	n      *node
	height int
}

func walk(root *node) {
	total := 0
	q := []walker{{root, 0}}
	if len(q) > 0 {
		total += len(root.items)
	}
	currHeight := 0
	var currParent *node
	for len(q) > 0 {
		curr := q[0]
		if curr.height != currHeight {
			fmt.Println()
			fmt.Println()
			currHeight = curr.height
		}
		//fmt.Printf("(p: %p, v: %v, c: %v) ", curr.n, curr.n.items, curr.n.children)
		if curr.n.parent != currParent {
			fmt.Printf("\n| ")
			currParent = curr.n.parent
		}
		fmt.Printf("(s: %p, p: %p, v: %v, nc: %d) ", curr.n, curr.n.parent, curr.n.items, len(curr.n.children))
		q = q[1:]
		for _, c := range curr.n.children {
			if c != nil {
				w := walker{c, curr.height + 1}
				q = append(q, w)
				total += len(c.items)
			}
		}
	}
	fmt.Printf("total nodes: %d\n", total)
}
