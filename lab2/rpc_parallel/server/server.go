/* Obtain a text file approximately 1 MB in size. You may use text from sources such as Wikipedia repositories, public-domain books, or other appropriate text datasets.
Implement a string-search algorithm in Go.
The search function must:
Accept the search string as an input parameter.
Search the approximately 1 MB text file.
Return the number of occurrences of the specified string.
Create both sequential and parallel/distributed versions so their performance can be compared.

Run the RPC implementation using:

1 worker/node
2 workers/nodes
4 workers/nodes
Compare the execution times of the sequential, parallel/distributed, and RPC implementations.
Discuss the effect of communication overhead and the number of workers/nodes on overall performance.*/

package main

import (
	"io/ioutil"
	"log"
	"net"
	"net/rpc"
	"strings"
)

//idea: to run processes looking for the strings(diff num of worker nodes)

var wordsFound = 0

type WordSearch struct{}
type Args struct {
	Target string
	FileName string
	Workers int
	CurrWorker int

}

func (t *WordSearch) Search(args *Args, reply *int) error { 
	
	//1 node: read file, search for words
	// if split into other nodes, then split the lines in file, return the number of found to reply
	searchedWord := args.Target
	//read file
	content, err := ioutil.ReadFile(args.FileName)
	if err != nil {
		log.Fatal(err)
	}

	lines := strings.Split(string(content), "\n")
	found := 0

	//formatted startline and endline
	StartLine := (len(lines)/args.Workers)*(args.CurrWorker - 1)
	EndLine := ((len(lines)/args.Workers)*args.CurrWorker)
	for line := StartLine; line < EndLine; line++{
		allWords := strings.Split(lines[line], " ")
			//case sensitive
		for _, w := range allWords{
			if w == searchedWord{
				found += 1
			}
		}
	}

	wordsFound += found
	*reply = found
	return nil
}

func(t *WordSearch) AddedSum(args *Args, finalreply *int) error{
	*finalreply = wordsFound
	wordsFound = 0
	return nil
}

func main(){

	wordSearch := new(WordSearch)
	rpc.Register(wordSearch)

	rpc.HandleHTTP()
	listen, error := net.Listen("tcp", ":8080")
	if error != nil{
		log.Fatal("listen error:", error)
	}
	defer listen.Close()

	for{
		conn, _ := listen.Accept()
		go rpc.ServeConn(conn)
	}

}