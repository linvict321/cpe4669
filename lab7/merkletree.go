package lab7

import(
	"fmt"
	"crypto/md5"
)

type struct MTNode {
	Key string,
	Hash []byte,
	Value []byte,
	Children []MTNode
}

func MTInsertNode() {
	
}

func BuildMerkleTree() MTNode {
	node := make(MTNode)
	return node
}

