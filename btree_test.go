package btree

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

func TestInsert(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	numItems := 1000
	massItems := make(map[int]*testItem, numItems)
	dupItems := make(map[int]*testItem, numItems)
	for i := 0; i < numItems; i++ {
		massItems[i] = &testItem{i, i}
		dupItems[i] = massItems[0]
	}
	insertTests := []struct {
		items map[int]*testItem
		order int
	}{
		// Many random unique items
		{items: massItems, order: 3},
		// Many random unique items with different order
		{items: massItems, order: 6},
		// Many random unique items with different order
		{items: massItems, order: 11},
		// Many random unique items with minimum order 2
		{items: massItems, order: 2},
		// Duplicate items
		{items: dupItems, order: 5},
	}
	for _, ti := range insertTests {
		b := New(ti.order)
		i := 0
		for _, item := range ti.items {
			b.Insert(item)
			i++

			if !isValidBTree(b) {
				walk(b.root)
				t.Fatalf("After Insert: BTree is not valid after %dth insert of item %v\n", i+1, item)
			}
		}
	}
}

func TestDelete(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	numItems := 1000
	massItems := make(map[int]*testItem, numItems)
	for i := 0; i < numItems; i++ {
		massItems[i] = &testItem{i, i}
	}
	emptyItems := make(map[int]*testItem)
	dneItems := map[int]*testItem{0: &testItem{key: -999, val: 0}}
	cases := []struct {
		items           map[int]*testItem
		order           int
		toDelete        map[int]*testItem
		shouldAlterTree bool
	}{
		// Delete should work on empty tree
		{items: emptyItems, order: 5, toDelete: emptyItems, shouldAlterTree: false},
		{items: emptyItems, order: 3, toDelete: emptyItems, shouldAlterTree: false},
		// Delete should work for item not in tree
		{items: massItems, order: 3, toDelete: dneItems, shouldAlterTree: false},
		// Should fully delete trees of various orders
		{items: massItems, order: 30, toDelete: massItems, shouldAlterTree: true},
		{items: massItems, order: 8, toDelete: massItems, shouldAlterTree: true},
		{items: massItems, order: 5, toDelete: massItems, shouldAlterTree: true},
		{items: massItems, order: 3, toDelete: massItems, shouldAlterTree: true},
	}

	for _, c := range cases {
		b := New(c.order)
		for _, v := range c.items {
			b.Insert(v)
		}

		for i, d := range c.toDelete {
			_, presentBefore := b.Search(d)
			b.Delete(d)

			if c.shouldAlterTree {
				_, presentAfter := b.Search(d)
				if presentBefore == presentAfter {
					t.Errorf("Item should have been deleted from tree\n")
				}
			}

			if !isValidBTree(b) {
				walk(b.root)
				t.Fatalf("After Delete: BTree is not valid after %dth deletion. Item was %v\n", i, c.toDelete)
			}
		}

	}
}

func TestSearch(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	numItems := 1000
	massItems := make(map[int]*testItem, numItems)
	for i := 0; i < numItems; i++ {
		massItems[i] = &testItem{i, i}
	}
	cases := []struct {
		items      map[int]*testItem
		order      int
		lookFor    []*testItem
		shouldFind bool
	}{
		// Should be able to search empty tree
		{items: map[int]*testItem{}, order: 4, lookFor: []*testItem{massItems[0]}, shouldFind: false},
		// Should successfully find present node with particular order
		{items: massItems, order: 5, lookFor: []*testItem{massItems[1], massItems[2], massItems[3]}, shouldFind: true},
		// Should successfully find present node with different order
		{items: massItems, order: 11, lookFor: []*testItem{massItems[1], massItems[2], massItems[3]}, shouldFind: true},
		// Should successfully not find missing node
		{items: massItems, order: 2, lookFor: []*testItem{{key: -9999, val: 0}}, shouldFind: false},
	}

	for _, c := range cases {
		b := New(c.order)
		for _, ti := range c.items {
			b.Insert(ti)
		}

		for _, target := range c.lookFor {
			_, err := b.Search(target)
			if err != nil && c.shouldFind == true {
				t.Errorf("Should have found: %v\n", target)
			} else if err == nil && c.shouldFind == false {
				t.Errorf("Should not have found: %v\n", target)
			}
		}
	}

}

func TestIteratorNext(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	numItems := 1000
	massItems := make(map[int]*testItem, numItems)
	for i := 0; i < numItems; i++ {
		massItems[i] = &testItem{i, i}
	}
	cases := []struct {
		items map[int]*testItem
		order int
	}{
		{items: massItems, order: 3},
	}

	for _, c := range cases {
		b := New(c.order)
		for _, v := range c.items {
			b.Insert(v)
		}
		iter := b.NewIterator()

		var prev Item
		for i := 0; i < len(c.items); i++ {
			next, err := iter.Next()
			if err != nil {
				t.Errorf("Call to Next() should not have returned non-nil error")
			}
			if prev != nil && !prev.Less(next) {
				t.Errorf("Values from Iterator should be ascending. Prev: %v, Next: %v", prev, next)
			}
			prev = next
		}

		if iter.HasNext() {
			t.Errorf("Iterator should no longer have next")
		}
		extraIterVal, err := iter.Next()
		if extraIterVal != nil || err == nil {
			t.Errorf("Extra call to Next() should have returned nil value and error. Instead got val: %v, and error: %v", extraIterVal, err)
		}
	}
}

func TestIteratorReverseNext(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	numItems := 1000
	massItems := make(map[int]*testItem, numItems)
	for i := 0; i < numItems; i++ {
		massItems[i] = &testItem{i, i}
	}
	cases := []struct {
		items map[int]*testItem
		order int
	}{
		{items: massItems, order: 3},
	}

	for _, c := range cases {
		b := New(c.order)
		for _, v := range c.items {
			b.Insert(v)
		}
		iter := b.NewReverseIterator()

		for i := 0; i < len(c.items); i++ {
			iter.Next()
		}
	}
}

func TestBulkload(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	numItems := 1000
	massItems := make([]Item, numItems)
	for i := 0; i < numItems; i++ {
		massItems[i] = &testItem{i, i}
	}
	cases := []struct {
		items []Item
		order int
	}{
		{items: massItems, order: 5},
		{items: massItems, order: 12},
		{items: massItems, order: 3},
	}

	for _, c := range cases {
		bt := Bulkload(c.order, c.items)

		if !isValidBTree(bt) {
			walk(bt.root)
			t.Errorf("Bulkloaded tree is not valid\n")
		}
	}

}

func TestMerge(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	numFirst := 5000
	numSecond := 5000
	firstItems := make([]Item, numFirst)
	secondItems := make([]Item, numSecond)
	for i := 0; i < numFirst; i++ {
		firstItems[i] = &testItem{i, i}
	}
	for i := 0; i < numSecond; i++ {
		secondItems[i] = &testItem{i + numFirst/2, i + numFirst/2}
	}
	cases := []struct {
		first  []Item
		second []Item
		order  int
	}{
		{first: firstItems, second: secondItems, order: 5},
	}

	for _, c := range cases {
		firstTree := Bulkload(c.order, c.first)
		secondTree := Bulkload(c.order, c.second)

		mt, err := Merge(firstTree, secondTree)
		if err != nil || !isValidBTree(mt) {
			walk(mt.root)
			t.Errorf("Merged tree should have been valid.")
		}
	}
}

//=============================================================================
//= Benchmarks
//=============================================================================

var bt *BTree

func init() {
	rand.Seed(time.Now().UnixNano())
	numItems := 10000
	massItems := make(map[int]*testItem, numItems)
	for i := 0; i < numItems; i++ {
		massItems[i] = &testItem{i, i}
	}
	bt = New(20)
	for _, v := range massItems {
		bt.Insert(v)
	}
}

func BenchmarkIterator(b *testing.B) {
	for n := 0; n < b.N; n++ {
		iter := bt.NewIterator()
		iterateThrough(iter)
	}
}

func BenchmarkIteratorReverse(b *testing.B) {
	for n := 0; n < b.N; n++ {
		iter := bt.NewReverseIterator()
		iterateThrough(iter)
	}
}

func iterateThrough(iter *Iterator) {
	for {
		_, err := iter.Next()
		if err != nil {
			return
		}
	}
}

//=============================================================================
//= Helpers
//=============================================================================

// A testItem is a simple type which implements the Item interface.
// In all tests, B-Trees will contain testItems.
type testItem struct {
	key int
	val int
}

func (ti *testItem) Less(other Item) bool {
	o := other.(*testItem)
	return ti.key < o.key
}

func (ti *testItem) String() string {
	return fmt.Sprintf("(k: %d, v: %d),", ti.key, ti.val)
}

// atMostChildren recursively checks that very node in a BTree has at most
// 'order' children (max = tree order).
func atMostChildren(curr *node, max int) bool {
	if len(curr.children) > max {
		return false
	}
	for _, c := range curr.children {
		if !atMostChildren(c, max) {
			return false
		}
	}
	return true
}

// atLeastChildren checks that every non-leaf AND non-root node has at least
// order / 2 children (min = order / 2).
func atLeastChildren(curr *node, min int) bool {
	if len(curr.children) == 0 {
		return true
	}

	// Check only non-leaf, non-root nodes.
	if curr.parent != nil && len(curr.children) < min {
		return false
	}
	for _, c := range curr.children {
		if !atLeastChildren(c, min) {
			return false
		}
	}
	return true

}

// atLeastChildrenRoot checks that the tree's root is either a leaf or has at
// least 2 children.
func atLeastChildrenRoot(root *node) bool {
	if len(root.children) == 0 {
		return true
	}
	if len(root.children) >= 2 {
		return true
	}
	return false
}

// rightNumKeys checks that for non-leaf nodes, the number of children is
// always 1 more than the number of items.
func rightNumKeys(curr *node) bool {
	if len(curr.children) == 0 {
		return true
	}
	if len(curr.children)-len(curr.items) != 1 {
		return false
	}
	for _, c := range curr.children {
		if !rightNumKeys(c) {
			return false
		}
	}
	return true
}

func allLeavesSameDepthRecurse(curr *node, currDepth, wantDepth int) bool {
	if len(curr.children) == 0 {
		if currDepth != wantDepth {
			return false
		}
	}
	for _, c := range curr.children {
		if !allLeavesSameDepthRecurse(c, currDepth+1, wantDepth) {
			return false
		}
	}
	return true
}

// allLeavesSameDepth checks that every leaf node in the tree has the same
// depth.
// It does this by first calculating the depth of the left most leaf, then
// recursively checking that all other leaves have that same depth.
func allLeavesSameDepth(root *node) bool {
	expectedDepth := 0
	curr := root
	for len(curr.children) > 0 {
		expectedDepth++
		curr = curr.children[0]
	}

	return allLeavesSameDepthRecurse(root, 0, expectedDepth)
}

// allBetweenBounds checks that the values in each subtree are correctly
// bounded.
func allBetweenBounds(curr *node) bool {
	for i, c := range curr.children {
		if i == 0 {
			// Check that every item in leftmost child is less than
			// leftmost item in current node.
			// We break if len(curr.items) == 0, as that implies no
			// comparisons are possible for current node's children.
			if len(curr.items) == 0 {
				break
			}
			for _, childItem := range c.items {
				if !childItem.Less(curr.items[i]) {
					return false
				}
			}
		} else if i < len(curr.items) {
			// For every child between 1 and last-1, check that every item
			// of that child is in the open interval
			// (curr.items[i-1], curr.items[i])
			for _, childItem := range c.items {
				if !(curr.items[i-1].Less(childItem) && childItem.Less(curr.items[i])) {
					return false
				}
			}
		} else if i == len(curr.items) {
			// For final child, check that every element is
			// strictly greater than last item.
			for _, childItem := range c.items {
				if !curr.items[i-1].Less(childItem) {
					return false
				}
			}
		}
	}

	for _, c := range curr.children {
		if !allBetweenBounds(c) {
			return false
		}
	}

	return true
}

// isValidBTree checks that given tree satisfies the definition of a
// B-tree.
// This function should be used at the end of each test.
func isValidBTree(tree *BTree) bool {
	// For BTree of order m:
	// 1. Every node has at most m children
	if !atMostChildren(tree.root, tree.order) {
		fmt.Printf("Every node must have at most order m children\n")
		return false
	}
	// 2. Every non-leaf node (except root) has at least [m/2] children
	if !atLeastChildren(tree.root, tree.order/2) {
		fmt.Printf("Every non-leaf node must have at least order m / 2 children\n")
		return false
	}
	// 3. The root has at least two children if it is not a leaf
	if !atLeastChildrenRoot(tree.root) {
		fmt.Printf("Every non-leaf root must have at least 2 children\n")
		return false
	}
	// 4. A non-leaf node with k children contains k-1 keys
	if !rightNumKeys(tree.root) {
		fmt.Printf("Every non-leaf node with k children must have k-1 keys\n")
		return false
	}
	// 5. All leaves appear in the same level
	if !allLeavesSameDepth(tree.root) {
		fmt.Printf("All leaves must have same depth\n")
		return false
	}
	// 6. Values in all subtrees are properly bounded by items in subtree's
	// root.
	if !allBetweenBounds(tree.root) {
		fmt.Printf("All subtrees must be properly bounded\n")
		return false
	}

	return true
}
