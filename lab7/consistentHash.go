package main

import (
	"errors"
	"hash/fnv"
	"slices"
	"sort"
	"strconv"
	"sync"
)

//slice of uint64 ints representing hashed key of nodes
type SortedKeys []uint64 
func (x SortedKeys) Len() int {	return len(x) }
func (x SortedKeys) Less(i, j int) bool{	return x[i] < x[j]}
func (x SortedKeys) Swap(i, j int) {x[i], x[j] = x[j], x[i]}
type ConsistentHash struct {
	mu sync.Mutex
	VirtualNodesNum int //num of virtual nodes
	hashSortedKeys SortedKeys //sorted slice of hash values
	circleRing map[uint64]int //map where key = hash, val = node id
	dataset map[int]bool //set of real nodes
}

func NewConsistentHash(numPhysNodes int, numVirtualsPerPhys int) *ConsistentHash {
	c := new(ConsistentHash)
	c.VirtualNodesNum = numVirtualsPerPhys
	c.hashSortedKeys = make(SortedKeys, numVirtualsPerPhys * numPhysNodes)
	c.circleRing = make(map[uint64]int)
	c.dataset = make(map[int]bool)
	return c
}

func cHash(key string) uint64{
	hasher := fnv.New64a()
	hasher.Write([]byte(key))
	return hasher.Sum64()
}

//consistent hashing functions
func(c *ConsistentHash) Add(node int){
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.dataset[node]{
		return
	}

	c.dataset[node] = true
	// key := cHash(strconv.Itoa(node)) //hash value i
	// c.circleRing[key] = node

	for i := range c.VirtualNodesNum {
		virtual_key := strconv.Itoa(i) + "-" + strconv.Itoa(node)
		virtual_hash := cHash(virtual_key)
		c.circleRing[virtual_hash] = node
		c.hashSortedKeys = append(c.hashSortedKeys, virtual_hash)
	}
	//insort but Go's version
	slices.Sort(c.hashSortedKeys)
}

func(c *ConsistentHash) Print() {
	for hash, node_id := range c.circleRing {
		println("hash:", hash, "node id:", node_id)
	}
}

//gets the first virtual node that is clockwise
func(c *ConsistentHash) Get(node string) (int, error){
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.dataset) == 0{
		return -1, errors.New("empty struct")
	}

	targetH := cHash(node)
	idx := sort.Search(len(c.hashSortedKeys), 
		func(i int) bool{ return c.hashSortedKeys[i] >= targetH})

	if idx >= len(c.hashSortedKeys) { //if its at the end of ring
		idx = 0
	}
	//uses binary search (sort.Search) to find the closest node

	nearestNode, _ := c.circleRing[c.hashSortedKeys[idx]]
	
	return nearestNode, nil
}
func(c *ConsistentHash) Remove(node int){
	c.mu.Lock()
	defer c.mu.Unlock()
	//if not in dataset
	if _, find := c.dataset[node]; !find{
		return
	}

	delete(c.dataset, node)
	key := cHash(strconv.Itoa(node))
	delete(c.circleRing, key)

	//delete the virtual nodes
	for i:= 0; i < c.VirtualNodesNum; i++{
		delete(c.circleRing, cHash(strconv.Itoa(i) + "-" + strconv.Itoa(node)))
	}
	slices.Sort(c.hashSortedKeys) //update list again
}