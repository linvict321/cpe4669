package main

import (
	"crypto/md5"
	"fmt"
)

type Node struct {
	Key string
	Hash string
	Value string
	Parent *Node
	Children map[string]*Node
}

func (node *Node) Insert(key string, value string) string {
	//TODO: have a check that sees if you need to do Insert recursively
	// would use the returned hash from the child to adjust your hash, up until the root

	dataNode := new(Node)
	dataNode.Key = key
	dataNode.Value = value
	dataNode.Parent = node
	dataNode.Children = nil

	hash := md5.New()
	hash.Write([]byte(value))
	dataNode.Hash = string(hash.Sum(nil))

	// now add new data block into parent node and modify parent hash
	node.Children[dataNode.Hash] = dataNode
	hash.Write([]byte(node.Hash))  // combine child's hash with our original parent hash to get the updated parent hash
	node.Hash = string(hash.Sum(nil))
	return node.Hash
}

func InitMerkleTree() *Node {
	node := new(Node)
	node.Children = make(map[string]*Node)

	hash := md5.New()
	node.Hash = string(hash.Sum(nil))
	return node
}

func main() {
	node := InitMerkleTree()

	fmt.Println(node)
	fmt.Printf("%x\n", []byte(node.Hash))
	node.Insert("key", "value")
	fmt.Printf("%x\n", []byte(node.Hash))
	
	node2 := InitMerkleTree()
	node2.Insert("key", "value")
	fmt.Println(node2)

	fmt.Println(node.Hash == node2.Hash)
}

