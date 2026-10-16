package client

import(
	"fmt"
	"net/rpc"
)

type Args struct{
	X int
}

func main(){

	client, err := rpc.Dial("tcp", "localhost:8080")
	if err != nil{
		fmt.Println("Error connecting", err)
		return
	}
	defer client.Close()

	//args + call + result

}