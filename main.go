package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

type Transaction struct {
	Data string
}

type Block struct {
	currentBlockHash string
	prevBlockHash    string
	timestamp        int64
	nonce            int
	merkleroot       string
	Transactions     []Transaction
}

type MerkleNode struct {
	hash  string
	left  *MerkleNode
	right *MerkleNode
}

// Calculating the hash of current Block
func (block1 *Block) currentBlockHashCalulation() string {
	//Block header consist of prevBlockhash,nonce,timestamp,merkleroot and trasactions in that block
	blockHeader := fmt.Sprintf("%s%d%d%s%s", block1.prevBlockHash, block1.timestamp, block1.nonce, block1.merkleroot, block1.Transactions)
	hash_value := sha256.Sum256([]byte(blockHeader))
	hash_string := hex.EncodeToString(hash_value[:])
	return hash_string
}

// Calculating the hash of the data in merkle root
func hashCalculation(data string) string {
	hash_value := sha256.Sum256([]byte(data))
	hash_string := hex.EncodeToString(hash_value[:])
	return hash_string
}

// Merkle Root implementation
func merkleRoot(data []Transaction) *MerkleNode {
	//Calculating the hash of all the data and then appending in nodes of merkle tree it will be leaf nodes
	var nodes []*MerkleNode
	for _, val := range data {
		hash_data := &MerkleNode{hash: hashCalculation(val.Data)}
		nodes = append(nodes, hash_data)
	}
	//if there are more than 1 node than we can create merkle tree
	for len(nodes) > 1 {
		var level []*MerkleNode
		//Merkle tree is also known as binary hash tree so we have iterated i to i+=2
		for i := 0; i < len(nodes); i += 2 {
			left_node := nodes[i]
			right_node := left_node
			//If there is right tree than right node will be this
			if i+1 < len(nodes) {
				right_node = nodes[i+1]
			}
			parent_node := &MerkleNode{hash: hashCalculation(left_node.hash + right_node.hash), left: left_node, right: right_node}
			level = append(level, parent_node)
		}
		nodes = level
	}
	return nodes[0]
}

// Creation of new Block
func newBlockCreation(prevBlockHash string, trasactions []Transaction) *Block {
	block := &Block{
		prevBlockHash: prevBlockHash,
		timestamp:     time.Now().Unix(),
		nonce:         0,
		Transactions:  trasactions,
	}

	block.currentBlockHash = block.currentBlockHashCalulation()
	block.merkleroot = merkleRoot(trasactions).hash
	return block

}

// Displaying merkle tree
func displayMerkleTree(root_node *MerkleNode, identation string) {
	if root_node != nil {
		fmt.Println(identation+"Hash_Value:", root_node.hash)
		if root_node.left != nil {
			displayMerkleTree(root_node.left, identation+"    ")
		}
		if root_node.right != nil {
			displayMerkleTree(root_node.right, identation+"    ")
		}
	}
}

func main() {
	transactions := []Transaction{
		{Data: "Huzaifa"},
		{Data: "Hamza"},
		{Data: "Ahmed"},
		{Data: "Daniyal"},
	}

	prevBlockHash := "50a504831bd50fee3581d287168a85a8dcdd6aa777ffd0fe35e37290268a0153"
	block := newBlockCreation(prevBlockHash, transactions)
	root_node := block.merkleroot
	fmt.Println("Root node hash is:", root_node)
	fmt.Println("********MERKLE TREE********")
	displayMerkleTree(merkleRoot(transactions), "")

}
