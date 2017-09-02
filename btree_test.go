package btree

import "testing"

func walk(tree *Btree) {

}

// checkBtreeInvariants checks that given tree satisfies the definition of a
// B-tree.
// This function should be used at the end of each test.
func checkBtreeInvariants(tree *Btree, order m) {
	// For Btree of order m:
	// 1. Every node has at most m children
	// 2. Every non-leaf node (except root) has at least [m/2] children
	// 3. The root has at least two children if it is not a leaf
	// 4. A non-leaf node with k children contains k-1 keys
	// 5. All leaves appear in the same level

}

func TestInsert(t *testing.T) {

}

func TestDelete(t *testing.T) {

}

func TestSearch(t *testing.T) {

}

func TestBulkload(t *testing.T) {

}
