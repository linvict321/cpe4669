package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"time"

	// "math"
	"math/rand/v2"
	"net/http"
	"net/rpc"
	"strconv"
	// "sync"
	// "time"
)

type CliArgs struct {
	id, N int
}

type State int
const ( // like an enum for the states
	Follower  State = iota // 0
	Candidate              // 1
	Leader                 // 2
)

type Node struct {
	// from gossip, might not need these three fields
	Hbcounter int
	Time      float64
	Alive     bool

	id              int
	currentTerm     int
	state           State
	votedFor        int
	voteCount       int // used if a candidate
	lastHb          int // in milliseconds
	ElectionTimeout int // in milliseconds, will be between [150, 300] ms
}

type HbArgs struct {
	term     int // leader's currentTerm
	leaderId int
}

type ReqVoteArgs struct {
	term        int // candidate's term
	candidateId int // candidate requesting vote
}

const BASE_PORT int = 9000
const MAX_RETRIES int = 10

func (n *Node) SetNewElectionTimeout() {
	n.ElectionTimeout = rand.IntN(151) + 150 // random int between [150, 300]
}

func (n *Node) AppendEntries(args *HbArgs, reply *int) error {
	return nil
}

func (n *Node) RequestVote(args *ReqVoteArgs, reply *int) error {
	return nil
}


func parseArgs() CliArgs {
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

func main() {
	args := parseArgs()
	port := BASE_PORT + args.id

	node := new(Node)
	node.id = args.id
	rpc.Register(node) // register node with `rpc.DefaultServer`
	rpc.HandleHTTP() // register an HTTP handler for RPC communication

	node.SetNewElectionTimeout()
	fmt.Println("Node", node.id, "Election Timeout:", node.ElectionTimeout)

	// sample test endpoint
	http.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {
		io.WriteString(res, "RPC SERVER LIVE!")
	})
	// listen and serve requests to the default HTTP server
	addr := "localhost:" + strconv.Itoa(port)

	wg := new(sync.WaitGroup)
	wg.Go(func() {
		http.ListenAndServe(addr, nil)
	})

	// now connect to the other nodes
	peers := make(map[int]*rpc.Client, args.N-1)
	for i := range(args.N) {
		if i == args.id {
			continue // don't connect to self
		}
		peerAddr := "localhost:" + strconv.Itoa(BASE_PORT + i)
		
		var peer *rpc.Client
		var err error
		for i := range(MAX_RETRIES) { // continuous for loop for retrying
			peer, err = rpc.DialHTTP("tcp", peerAddr)
			if err == nil {
				break
			}
			fmt.Println("Waiting for node ", i, "'s server to setup")
			time.Sleep(200 * time.Millisecond)
		}
		peers[i] = peer
		fmt.Println("Node", node.id, "connected to node", i)
	}
	
	wg.Wait()
}