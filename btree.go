// Package btree implements the B-Tree data structure.
package btree

import (
	"errors"
	"fmt"
	"math"
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

// Insert inserts a new item into the BTree. If needed, it also rebalances the
// tree.
//
// Duplicate values cannot be inserted. If the item to insert is found in the
// tree, the method will fail silently.
func (b *BTree) Insert(item Item) {
	curr := b.root
	for {
		i := curr.items.find(item)

		if curr.items.match(item, i-1) {
			return
		} else if i >= len(curr.children) {
			break
		}

		curr = curr.children[i]
	}

	b.split(curr, item)
}

// Delete deletes an item from the B-Tree. If needed, it also rebalances the
// tree.
//
// If the item to delete does not exist in the tree, the method will fail
// silently.
func (b *BTree) Delete(item Item) {
	del, i := b.search(item)
	if i == -1 {
		return
	}
	// 1. Delete the item from its container node.
	// The container node must be either an internal node or a leaf.
	// If it is a leaf, we can simply delete the item, as leaves do not
	// have children that would be affected by a missing separator.
	// If it is a container, we replace the deleted value with the maximum
	// value of its left subtree, and delete the value from its original
	// container node.
	var affected *node
	if len(del.children) == 0 {
		del.items.delete(i)
		affected = del
	} else {
		maxNode := b.max(del.children[i])
		if len(maxNode.items) > 0 {
			del.items[i] = maxNode.items[len(maxNode.items)-1]
			maxNode.items.delete(len(maxNode.items) - 1)
		}
		affected = maxNode
	}
	// 2. Rebalance the tree around affected node.
	// NOTE: Because affected is a leaf, it is only considered unbalanced
	// if it is empty and not the root.
	if len(affected.items) == 0 && affected.parent != nil {
		minItems := 1
		b.rebalance(affected, minItems)
	}
}

// Search searches for an item in the Btree.
//
// If the item is found, the method returns a pointer to it.
// Otherwise, the function returns nil and an error indicating failure.
func (b *BTree) Search(item Item) (*Item, error) {
	curr := b.root
	for {
		i := curr.items.find(item)
		if curr.items.match(item, i-1) {
			return &curr.items[i-1], nil
		} else if i >= len(curr.children) {
			break
		}

		curr = curr.children[i]
	}

	return nil, errors.New("item not found in BTree")
}

// Bulkload initializes a BTree using an array of Items.
func (b *BTree) Bulkload(items items) {

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

// rebalance attempts to rebalance the tree around a given node.
// To do this, the function
func (b *BTree) rebalance(n *node, minItems int) {
	// Root does not have same invariants as other nodes so it is ignored.
	if n.parent == nil {
		return
	}

	// Positions of separator items.
	ptrIndex := n.nthChildOfParent()
	lSepPos, rSepPos := ptrIndex-1, ptrIndex
	var leftSib, rightSib, sibling *node
	if ptrIndex > 0 {
		leftSib = n.parent.children[ptrIndex-1]
	}
	if ptrIndex < len(n.parent.children)-1 {
		rightSib = n.parent.children[ptrIndex+1]
	}
	// Left rotation
	// NOTE: Important to also copy child nodes.
	if sibling = rightSib; sibling != nil && len(sibling.items) > minItems {
		n.items = append(n.items, n.parent.items[rSepPos])
		n.parent.items[rSepPos] = sibling.items[0]
		sibling.items.delete(0)
		if len(sibling.children) > 0 {
			sibling.children[0].parent = n
			n.children = append(n.children, sibling.children[0])
			sibling.children.delete(0)
		}
		return
	}

	// Right rotation
	// NOTE: Important to also copy child nodes.
	if sibling = leftSib; sibling != nil && len(sibling.items) > minItems {
		n.items = append(items{n.parent.items[lSepPos]}, n.items...)
		n.parent.items[lSepPos] = sibling.items[len(sibling.items)-1]
		sibling.items.delete(len(sibling.items) - 1)
		if len(sibling.children) > 0 {
			lastChild := sibling.children[len(sibling.children)-1]
			lastChild.parent = n
			n.children = append(children{lastChild}, n.children...)
			sibling.children.delete(len(sibling.children) - 1)
		}
		return
	}

	// Merge left node, separator, and right node, in that order.
	// NOTE: Must update right's children with their new parent (left).
	var left, right *node
	var sepPos, rightPos int
	if sibling = leftSib; sibling != nil {
		left = sibling
		right = n
		sepPos = lSepPos
		rightPos = ptrIndex
	} else {
		sibling = rightSib
		left = n
		right = sibling
		sepPos = rSepPos
		rightPos = ptrIndex + 1
	}
	left.items = append(left.items, n.parent.items[sepPos])
	left.items = append(left.items, right.items...)
	for _, c := range right.children {
		c.parent = left
	}
	left.children = append(left.children, right.children...)
	n.parent.items.delete(sepPos)
	n.parent.children.delete(rightPos)

	// Left becomes new root if parent is root and empty.
	if n.parent.parent == nil && len(n.parent.items) == 0 {
		right.parent = left
		left.parent = nil
		b.root = left
		return
	}

	// If B-Tree invariants don't hold for parent, rebalance around parent.
	minItems = int(math.Ceil(float64(b.order)/2.0)) - 1
	if len(n.parent.items) < minItems {
		b.rebalance(n.parent, minItems)
	}
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
		} else if i >= len(curr.children) {
			return nil, -1
		}
		curr = curr.children[i]
	}
}

// max returns the rightmost node of a particular subtree.
func (b *BTree) max(root *node) *node {
	curr := root
	for {
		if len(curr.children) == 0 {
			return curr
		}
		curr = curr.children[len(curr.children)-1]
	}
}

// find returns the index of the item in items.
// If item does not exist in items, return where it would be located
// (where 0 <= index <= len(array)).
func (its *items) find(it Item) int {
	return sort.Search(len(*its), func(i int) bool { return it.Less((*its)[i]) })
}

// match checks if item and given index is equal to given item.
func (its *items) match(item Item, index int) bool {
	if index >= 0 && index < len(*its) &&
		!(item.Less((*its)[index]) || (*its)[index].Less(item)) {
		return true
	}
	return false
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

func (chi *children) delete(index int) {
	copy((*chi)[index:], (*chi)[index+1:])
	(*chi)[len(*chi)-1] = nil
	*chi = (*chi)[:len(*chi)-1]
}

func (chi *children) truncate(newLen int) {
	for i := newLen; i < len(*chi); i++ {
		(*chi)[i] = nil
	}
	*chi = (*chi)[:newLen]
}

func (chi *children) indexOf(n *node) int {
	for i, p := range *chi {
		if p == n {
			return i
		}
	}
	return -1
}

// nthChildOfParent returns the index of the child in n.parent which points to
// n.
// NOTE: Because find() uses binary search, we defer to it when possible.
func (n *node) nthChildOfParent() int {
	if n.parent == nil {
		return -1
	}
	if len(n.items) == 0 {
		return n.parent.children.indexOf(n)
	}
	return n.parent.items.find(n.items[0])
}

//=============================================================================
//= Functions
//=============================================================================

// NewBTree returns a new BTree.
func NewBTree(order int) *BTree {
	return &BTree{
		order: order,
		root:  newNode(order, nil, nil, nil),
	}
}

// newNode returns a new node.
func newNode(order int, i items, c children, parent *node) *node {
	return &node{
		items:    i,
		children: c,
		parent:   parent,
	}
}

type walker struct {
	n      *node
	height int
}

// walk prints a representation of the BTree in level order.
// Exists for testing purposes.
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
