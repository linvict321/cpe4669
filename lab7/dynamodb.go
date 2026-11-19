package main

import (
	"fmt"
	"slices"
	"sort"

	// "hash/fnv"
	"log"
	"net/http"
	"net/rpc"
	"os"

	// "slices"
	// "sort"
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
	consistentHash *ConsistentHash
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
const NUM_REPLICAS int = 3

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

	// set up the consistent hash ring
	node.consistentHash = NewConsistentHash(args.N, NUM_VIRTUAL_NODES_PER_PHYS_NODE)
	for i := range args.N {
		node.consistentHash.Add(i)
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

func(node *Node) Get(req *GetArgs, reply *GetReply) error {
	node_id_with_key, err := node.consistentHash.Get(req.Key)
	if err != nil {
		return fmt.Errorf("Key '%s' not found!", req.Key)
	}

	err = node.peers[node_id_with_key].Call("KV.Get", req, reply)
	if err != nil {
		return fmt.Errorf("Key '%s' not found!", req.Key)
	}
	return nil
}

func(node *Node) Put(req *PutArgs, reply *PutReply) error {
	c := node.consistentHash // for brevity
	node_id_for_key, err := c.Get(req.Key)
	
	err = node.peers[node_id_for_key].Call("KV.Put", req, reply)
	if err != nil {
		log.Fatal("error:", err)
	}

	// put in the 2 replicas as well (into different physical nodes)
	targetH := cHash(req.Key)
	idx := sort.Search(len(c.hashSortedKeys), 
		func(i int) bool{ return c.hashSortedKeys[i] >= targetH})
	if idx >= len(c.hashSortedKeys) { //if its at the end of ring
		idx = 0
	}

	physNodeIDsKeyPutIn := []int{node_id_for_key}
	for range (NUM_REPLICAS-1) {
		idx = (idx + 1) % len(c.hashSortedKeys)
		for slices.Contains(physNodeIDsKeyPutIn, c.circleRing[c.hashSortedKeys[idx]]) {
			idx = (idx + 1) % len(c.hashSortedKeys)
		}
		replicaNodeId, _ := c.circleRing[c.hashSortedKeys[idx]]
		err = node.peers[replicaNodeId].Call("KV.Put", req, reply)
		if err != nil {
			log.Fatal("error:", err)
		}
	}

	return nil
}

//
// Server
//
//type VectorClock map[string]uint64 //compare for updates; (node, counter) pair

type KV struct {
	mu sync.Mutex
	nodeId int
	store *Store
}

func (kv *KV) Get(args *GetArgs, reply *GetReply) error {
	kv.mu.Lock()
	defer kv.mu.Unlock()

	versions := kv.store.GetVersions(args.Key)
	if len(versions) <= 0 {
		return fmt.Errorf("Key '%s' not found!", args.Key)
	}

	//TODO: do the version reconciliation before returning the value
	reply.Value = versions[len(versions)-1].Value
	return nil
}

func (kv *KV) Put(args *PutArgs, reply *PutReply) error {
	kv.mu.Lock()
	defer kv.mu.Unlock()

	//TODO: how to input the vector clock for this object?
	kv.store.PutVersion(args.Key, args.Value, nil)
	fmt.Println("[ NODE", kv.nodeId, "] -", "put", args.Key, ":", args.Value)
	return nil
}

func main() {
	args := ParseArgs()
	node := initNode(args)
	kv := new(KV)
	kv.store = NewStore()
	kv.nodeId = args.id

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

	wg.Wait() // keep at end of main to keep the nodes running forever
}