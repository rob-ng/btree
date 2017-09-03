package btree

import (
	"fmt"
	"testing"
)

func walk(tree *BTree) {

}

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

func atLeastChildren(curr *node, min int) bool {
	if len(curr.children) == 0 {
		return true
	}

	if len(curr.children) < min {
		return false
	}
	for _, c := range curr.children {
		if !atLeastChildren(c, min) {
			return false
		}
	}
	return true

}

func atLeastChildrenRoot(root *node) bool {
	if len(root.children) == 0 {
		return true
	}
	if len(root.children) >= 2 {
		return true
	}
	return false
}

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

func allLeavesSameDepth(root *node) bool {
	expectedDepth := 0
	curr := root
	for len(curr.children) > 0 {
		expectedDepth++
		curr = curr.children[0]
	}

	return allLeavesSameDepthRecurse(root, 0, expectedDepth)
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

	return true
}

func TestInsert(t *testing.T) {
	b := NewBTree(1)
	fmt.Printf("%v\n", isValidBTree(b))
}

func TestDelete(t *testing.T) {

}

func TestSearch(t *testing.T) {

}

func TestBulkload(t *testing.T) {

}
