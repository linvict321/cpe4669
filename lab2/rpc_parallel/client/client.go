package main

import(
	"fmt"
	"log"
	"net/rpc"
	"time"
)

//was thinking i'd have client request server to find string search in 1MB file

//placeholder
type Args struct{
	Target string
	FileName string
}

func main(){
	client, err := rpc.Dial("tcp", "127.0.0.1: 8080")
	start := time.Now()

	if err != nil {
		log.Fatalf("Connection failed: %v", err)
	}
	defer client.Close()

	args := Args{Target: "Wikipedia", FileName: "RPC_text.txt"}
	var reply int

	err = client.Call("WordSearch.Search", args, &reply)
	if err != nil {
		log.Fatalf("RPC execution failed: %v", err)
	}
	t := time.Now()
	fmt.Printf("Found %s in %s this many times: %d in %d", 
				args.Target, args.FileName, reply, t.Sub(start))
}