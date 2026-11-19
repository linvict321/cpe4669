package main

import (
	"errors"
	"fmt"
	"hash/fnv"
	"log"
	"net/http"
	"net/rpc"
	"os"
	"slices"
	"sort"
	"strconv"
	"sync"
	"time"
)

//coordinator handles looking at temp node (sloppy quorum and hinted handoff) to get info back to revived node, the node itself won't know

//what should we set our fail time, and x_time, y_time, variables to?
//note: x_time = increment heartbeat, y_time = send membership, fail_time = if not manually killing nodes can set time to kill

//gossip/membership protocol structs
type Membership struct {
	id int
	hbCount int64
	time time.Time
	alive bool
}

type GossipArgs struct { //sending request
	Members map[int]Membership
}

type GossipReply struct {
}

type Node struct {
	id              int
	totalNodes      int // includes dead peers
	port            int
	peers           map[int]*rpc.Client
	//might need mutex for membership table
	membership map[int]Membership //membership table
}

//
// Common RPC request/reply definitions
//
type PutArgs struct {
	Key string
	Value string
}

type PutReply struct {
}

type GetArgs struct {
	Key string
}

type GetReply struct {
	Value string
}

type CliArgs struct {
	id, N int
}

const BASE_PORT int = 18080
const MAX_RETRIES int = 10
const NUM_VIRTUAL_NODES_PER_PHYS_NODE int = 4

func ParseArgs() CliArgs {
	argv := os.Args[1:]
	// Get id from command line argument
	if len(argv) != 2 {
		log.Fatal("Usage: go run raft.go <ID> <N>")
	}
	id, err := strconv.Atoi(argv[0])
	if err != nil {
		log.Fatal("Found Error", err)
	} else if id < 0 || id > 999 {
		log.Fatal("id must an int between [0, 999], but was ", id)
	}
	N, err := strconv.Atoi(argv[1])
	if err != nil {
		log.Fatal("Found Error", err)
	} else if N < 1 || N > 999{
		log.Fatal("N must be within [0, 999] but was ", N)
	}

	if id >= N {
		log.Fatal("id (", id, ") is errneously greater than N (", N, ")")
	}

	return CliArgs{id: id, N: N}
}

func initNode(args CliArgs) *Node { // Go garbage collector means don't need to worry about dangling pointer like in C
	time := time.Now()
	node := new(Node)
	node.id = args.id
	node.port = BASE_PORT + args.id
	node.totalNodes = args.N
	node.peers = make(map[int]*rpc.Client, args.N)

	node.membership = make(map[int]Membership, args.N)
	node.membership[node.id] = Membership{ //init self membership
		id: node.id, hbCount: 0, time: time, alive: true,
	}

	return node
}

//heartbeat/gossip funcs
func(node *Node) incrementHb() {
	self := node.membership[node.id]
	self.hbCount++
	self.time = time.Now()
	self.alive = true
	node.membership[node.id] = self
}

func (node *Node) combineTables(incoming map[int]Membership) {
	for id, receiving := range incoming { //iterate through node ids in membership table
		old, exists := node.membership[id] //check if have node in membership table
		if exists {
			if receiving.hbCount > old.hbCount { //update membership entry
				receiving.time = time.Now()
				node.membership[id] = receiving
			}
		} else { //add to membership table
			receiving.time = time.Now()
			node.membership[id] = receiving
		}
	}
}

//gossip round func to send to two random neighbors (if same as last lab)

//actual gossipRPC function to connect to the client and send
func GossipRPC(peerId int, args *GossipArgs, reply *GossipReply) {
/* From Lab5:
var req = shared.Request{ID: id, Table: membership}
	var self_req_response bool
	if err := server.Call("Requests.Add", req, &self_req_response); err != nil {
		fmt.Println("Error: Requests.Add()", err)
	} else {
		//fmt.Printf("Success: Message added to request list for Node %d\n", id)//commented out for cleaner debugging screen
	}
*/
}

//checkFailures() from lab5 should be pretty similar
func(node *Node) CheckFailures() {
	for id, member := range node.membership {
		//need fail_timeout variables
		//if member.Alive && time.Since(member.Time) > FAIL_TIMEOUT {
		//}
		id = id
		member = member
	}
}



func(node *Node) PrintMsg(msg string) {
	fmt.Println("[ NODE", node.id, "] -", msg)
}

//
// Client
//

func(node *Node) get(key string, c *ConsistentHash) string {
	node_id_with_key, err := c.Get(key)
	if err != nil {
		log.Fatal("error:", err)
	}

	args := GetArgs{ key }
	reply := GetReply{}
	err = node.peers[node_id_with_key].Call("KV.Get", &args, &reply)
	if err != nil {
		log.Fatal("error:", err)
	}
	return reply.Value
}

func(node *Node) put(key string, val string, c *ConsistentHash) {
	node_id_for_key, err := c.Get(key)
	
	args := PutArgs{ key, val }
	reply := PutReply{}
	err = node.peers[node_id_for_key].Call("KV.Put", &args, &reply)
	if err != nil {
		log.Fatal("error:", err)
	}
}

//
// Server
//
//type VectorClock map[string]uint64 //compare for updates; (node, counter) pair

type KV struct {
	mu sync.Mutex
	//data map[string]string
	store *Store
}

func (kv *KV) Get(args *GetArgs, reply *GetReply) error {
	kv.mu.Lock()
	defer kv.mu.Unlock()

	versions := kv.store.GetVersions(args.Key)
	if len(versions) < 0 {
		return errors.New(fmt.Sprintf("Key '%s' not found!", args.Key))
	}

	reply.Value = versions[len(versions)-1].Value
	return nil
}

func (kv *KV) Put(args *PutArgs, reply *PutReply) error {
	kv.mu.Lock()
	defer kv.mu.Unlock()

	kv.store.PutVersion(args.Key, args.Value, nil) //TODO: how to input the vector clock for this object?
	return nil
}

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


func main() {
	args := ParseArgs()
	node := initNode(args)
	kv := new(KV)
	kv.store = NewStore()

	rpc.Register(node) // register node with `rpc.DefaultServer`
	rpc.Register(kv)
	rpc.HandleHTTP() // register an HTTP handler for RPC communication

	// listen and serve requests to the default HTTP server
	addr := "localhost:" + strconv.Itoa(node.port)
	wg := new(sync.WaitGroup)
	wg.Go(func() {
		err := http.ListenAndServe(addr, nil)
		if err != nil {
			log.Fatalf("Node %d listen error: %v", node.id, err)
		}
	})
	time.Sleep(time.Second)

	// now connect to the other nodes
	for i := range args.N {
		peerAddr := "localhost:" + strconv.Itoa(BASE_PORT + i)
		
		var peer *rpc.Client
		var err error
		for range MAX_RETRIES { // continuous for loop for retrying
			peer, err = rpc.DialHTTP("tcp", peerAddr)
			if err == nil {
				break
			}
			// fmt.Println("Waiting for node ", i, "'s server to setup")
			time.Sleep(200 * time.Millisecond)
		}
		if err != nil {
			node.PrintMsg(fmt.Sprintf("Failed to connect to node %d: %v", i, err))
			continue
		}
		node.peers[i] = peer
		msg := fmt.Sprintf("Connected to node %d\n", i)
		node.PrintMsg(msg)
	}
	time.Sleep(time.Duration(3) * time.Second)

	// set up the consistent hash ring
	consistentHash := NewConsistentHash(args.N, NUM_VIRTUAL_NODES_PER_PHYS_NODE)
	for i := range args.N {
		consistentHash.Add(i)
	}

	// node.PrintMsg(fmt.Sprintf("ConsistentHash: %d", consistentHash.VirtualNodesNum))
	// if node.id == 0 {
	// 	consistentHash.Print()
	// }

	go func() {
		for {
			time.Sleep(1000)
		}
	}()

	wg.Wait() // keep at end of main to keep the nodes running forever
}