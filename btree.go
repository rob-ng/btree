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
	var lc children
	var rc children
	if len(node.children) > 0 {
		lc = append(lc, node.children[:mid+1]...)
		rc = append(rc, node.children[mid+1:]...)
	}
	li := append(items(nil), node.items[:mid]...)
	ri := append(items(nil), node.items[mid+1:]...)
	leftNode := newNode(b.order, li, lc, node.parent)
	rightNode := newNode(b.order, ri, rc, node.parent)
	for _, c := range leftNode.children {
		c.parent = leftNode
	}
	for _, c := range rightNode.children {
		c.parent = rightNode
	}

	// If at root, create new root
	if node.parent == nil {
		newRoot := newNode(b.order, items{midItem}, children{leftNode, rightNode}, nil)
		leftNode.parent = newRoot
		rightNode.parent = newRoot
		b.root = newRoot
		return
	}

	// Overwrite reference to existing parent at index i.
	i := node.parent.items.find(item)
	node.parent.children = append(node.parent.children, nil)
	copy(node.parent.children[i+1:], node.parent.children[i:])
	node.parent.children[i] = leftNode
	node.parent.children[i+1] = rightNode
	b.split(node.parent, midItem)
}

// Insert inserts a new item into the BTree.
// TODO: Handle case where item is equal to item alreay in tree
// Try during for loop below.
func (b *BTree) Insert(item Item) {
	curr := b.root
	i := curr.items.find(item)
	if i-1 >= 0 && i-1 < len(curr.items) && !(item.Less(curr.items[i-1]) || curr.items[i-1].Less(item)) {
		return
	}
	for len(curr.children) > 0 {
		i := curr.items.find(item)
		// If equal item is found, return without inserting.
		// NOTE: find() returns the smallest index i such that
		// items[i] < item.
		// So if item is a duplicate, exisitng value will be at i-1.
		if i-1 >= 0 && i-1 < len(curr.items) && !(item.Less(curr.items[i-1]) || curr.items[i-1].Less(item)) {
			return
		}
		curr = curr.children[i]
	}
	b.split(curr, item)
}

// Delete deletes an item from the Btree.
func (b *BTree) Delete(item Item) {

}

// findEqual is a stricter version of find().
// The function will only return a valid index if an exact match is found.
// Otherwise the function returns -1.
func (its *items) findEqual(item Item) int {
	i := (*its).find(item)
	if i-1 >= 0 && i-1 < len(*its) && !(item.Less((*its)[i-1]) || (*its)[i-1].Less(item)) {
		return i - 1
	}
	return -1
}

// Search searches for an item in the Btree.
// On success, returns a pointer to the value and sets error to nil.
// On failure, returns nil and error.
func (b *BTree) Search(item Item) (*Item, error) {
	curr := b.root
	for {
		if found := curr.items.findEqual(item); found != -1 {
			return &curr.items[found], nil
		}

		i := curr.items.find(item)
		if i >= len(curr.children) || curr.children == nil {
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
	if i == nil {
		i = make(items, 0, order-1)
	}
	if c == nil {
		c = make(children, 0, order)
	}
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
