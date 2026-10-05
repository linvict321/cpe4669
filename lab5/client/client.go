package main

import (
	"fmt"
	"lab5/shared"
	"math/rand"
	"net/rpc"
	"os"
	"strconv"
	"sync"
	"time"
)

const (
	MAX_NODES  = 8
	X_TIME     = 1   //increments heartbeat counter every x seconds
	Y_TIME     = 2   //sende membership table every y seconds
	Z_TIME_MAX = 100 //z = time for node to fail, log(N) = log(8) = 3
	Z_TIME_MIN = 10
)

var self_node shared.Node

// Send the current membership table to a neighboring node with the provided ID
func sendMessage(server rpc.Client, id int, membership shared.Membership) {
	request := shared.Request{
		ID: id,
		Table: membership,
	}
	var reply bool
	err := server.Call("Requests.Add", request, &reply)
	if err != nil {
		fmt.Printf("Node %d failed to send to Node %d: %v\n", self_node.ID, id, err)
		return
	}

	fmt.Printf("Node %d sent table to Node %d\n", self_node.ID, id)

}

// Read incoming messages from other nodes
func readMessages(server rpc.Client, id int, membership shared.Membership) *shared.Membership {
	//look in our mailbox
	var receivedTable shared.Membership
	err := server.Call("Requests.Listen", id, &receivedTable)
	if err != nil {
		fmt.Printf("Node %d read error: %v\n", id, err)
		return &membership
	}

	if len(receivedTable.Members) == 0 { //if Requets.Listen is empty, return our existing table, no gossip arrived
		return &membership
	}

	fmt.Printf("Node %d received table with %d entries\n", id, len(receivedTable.Members))

	//merge the received table
	for nodeID, recNode := range receivedTable.Members {
		myNode, exists := membership.Members[nodeID]
		if !exists || recNode.Hbcounter > myNode.Hbcounter {
			membership.Members[nodeID] = recNode
		}
	}
	return &membership
}

func checkFailures(membership **shared.Membership, id int) {
	//a node is failed if heartbeat hasn't been updated within timeout
	currTime := float64(time.Now().UnixNano()) / float64(time.Second)
	failTime := float64(3 * X_TIME)

	for nodeID, node := range (*membership).Members {
		if nodeID == id {
			continue
		}
		if node.Alive && (currTime - node.Time) > failTime {
			node.Alive = false
			(*membership).Members[nodeID] = node
			fmt.Printf("Node %d detected failure of Node %d\n", id, nodeID)
		}
	}
}

// func calcTime() float64 { //don't think i'll use this
// 	//TODO
// }

var wg = &sync.WaitGroup{}

func main() {
	rand.Seed(time.Now().UnixNano())
	Z_TIME := rand.Intn(Z_TIME_MAX-Z_TIME_MIN) + Z_TIME_MIN

	// Connect to RPC server
	server, _ := rpc.DialHTTP("tcp", "localhost:9005")

	args := os.Args[1:]

	// Get ID from command line argument
	if len(args) == 0 {
		fmt.Println("No args given")
		return
	}
	id, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Println("Found Error", err)
	}

	fmt.Println("Node", id, "will fail after", Z_TIME, "seconds")

	//currTime := calcTime()
	currTime := float64(time.Now().UnixNano()) / float64(time.Second) //convert from ns to sec, and typecast to float

	// Construct self
	self_node = shared.Node{ID: id, Hbcounter: 0, Time: currTime, Alive: true}
	var self_node_response shared.Node // Allocate space for a response to overwrite this

	// Add node with input ID
	if err := server.Call("Membership.Add", self_node, &self_node_response); err != nil {
		fmt.Println("Error:2 Membership.Add()", err)
	} else {
		fmt.Printf("Success: Node created with id= %d\n", id)
	}

	neighbors := self_node.InitializeNeighbors(id)
	fmt.Println("Neighbors:", neighbors)

	membership := shared.NewMembership()
	membership.Add(self_node, &self_node)

	sendMessage(*server, neighbors[0], *membership)

	// crashTime := self_node.CrashTime()

	time.AfterFunc(time.Second*X_TIME, func() { runAfterX(server, &self_node, &membership, id) })
	time.AfterFunc(time.Second*Y_TIME, func() { runAfterY(server, neighbors, &membership, id) })
	time.AfterFunc(time.Second*time.Duration(Z_TIME), func() { runAfterZ(server, id) })

	wg.Add(1)
	wg.Wait()
}

func runAfterX(server *rpc.Client, node *shared.Node, membership **shared.Membership, id int) {
	//if a node stops incrementing its heartbeat other nodes call it dead
	//runAfterX helps with failure detection, slide 21
	
	//increment heartbeat every x seconds
	node.Hbcounter++
	node.Time = float64(time.Now().UnixNano())/float64(time.Second)
	node.Alive = true

	//update in local membership before gossiping
	(*membership).Members[id] = *node
	fmt.Printf("Node %d: heartbeat incremented to %d\n", id, node.Hbcounter)

	//reschedule for next x seconds
	time.AfterFunc(time.Second*X_TIME, func() { runAfterX(server, node, membership, id)})
}

func runAfterY(server *rpc.Client, neighbors [2]int, membership **shared.Membership, id int) {
	//TODO
}

func runAfterZ(server *rpc.Client, id int) {
	//TODO
}

func printMembership(m shared.Membership) {
	for _, val := range m.Members {
		status := "is Alive"
		if !val.Alive {
			status = "is Dead"
		}
		fmt.Printf("Node %d has hb %d, time %.1f and %s\n", val.ID, val.Hbcounter, val.Time, status)
	}
	fmt.Println("")
}
