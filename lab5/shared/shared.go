package shared

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

const (
	MAX_NODES = 8
)

// Node struct represents a computing node.
type Node struct {
	ID        int
	Hbcounter int
	Time      float64
	Alive     bool
}

// Generate random crash time from 10-60 seconds
func (n Node) CrashTime() int {
	rand.Seed(time.Now().UnixNano())
	max := 60
	min := 15
	return rand.Intn(max-min) + min
}

func (n Node) InitializeNeighbors(id int) [2]int {
	neighbor1 := RandInt()
	for neighbor1 == id {
		neighbor1 = RandInt()
	}
	neighbor2 := RandInt()
	for neighbor1 == neighbor2 || neighbor2 == id {
		neighbor2 = RandInt()
	}
	return [2]int{neighbor1, neighbor2}
}

func RandInt() int { //fix outdated rand seed
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(MAX_NODES-1+1) + 1
}

/*---------------*/

// Membership struct represents participanting nodes
type Membership struct {
	Members map[int]Node //Members variable represents mapped key value pairs, with key int (which node it came from), and value Node
}

// Returns a new instance of a Membership (pointer).
func NewMembership() *Membership {
	return &Membership{
		Members: make(map[int]Node),
	}
}

// Adds a node to the membership list.
func (m *Membership) Add(payload Node, reply *Node) error {
	_, exists := m.Members[payload.ID]
	//check if we already have node with that id in the membership table
	if exists {
		return nil
	} else { //add using key of node id, and node data as value
		m.Members[payload.ID] = payload
		*reply = payload //send back node that was added to membership table
	}

	return nil
}

// Updates a node in the membership list.
func (m *Membership) Update(payload Node, reply *Node) error {
	//payload is copy of node so can change time and put into membership table entry
	payload.Time = float64(time.Now().UnixNano()) / float64(time.Second)
	m.Members[payload.ID] = payload

	return nil
}

// Returns a node with specific ID.
func (m *Membership) Get(id int, reply *Node) error {
	node, exists := m.Members[id]
	if !exists {
		fmt.Printf("Node %d not in Membership Table\n", id)
		return nil
	}

	*reply = node
	return nil
}

// Request struct represents a new message request to a client
type Request struct {
	ID    int        //id of destination node (client)
	Table Membership //gossip table sending to client
}

// Requests struct represents pending message requests
type Requests struct {
	Pending map[int]Membership
	mu      sync.Mutex
}

// Returns a new instance of a Requests (pointer).
func NewRequests() *Requests {
	return &Requests{
		Pending: make(map[int]Membership),
	}
}

// Adds a new message request to the pending list
func (req *Requests) Add(request Request, added *bool) error {
	req.mu.Lock() //mutex lock for race condition w/multiple nodes accessing at same time

	old_table, exists := req.Pending[request.ID]

	if !exists { //first thing to get to requests mailbox
		req.Pending[request.ID] = request.Table
	} else { //if something already there combine tables for one bigger membership table
		req.Pending[request.ID] = *CombineTables(&old_table, &request.Table)
	}
	*added = true
	req.mu.Unlock()
	return nil
}

// Listens to communication from neighboring nodes.
func (req *Requests) Listen(ID int, rec_table *Membership) error {
	table, exists := req.Pending[ID]
	if !exists { //if no table at that ID exists in request table, then no gossip
		*rec_table = *NewMembership() //return empty membership table
		return nil
	}

	*rec_table = table
	return nil
}

func CombineTables(recieving *Membership, sending *Membership) *Membership {
	for id, sending_node := range sending.Members {
		rec_node, exists := recieving.Members[id]
		var reply Node
		if exists {
			if sending_node.Hbcounter > rec_node.Hbcounter { //if heartbeat higher, update
				recieving.Update(sending_node, &reply)
			}
		} else { //if doesn't exist, add to membership table
			recieving.Add(sending_node, &reply)
		}
	}

	return recieving //return the updated membership table after recieving all changes
}
