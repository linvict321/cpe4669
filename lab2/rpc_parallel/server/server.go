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
    "log"
    "net"
    "net/rpc"

)

//idea: to run processes looking for the strings(diff num of worker nodes)

//placeholder struct
type wordSearch struct{}


type Args struct {
	target string
	fileName string
}

func (t *wordSearch) wordSearch(args *Args, reply *int) error { 
	
	
	*reply = 0
	return nil
}

func main(){

	//placeholder
	words := new(wordSearch)

	rpc.Register(words)
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