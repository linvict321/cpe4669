package server

import(
	"fmt"
	"net"
	"net/rpc"
	"errors"
)

//3 steps
type ConsensusModule struct{}
type Log struct{}
type StateMachine struct{}

type Args struct{
	X int
}

func main(){

	//insert new struct + register it

	listener, err := net.Listen("tcp", ":8080")
	if err != nil{
		fmt.Println("Error Listening: ", err)
		return
	}
	defer listener.Close()

	fmt.Println("Server is listening on port 8080")
	rpc.Accept(listener)

}