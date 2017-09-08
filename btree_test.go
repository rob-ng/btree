package btree

import (
	"fmt"
	"math/rand"
	"testing"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func TestPrint(t *testing.T) {
	massItems := uniqueInputsN(500)
	b := New(5)
	for _, item := range massItems {
		b.Insert(item)
	}
	b.Print()
}

func TestInsert(t *testing.T) {
	massItems := uniqueInputsN(1000)
	dupItems := duplicateInputsN(1000)
	insertTests := []struct {
		items []Item
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
		for i, item := range ti.items {
			b.Insert(item)

			if !isValidBTree(b) {
				walk(b.root)
				t.Fatalf("After Insert: BTree is not valid after %dth insert of item %v\n", i+1, item)
			}
		}
	}
}

func TestDelete(t *testing.T) {
	massItems := uniqueInputsN(1000)
	emptyItems := uniqueInputsN(0)
	dneItems := []Item{&testItem{key: -999, val: 0}}
	cases := []struct {
		items           []Item
		toDelete        []Item
		order           int
		shouldAlterTree bool
	}{
		// Delete should work on empty tree
		{items: emptyItems, toDelete: massItems, order: 3, shouldAlterTree: false},
		{items: emptyItems, toDelete: massItems, order: 6, shouldAlterTree: false},
		// Delete should work for item not in tree
		{items: massItems, toDelete: dneItems, order: 3, shouldAlterTree: false},
		{items: massItems, toDelete: dneItems, order: 6, shouldAlterTree: false},
		// Delete should be able to fully delete trees of various
		// orders
		{items: massItems, toDelete: massItems, order: 3, shouldAlterTree: false},
		{items: massItems, toDelete: massItems, order: 6, shouldAlterTree: false},
		{items: massItems, toDelete: massItems, order: 18, shouldAlterTree: false},
		{items: massItems, toDelete: massItems, order: 30, shouldAlterTree: false},
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
				if presentBefore != nil || presentAfter == nil {
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
	massItems := uniqueInputsN(1000)
	emptyItems := uniqueInputsN(0)
	dneItems := []Item{&testItem{key: -999, val: 0}}
	cases := []struct {
		items      []Item
		order      int
		lookFor    []Item
		shouldFind bool
	}{
		// Should be able to search empty tree
		{items: emptyItems, order: 4, lookFor: massItems[0:2], shouldFind: false},
		// Should successfully not find missing items
		{items: massItems, order: 3, lookFor: dneItems, shouldFind: false},
		{items: massItems, order: 6, lookFor: dneItems, shouldFind: false},
		{items: massItems, order: 11, lookFor: dneItems, shouldFind: false},
		// Should successfully find items
		{items: massItems, order: 3, lookFor: massItems, shouldFind: true},
		{items: massItems, order: 6, lookFor: massItems, shouldFind: true},
		{items: massItems, order: 11, lookFor: massItems, shouldFind: true},
	}

	for _, c := range cases {
		b := New(c.order)
		for _, ti := range c.items {
			b.Insert(ti)
		}

		for _, target := range c.lookFor {
			res, err := b.Search(target)
			if (res == nil || err != nil) && c.shouldFind == true {
				t.Errorf("Should have found: %v\n", target)
			} else if (res != nil || err == nil) && c.shouldFind == false {
				t.Errorf("Should not have found: %v\n", target)
			}
		}
	}
}

func TestIteratorNext(t *testing.T) {
	massItems := uniqueInputsN(1000)
	oneItems := uniqueInputsN(1)
	emptyItems := uniqueInputsN(0)
	cases := []struct {
		items []Item
		order int
	}{
		// Should work for trees of various orders
		{items: massItems, order: 3},
		{items: massItems, order: 8},
		{items: massItems, order: 15},
		// Should work for empty trees
		{items: emptyItems, order: 3},
		{items: emptyItems, order: 4},
		// Should work for tree with single item
		{items: oneItems, order: 3},
		{items: oneItems, order: 4},
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
				t.Fatalf("Call to Next() should not have returned non-nil error")
			}
			if prev != nil && !prev.Less(next) {
				t.Fatalf("Values from Iterator should be ascending. Prev: %v, Next: %v", prev, next)
			}
			prev = next
		}

		for i := 0; i < len(c.items); i++ {
			if iter.HasNext() {
				t.Fatalf("Iterator should no longer have next")
			}
			extraIterVal, err := iter.Next()
			if extraIterVal != nil || err == nil {
				t.Fatalf("Extra call to Next() should have returned nil value and error."+
					"Instead got val: %v, and error: %v", extraIterVal, err)
			}
		}
	}
}

func TestIteratorReverseNext(t *testing.T) {
	massItems := uniqueInputsN(1000)
	oneItems := uniqueInputsN(1)
	emptyItems := uniqueInputsN(0)
	cases := []struct {
		items []Item
		order int
	}{
		// Should work for trees of various orders
		{items: massItems, order: 3},
		{items: massItems, order: 8},
		{items: massItems, order: 15},
		// Should work for empty trees
		{items: emptyItems, order: 3},
		{items: emptyItems, order: 4},
		// Should work for tree with single item
		{items: oneItems, order: 3},
		{items: oneItems, order: 4},
	}

	for _, c := range cases {
		b := New(c.order)
		for _, v := range c.items {
			b.Insert(v)
		}

		iter := b.NewReverseIterator()

		var prev Item
		for i := 0; i < len(c.items); i++ {
			next, err := iter.Next()
			if err != nil {
				t.Errorf("Call to Next() should not have returned non-nil error")
			}
			if prev != nil && !next.Less(prev) {
				t.Errorf("Values from Iterator should be descending. Prev: %v, Next: %v", prev, next)
			}
			prev = next
		}
		for i := 0; i < len(c.items); i++ {
			if iter.HasNext() {
				t.Errorf("Iterator should no longer have next")
			}
			extraIterVal, err := iter.Next()
			if extraIterVal != nil || err == nil {
				t.Errorf("Extra call to Next() should have returned nil value and error."+
					"Instead got val: %v, and error: %v", extraIterVal, err)
			}
		}
	}
}

func TestBulkload(t *testing.T) {
	massItems := uniqueInputsN(1000)
	cases := []struct {
		items []Item
		order int
	}{
		{items: massItems, order: 3},
		{items: massItems, order: 8},
		{items: massItems, order: 50},
		{items: massItems, order: 121},
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
	first := uniqueInputsN(2000)
	distinct := uniqueInputsN(2000)
	diffSize := uniqueInputsN(100)
	overlapsFirst := duplicateInputsN(5000)
	empty := uniqueInputsN(0)
	cases := []struct {
		first  []Item
		second []Item
		order  int
	}{
		// Should successfully merge with empty tree
		{first: first, second: empty, order: 6},
		{first: empty, second: first, order: 6},
		// Should successfully merge with distinct other tree
		{first: first, second: distinct, order: 3},
		{first: distinct, second: first, order: 3},
		// Should successfully merge with other tree with equal items
		{first: first, second: overlapsFirst, order: 5},
		{first: overlapsFirst, second: first, order: 5},
		// Should successfully merge with tree of different size
		{first: first, second: diffSize, order: 12},
		{first: diffSize, second: first, order: 12},
	}

	for _, c := range cases {
		firstTree := Bulkload(c.order, c.first)
		secondTree := Bulkload(c.order, c.second)

		mt, err := Merge(firstTree, secondTree)
		if err != nil || !isValidBTree(mt) {
			walk(mt.root)
			t.Errorf("Merged tree should have been valid")
		}
	}
}

//=============================================================================
//= Benchmarks
//=============================================================================

func benchmarkIterator(size, order int, b *testing.B) {
	massItems := uniqueInputsN(size)
	bt := Bulkload(order, massItems)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		iter := bt.NewIterator()
		iterateThrough(iter)
	}
}

func benchmarkIteratorReverse(size, order int, b *testing.B) {
	massItems := uniqueInputsN(size)
	bt := Bulkload(order, massItems)
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		iter := bt.NewReverseIterator()
		iterateThrough(iter)
	}
}

func iterateThrough(iter *Iterator) {
	for iter.HasNext() {
		iter.Next()
	}
}

func BenchmarkIterator10(b *testing.B)     { benchmarkIterator(10, 3, b) }
func BenchmarkIterator100(b *testing.B)    { benchmarkIterator(100, 3, b) }
func BenchmarkIterator1000(b *testing.B)   { benchmarkIterator(1000, 3, b) }
func BenchmarkIterator10000(b *testing.B)  { benchmarkIterator(10000, 3, b) }
func BenchmarkIterator100000(b *testing.B) { benchmarkIterator(100000, 3, b) }

func BenchmarkIteratorReverse10(b *testing.B)     { benchmarkIteratorReverse(10, 3, b) }
func BenchmarkIteratorReverse100(b *testing.B)    { benchmarkIteratorReverse(100, 3, b) }
func BenchmarkIteratorReverse1000(b *testing.B)   { benchmarkIteratorReverse(1000, 3, b) }
func BenchmarkIteratorReverse10000(b *testing.B)  { benchmarkIteratorReverse(10000, 3, b) }
func BenchmarkIteratorReverse100000(b *testing.B) { benchmarkIteratorReverse(100000, 3, b) }

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

// Return slice of *testItems with key/val between 0 and n.
// Values will be randomly ordered, as ranging over maps is random.
func uniqueInputsN(n int) []Item {
	numItems := n
	itemsMap := make(map[int]*testItem, n)
	for i := 0; i < numItems; i++ {
		itemsMap[i] = &testItem{i, i}
	}
	itemSlice := make([]Item, n)
	for i, v := range itemsMap {
		itemSlice[i] = v
	}
	return itemSlice
}

func duplicateInputsN(n int) []Item {
	itemSlice := uniqueInputsN(n)
	for i := n / 2; i < n/2; i++ {
		itemSlice[i] = itemSlice[0]
	}
	return itemSlice
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
