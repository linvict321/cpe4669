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
	// lastHb          int // in milliseconds
	electionTimeout int // in milliseconds, will be between [150, 300] ms
	electionTimer   *time.Timer
	totalNodes      int // includes dead peers
	majority        int // strict majority = N // 2 + 1
	port            int
	peers           map[int]*rpc.Client
	mu              sync.Mutex // used within RequestVote
	hbCh            chan HbArgs
	hbTimer			*time.Timer // only used if leader
	leaderId        int
}

type HbArgs struct {
	Term     int // leader's currentTerm
	LeaderId int
}

type ReqVoteArgs struct {
	Term        int // candidate's term
	CandidateId int // candidate requesting vote
}

type ReqVoteRespArgs struct {
	VoteGranted bool // true if node voted for candidate, otherwise false
	Term        int // candidate and node's term
	VoterId     int // id of voting (responding) node
}

const BASE_PORT int = 8080
const MAX_RETRIES int = 10
const NOT_YET_VOTED_THIS_TERM int = -1
const HEARTBEAT_PERIOD int = 50 // how often leader sends a heartbeat, in ms

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

func (node *Node) SetNewElectionTimeout() {
	// TODO: Return arg of IntN back to 151 after debugging!
	node.electionTimeout = rand.IntN(10000) + 150 // random int between [150, 300]
	if node.electionTimer != nil {
		node.electionTimer.Reset(time.Duration(node.electionTimeout) * time.Millisecond)
	}
}

func(node *Node) PrintMsg(msg string) {
	fmt.Println("[ NODE", node.id, "] -", msg)
}

func initNode(args CliArgs) *Node { // Go garbage collector means don't need to worry about dangling pointer like in C
	node := new(Node)
	node.id = args.id
	node.port = BASE_PORT + args.id
	node.state = Follower
	node.votedFor = NOT_YET_VOTED_THIS_TERM
	node.voteCount = 0
	node.SetNewElectionTimeout()
	node.totalNodes = args.N
	node.majority = int(args.N / 2) + 1
	node.peers = make(map[int]*rpc.Client, args.N-1)
	//node.hbCh = make(chan HbArgs)
	node.electionTimer = time.NewTimer(time.Duration(node.electionTimeout) * time.Millisecond)
	node.hbTimer = time.NewTimer(time.Duration(HEARTBEAT_PERIOD) * time.Millisecond)

	msg := fmt.Sprintf("ElectionTimeout: %d ms", node.electionTimeout)
	node.PrintMsg(msg)
	return node
}

// TODO: fix race condition of node state when node is Candidate
func (node *Node) AppendEntries(hbArgs *HbArgs, reply *bool) error {

	node.mu.Lock()
	defer node.mu.Unlock()

	// only notify to reset electionTimer and stuff for an up-to-date leader
	if hbArgs.Term >= node.currentTerm {
		if hbArgs.Term > node.currentTerm{
			node.votedFor = NOT_YET_VOTED_THIS_TERM
		}
		node.currentTerm = hbArgs.Term
		node.state = Follower //candidate goes back to follower if it loses
		node.voteCount = 0
		node.leaderId = hbArgs.LeaderId
		node.SetNewElectionTimeout()
		*reply = true
	} else {
		*reply = false
	}

	node.PrintMsg("In AppendEntries")
	return nil
}

func (node *Node) RequestVote(voteReq *ReqVoteArgs, reply *ReqVoteRespArgs) error {
	node.mu.Lock()
	defer node.mu.Unlock()

	//adopt the higher term
	if voteReq.Term > node.currentTerm {
		node.currentTerm = voteReq.Term
		node.votedFor = NOT_YET_VOTED_THIS_TERM
		node.state = Follower
	}

	if voteReq.Term == node.currentTerm && (node.votedFor == NOT_YET_VOTED_THIS_TERM || node.votedFor == voteReq.CandidateId) {
		reply.VoteGranted = true
		node.SetNewElectionTimeout() //so no competing election starts
		node.votedFor = voteReq.CandidateId
		msg := fmt.Sprintf("Term %d: voted for Node %d", node.currentTerm, voteReq.CandidateId)
		node.PrintMsg(msg)
	} else {
		reply.VoteGranted = false
	}
	reply.Term = node.currentTerm
	reply.VoterId = node.id
	fmt.Printf("Voted by %d \n", node.id)
	return nil
}

func (node *Node) RequestVoteFromAllPeers() {
	node.mu.Lock()
	if node.state == Leader { //don't want to accidentally set leader back to candidate
		node.mu.Unlock()
		return
	}
	node.state = Candidate
	node.currentTerm++; //increment new term w/ election
	node.votedFor = node.id
	node.voteCount = 1 // votes for itself
	node.SetNewElectionTimeout()

	reqVoteArgs := &ReqVoteArgs{Term: node.currentTerm, CandidateId: node.id}
	resps := make(chan ReqVoteRespArgs, len(node.peers))
	for id, peer := range node.peers {
		resp := new(ReqVoteRespArgs)
		go func () {
			msg := fmt.Sprintf("Requesting vote from Node %d", id)
			node.PrintMsg(msg)
			peer.Call("Node.RequestVote", reqVoteArgs, resp)
			resps <- *resp
			}()
		}
	node.mu.Unlock()

	go func(peerCount int) {
		for range peerCount {
			resp := <-resps

			node.mu.Lock()
			if resp.VoteGranted {
				node.voteCount++
			}
			if node.voteCount >= node.majority {
				node.currentTerm++
				node.state = Leader 
				node.leaderId = node.id
				fmt.Printf("Node %d is leader\n", node.id);
				node.hbTimer.Reset(0)
				node.mu.Unlock()
				return
			}
			node.mu.Unlock()
		}
	}(len(node.peers))
}

func (node *Node) SendHbToAllPeers() {
	hbArgs := HbArgs{Term: node.currentTerm, LeaderId: node.id}
	var reply bool
	for id, peer := range node.peers {
		go func () {
			peer.Call("Node.AppendEntries", &hbArgs, &reply)
			if (reply == false) {
				msg := fmt.Sprintf("AppendEntry returned false from node %d\n", id)
				node.PrintMsg(msg)
			}
		}()
	}
}

func main() {
	args := ParseArgs()
	node := initNode(args)

	rpc.Register(node) // register node with `rpc.DefaultServer`
	rpc.HandleHTTP() // register an HTTP handler for RPC communication

	// fmt.Println("Node", node.id, "Election Timeout:", node.electionTimeout)

	// sample test endpoint
	http.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {
		io.WriteString(res, "RPC SERVER LIVE!")
	})
	// listen and serve requests to the default HTTP server
	addr := "localhost:" + strconv.Itoa(node.port)

	wg := new(sync.WaitGroup)
	wg.Go(func() {
		http.ListenAndServe(addr, nil)
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
			fmt.Println("Waiting for node ", i, "'s server to setup")
			time.Sleep(200 * time.Millisecond)
		}
		node.peers[i] = peer
		msg := fmt.Sprintf("Connected to node %d\n", i)
		node.PrintMsg(msg)
	}
	time.Sleep(time.Duration(3) * time.Second)

	// 
	// main logic loop
	for { 
		node.PrintMsg("In main loop!")
		select {
		// case hbArgs := <- node.hbCh:
		// 	node.electionTimer.Reset(time.Duration(node.electionTimeout) * time.Millisecond)
		// 	fmt.Println("Node", node.id, "received heartbrea from Node", hbArgs.LeaderId)
		case <- node.electionTimer.C:
			node.PrintMsg("ElectionTimer expired!")
			go node.RequestVoteFromAllPeers()
		case <- node.hbTimer.C:
			node.hbTimer.Reset(time.Duration(HEARTBEAT_PERIOD) * time.Millisecond)

			node.mu.Lock()
			isLeader := node.state == Leader
			node.mu.Unlock()
			if isLeader {
				node.SendHbToAllPeers()
			}
		}
	}

	wg.Wait()
	node.electionTimer.Stop()

}