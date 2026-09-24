package main

import "sort"

const order = 3 // max children per node

type BTree struct {
	root *bNode
}

type bNode struct {
	keys     []int
	values   []string
	children []*bNode
	leaf     bool
}

func NewBTree() *BTree {
	return &BTree{root: &bNode{leaf: true}}
}

func (t *BTree) Insert(key int, val string) {
	r := t.root
	if len(r.keys) == 2*order-1 {
		s := &bNode{}
		t.root = s
		s.children = append(s.children, r)
		t.splitChild(s, 0)
		t.insertNonFull(s, key, val)
	} else {
		t.insertNonFull(r, key, val)
	}
}

func (t *BTree) insertNonFull(n *bNode, key int, val string) {
	i := len(n.keys) - 1
	if n.leaf {
		n.keys = append(n.keys, 0)
		n.values = append(n.values, "")
		for i >= 0 && key < n.keys[i] {
			n.keys[i+1] = n.keys[i]
			n.values[i+1] = n.values[i]
			i--
		}
		n.keys[i+1] = key
		n.values[i+1] = val
	} else {
		for i >= 0 && key < n.keys[i] {
			i--
		}
		i++
		if len(n.children[i].keys) == 2*order-1 {
			t.splitChild(n, i)
			if key > n.keys[i] {
				i++
			}
		}
		t.insertNonFull(n.children[i], key, val)
	}
}

func (t *BTree) splitChild(parent *bNode, i int) {
	child := parent.children[i]
	mid := order - 1

	newNode := &bNode{leaf: child.leaf}
	newNode.keys = append(newNode.keys, child.keys[mid+1:]...)
	newNode.values = append(newNode.values, child.values[mid+1:]...)

	if !child.leaf {
		newNode.children = append(newNode.children, child.children[mid+1:]...)
		child.children = child.children[:mid+1]
	}

	parent.keys = append(parent.keys, 0)
	parent.values = append(parent.values, "")
	copy(parent.keys[i+1:], parent.keys[i:])
	copy(parent.values[i+1:], parent.values[i:])
	parent.keys[i] = child.keys[mid]
	parent.values[i] = child.values[mid]

	parent.children = append(parent.children, nil)
	copy(parent.children[i+2:], parent.children[i+1:])
	parent.children[i+1] = newNode

	child.keys = child.keys[:mid]
	child.values = child.values[:mid]
}

func (t *BTree) Search(key int) (string, bool) {
	return searchNode(t.root, key)
}

func searchNode(n *bNode, key int) (string, bool) {
	i := sort.SearchInts(n.keys, key)
	if i < len(n.keys) && n.keys[i] == key {
		return n.values[i], true
	}
	if n.leaf {
		return "", false
	}
	return searchNode(n.children[i], key)
}

// RangeSearch returns all values with keys in [lo, hi].
func (t *BTree) RangeSearch(lo, hi int) []string {
	var results []string
	rangeNode(t.root, lo, hi, &results)
	return results
}

func rangeNode(n *bNode, lo, hi int, results *[]string) {
	i := 0
	for i < len(n.keys) {
		if !n.leaf {
			rangeNode(n.children[i], lo, hi, results)
		}
		if n.keys[i] >= lo && n.keys[i] <= hi {
			*results = append(*results, n.values[i])
		}
		i++
	}
	if !n.leaf {
		rangeNode(n.children[i], lo, hi, results)
	}
}
