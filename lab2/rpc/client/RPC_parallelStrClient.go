package main

import(
	"fmt"
	"log"
	"net/rpc"
)

//was thinking i'd have client request server to find string search in 1MB file

//placeholder
type Args struct{
	A, B int
}

func main(){
	client, err := rpc.Dial("tcp", "127.0.0.1: 8080")

	if err != nil {
		log.Fatalf("Connection failed: %v", err)
	}

	defer client.Close()

	args := Args{A:7, B:6}
	var reply int

	err = client.Call("filler words", args, &reply)
	if err != nil {
		log.Fatalf("RPC execution failed: %v", err)
	}


	fmt.Printf("RPC Result: %d", args.A)
}