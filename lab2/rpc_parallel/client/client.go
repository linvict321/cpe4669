package main

import (
	"fmt"
	"log"
	"net/rpc"
	"sync"
	"time"

)

//was thinking i'd have client request server to find string search in 1MB file

type Args struct{
	Target string
	FileName string
	Workers int
	CurrWorker int
}

func callServer(clientId int, wg *sync.WaitGroup, numclients int){
	defer wg.Done() // so that it runs
	client, err := rpc.Dial("tcp", "127.0.0.1: 8080")

	if err != nil {
		log.Fatalf("Connection failed: %v", err)
	}
	defer client.Close()

	args := Args{Target: "Wikipedia", FileName: "RPC_text.txt", Workers: numclients, CurrWorker: clientId}
	var reply int

	err = client.Call("WordSearch.Search", args, &reply)
	if err != nil {
		log.Fatalf("RPC execution failed: %v", err)
	}
	fmt.Printf("Worker %d Found \" %s \" in %s this many times: %d\n", clientId,
			args.Target, args.FileName, reply)

}

//run several clients w/ different parameters
func main(){
	start := time.Now()

	numclients := 1
	var wg sync.WaitGroup
	wg.Add(numclients)
	for i := 1; i<= numclients; i++ {

		go callServer(i, &wg, numclients)
	}
	wg.Wait()

	//calls the adder function inside the server
	client, err := rpc.Dial("tcp", "127.0.0.1: 8080")
	if err != nil {
		log.Fatalf("Connection failed: %v", err)
	}
	defer client.Close()
	occurrences := new(int)
	OCall := client.Go("WordSearch.AddedSum", struct{}{}, occurrences, nil)
	replyCall := <- OCall.Done
	if replyCall.Error != nil {
		log.Fatalf("RPC execute failed: %v\n", replyCall.Error)
	}
	
	fmt.Printf("Total: %d \n", *occurrences)
	fmt.Printf("Time elapsed: %d", time.Since(start))
}