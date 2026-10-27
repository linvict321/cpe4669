package lab7

import(
	"fmt"
	"crypto/md5"
)

type struct MTNode {
	Key []byte,
	Value []byte,
	Children []MTNode
}

func BuildMerkleTree() MTNode {
	node := make(MTNode)
	return node
}

