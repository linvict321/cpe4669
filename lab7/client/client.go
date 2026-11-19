package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net/rpc"
	"os"
	"strings"
)

// RPC matches exported struct fields, so these match the server's wire format.
type clientPutArgs struct {
	Key   string
	Value string
}

type clientGetArgs struct {
	Key string
}

type clientGetReply struct {
	Value string
}

func main() {
	serverAddr := flag.String("addr", "localhost:18080", "node RPC address")
	flag.Parse()
	client, err := rpc.DialHTTP("tcp", *serverAddr)
	if err != nil {
		log.Fatalf("Failed to connect to node: %v", err)
	}
	defer client.Close()

	fmt.Printf("Connected to %s. Commands: put <key> <value>, get <key>, quit\n", *serverAddr)
	scanner := bufio.NewScanner(os.Stdin)
	for fmt.Print("> "); scanner.Scan(); fmt.Print("> ") {
		fields := strings.Fields(scanner.Text())
		if len(fields) == 0 {
			continue
		}
		switch fields[0] {
		case "quit", "exit":
			return
		case "put":
			if len(fields) < 3 {
				fmt.Println("Usage: put <key> <value>")
				continue
			}
			args := clientPutArgs{Key: fields[1], Value: strings.Join(fields[2:], " ")}
			if err := client.Call("KV.Put", &args, &struct{}{}); err != nil {
				fmt.Println("Put failed:", err)
			} else {
				fmt.Println("OK")
			}
		case "get":
			if len(fields) != 2 {
				fmt.Println("Usage: get <key>")
				continue
			}
			args := clientGetArgs{Key: fields[1]}
			var reply clientGetReply
			if err := client.Call("KV.Get", &args, &reply); err != nil {
				fmt.Println("Get failed:", err)
			} else {
				fmt.Println(reply.Value)
			}
		default:
			fmt.Println("Commands: put <key> <value>, get <key>, quit")
		}
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
}
