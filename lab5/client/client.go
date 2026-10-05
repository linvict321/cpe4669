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

// Send the current membership table to a neighboring node (to their ID)
func sendMessage(server *rpc.Client, id int, membership shared.Membership) {
	var req = shared.Request{ID: id, Table: membership}
	var self_req_response bool
	if err := server.Call("Requests.Add", req, &self_req_response); err != nil {
		fmt.Println("Error: Requests.Add()", err)
	} else {
		fmt.Printf("Success: Message added to request list for Node %d\n", id)
	}
}

// Read incoming messages from other nodes
func readMessages(server *rpc.Client, id int) *shared.Membership {
	var self_recieved_table *shared.Membership
	//use own id to check requests table if you have any message for your node id
	if err := server.Call("Requests.Listen", id, &self_recieved_table); err != nil {
		fmt.Println("Error: Requests.Listen()", err)
		return nil //have to check this when calling readMessages()
	} else {
		fmt.Printf("Success: Node %d recieved pending gossip\n", id)
		return self_recieved_table
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
		fmt.Println("Error: Membership.Add()", err)
	} else {
		fmt.Printf("Success: Node created with id= %d\n", id)
	}

	neighbors := self_node.InitializeNeighbors(id)
	fmt.Println("Neighbors:", neighbors)

	membership := shared.NewMembership()
	membership.Add(self_node, &self_node)

	sendMessage(server, neighbors[0], *membership)

	// crashTime := self_node.CrashTime()

	time.AfterFunc(time.Second*X_TIME, func() { runAfterX(server, &self_node, &membership, id) })
	time.AfterFunc(time.Second*Y_TIME, func() { runAfterY(server, neighbors, &membership, id) })
	time.AfterFunc(time.Second*time.Duration(Z_TIME), func() { runAfterZ(server, id) })

	wg.Add(1)
	wg.Wait()
}

func runAfterX(server *rpc.Client, node *shared.Node, membership **shared.Membership, id int) {
	//TODO
}

func runAfterY(server *rpc.Client, neighbors [2]int, membership **shared.Membership, id int) {
	//TODO
	//pick neighbor to send to
	//sendMessage(server, )
	//also recieve message, need to send and read message each gossip round
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
