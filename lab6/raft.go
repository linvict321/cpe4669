package main

import (
	// "io"
	"fmt"
	// "math"
	"math/rand/v2"
	// "net"
	// "net/rpc"
	// "sync"
	// "time"
)

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

	ID              int
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

func (n *Node) SetNewElectionTimeout() {
	n.ElectionTimeout = rand.IntN(151) + 150 // random int between [150, 300]
}

func (n *Node) AppendEntries(args *HbArgs, reply *int) error {
	return nil
}

func (n *Node) RequestVote(args *ReqVoteArgs, reply *int) error {
	return nil
}


func main() {
	// just a quick test rn
	node := new(Node)
	node.SetNewElectionTimeout()
	fmt.Println("Election Timeout: ", node.ElectionTimeout)
}