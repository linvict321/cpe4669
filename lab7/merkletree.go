package main

import (
	"crypto/md5"
	"fmt"
)

type MTNode struct {
	Key string
	Hash string
	Value string
	Parent *MTNode
	Children map[string]*MTNode
}

func (node *MTNode) Insert(key string, value string) string {
	//TODO: have a check that sees if you need to do Insert recursively
	// would use the returned hash from the child to adjust your hash, up until the root

	dataNode := new(MTNode)
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

<<<<<<< HEAD
func (node *Node) DeleteNode(key string) error {

	
	return nil
}

func InitMerkleTree() *Node {
	node := new(Node)
	node.Children = make(map[string]*Node)
=======
func InitMerkleTree() *MTNode {
	node := new(MTNode)
	node.Children = make(map[string]*MTNode)
>>>>>>> a44640436ff33951405422e2aa24c49f5214cde7

	hash := md5.New()
	node.Hash = string(hash.Sum(nil))
	return node
}
<<<<<<< HEAD
/*
func main() {
=======

func TestMerkleTree() {
>>>>>>> a44640436ff33951405422e2aa24c49f5214cde7
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
*/
