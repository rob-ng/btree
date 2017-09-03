package btree

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
}

//=============================================================================
//= Methods
//=============================================================================

// Insert inserts a new item into the BTree.
func (b *BTree) Insert(item Item) {

}

// Delete deletes an item from the Btree.
func (b *BTree) Delete(item Item) {

}

// Search searches for an item in the Btree.
func (b *BTree) Search(item Item) {

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
func newNode(order int) *node {
	return &node{
		items:    make(items, 0, order-1),
		children: make(children, 0, order),
	}
}

// NewBTree returns a new BTree.
func NewBTree(order int) *BTree {
	return &BTree{
		order: order,
		root:  newNode(order),
	}
}
