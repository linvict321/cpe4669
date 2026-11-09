package main

import (
	"errors"
	"fmt"
	"hash"
	"hash/fnv"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"slices"
	"sort"
	"strconv"
	"sync"
	"time"
)

type Node struct {
	id              int
	totalNodes      int // includes dead peers
	port            int
	peers           map[int]*rpc.Client
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
	node := new(Node)
	node.id = args.id
	node.port = BASE_PORT + args.id
	node.totalNodes = args.N
	node.peers = make(map[int]*rpc.Client, args.N-1)

	return node
}

func(node *Node) PrintMsg(msg string) {
	fmt.Println("[ NODE", node.id, "] -", msg)
}

//
// Client
//
func connect() *rpc.Client {
	client, err := rpc.Dial("tcp", ":1234")
	if err != nil {
		log.Fatal("dialing:", err)
	}
	return client
}

func get(Key string) string {
	client := connect()
	args := GetArgs{ key }
	reply := GetReply{}
	err := client.Call("KV.Get", &args, &reply)
	if err != nil {
		log.Fatal("error:", err)
	}
	client.Close()
	return reply.Value
}

func put(key string, val string) {
	client := connect()
	args := PutArgs{ key, val }
	reply := PutReply{}
	err := client.Call("KV.Put", &args, &reply)
	if err != nil {
		log.Fatal("error:", err)
	}
	client.Close()
}

//
// Server
//
//type VectorClock map[string]uint64 //compare for updates; (node, counter) pair

//sice of uint32 ints representing hashed key of nodes
type SortedKeys []uint64 
func (x SortedKeys) Len() int {	return len(x) }
func (x SortedKeys) Less(i, j int) bool{	return x[i] < x[j]}
func (x SortedKeys) Swap(i, j int) {x[i], x[j] = x[j], x[i]}
type ConsistentHash struct {
	VirtualNodesNum int //num of virtual nodes
	hashSortedKeys SortedKeys //sorted slice of hash values
	circleRing map[uint64]string //map where key = hash, val = node
	dataset map[string]bool //set of real nodes
}
type KV struct {
	mu sync.Mutex
	data map[string]string
}

func server() {
	kv := &KV{data: map[string]string{}}
	ch := &ConsistentHash{VirtualNodesNum: 0, hashSortedKeys: SortedKeys{}, 
			circleRing: map[uint64]string{}, dataset: map[string]bool{}}
	rpcs := rpc.NewServer()
	rpcs.Register(kv)
	rpcs.Register(ch)
	l, e := net.Listen("tcp", ":1234")
	if e != nil {
		log.Fatal("listen error:", e)
	}
	go func() {
		for {
			conn, err := l.Accept()
			if err == nil {
				go rpcs.ServeConn(conn)
			} else {
				break
			}
		}
		l.Close()
	}()
}

func (kv *KV) Get(args *GetArgs, reply *GetReply) error {
	kv.mu.Lock()
	defer kv.mu.Unlock()
	reply.Value = kv.data[args.Key]
	return nil
}

func (kv *KV) Put(args *PutArgs, reply *PutReply) error {
	kv.mu.Lock()
	defer kv.mu.Unlock()
	kv.data[args.Key] = args.Value
	return nil
}


//consistent hashing functions
func(c *ConsistentHash) hash(key string) uint64{
	hasher := fnv.New64a()
	hasher.Write([]byte(key))
	return hasher.Sum64()

}
func(c *ConsistentHash) Add(node string){

	c.dataset[node] = true
	key := c.hash(node) //hash value i
	c.circleRing[key] = node

	for i:= 0; i < c.VirtualNodesNum; i++ {
		virtual_key := strconv.Itoa(i) + "-" + node
		virtual_hash := c.hash(virtual_key)
		c.circleRing[virtual_hash] = node
	}
	//insort but Go's version
	slices.Sort(c.hashSortedKeys)
}

//gets the first virtual node that is clockwise
func(c *ConsistentHash) Get(node string) (string, error){
	if len(c.dataset) == 0{
		return "", errors.New("empty struct")
	}

	targetH := c.hash(node)
	idx := sort.Search(len(c.hashSortedKeys), 
		func(i int) bool{ return c.hashSortedKeys[i] >= targetH})

	if idx >= len(c.hashSortedKeys) { //if its at the end of ring
		idx = 0
	}
	//uses binary search (sort.Search) to find the closest node

	nearestNode, _ := c.circleRing[c.hashSortedKeys[idx]]
	
	return nearestNode, nil
}
func(c *ConsistentHash) Remove(node string){

	//if not in dataset
	if _, find c.dataset[node]; !find{
		return
	}

	delete(c.dataset, node)
	key = c.hash(node)
	delete(c.circleRing, key)

	//delete the virtual nodes
	for i:= 0; i < c.VirtualNodesNum; i++{
		vk := c.getVirtualNodeKey(i, node)
		delete(c.circleRing, vk)
	}
	slices.Sort(c.hashSortedKeys) //update list again
}


//
// main
//
func main() {
	// server()
	
	// put("subject", "6.5840")
	// fmt.Printf("Put(subject, 6.5840) done\n")
	// fmt.Printf("get(subject) -> %s\n", get("subject"))

	// data := "test data"
	// hash := md5.New()
	// n, _ := hash.Write([]byte(data))

	// fmt.Printf("Hashed is type %T and value %x\n", hash.Sum(nil), hash.Sum(nil))
	// fmt.Println(n, hash.Sum(nil))


	args := ParseArgs()
	node := initNode(args)

	rpc.Register(node) // register node with `rpc.DefaultServer`
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
		if i == args.id {
			continue // don't connect to self
		}
		peerAddr := "localhost:" + strconv.Itoa(BASE_PORT + i)
		
		var peer *rpc.Client
		var err error
		for i := range MAX_RETRIES { // continuous for loop for retrying
			peer, err = rpc.DialHTTP("tcp", peerAddr)
			if err == nil {
				break
			}
			// fmt.Println("Waiting for node ", i, "'s server to setup")
			i = i
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