package main

import(
	"fmt"
	"log"
	"net/rpc"
)

//was thinking i'd have client request server to find string search in 1MB file

//placeholder
type Args struct{
	Target string
	FileName string
}

func main(){
	client, err := rpc.Dial("tcp", "127.0.0.1: 8080")

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

	fmt.Printf("Found %s in %s this many times: %d", args.Target, args.FileName, reply)
}