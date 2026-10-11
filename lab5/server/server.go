package main

import (
	"io"
	"lab5/shared"
	"net/http"
	"net/rpc"
)

func main() {
	// create a Membership list
	nodes := shared.NewMembership()
	requests := shared.NewRequests()
	//should we change requests to be a queue??? FIFO

	// register nodes with `rpc.DefaultServer`
	rpc.Register(nodes)
	rpc.Register(requests)

	// register an HTTP handler for RPC communication
	rpc.HandleHTTP()

	// sample test endpoint
	http.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {
		io.WriteString(res, "RPC SERVER LIVE!")
	})

	// listen and serve default HTTP server
	http.ListenAndServe("localhost:9005", nil)
	//we need to implement a queue in case two neighbors send to the same node at same time, want to only send once at a time

	/* From Lab2:
	for{
		conn, _ := listen.Accept()
		go rpc.ServeConn(conn)
	}
	*/
}
