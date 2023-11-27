package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"
)

const _MINING_DIFFICULTY_ = 4

type Transaction struct {
	Data string
}

type Block struct {
	currentBlockHash  string
	prevBlockHash     string
	timestamp         int64
	nonce             int
	merkleroot        string
	BlockTransactions []Transaction
}

type MerkleNode struct {
	hash  string
	left  *MerkleNode
	right *MerkleNode
}

// Creating a linkedList to implement BlockChain
type BlockChainNode struct {
	data *Block
	next *BlockChainNode
}

type BlockChain struct {
	head *BlockChainNode
}

func (obj *BlockChain) addBlock(b *Block) {
	newNode := &BlockChainNode{data: b, next: nil}
	if obj.head == nil {
		obj.head = newNode
		return
	}
	currentNode := obj.head
	for currentNode.next != nil {
		currentNode = currentNode.next
	}
	currentNode.next = newNode
}

// Displaying the BlockChain
func (obj *BlockChain) displayBlockChain() {
	currentNode := obj.head
	for currentNode != nil {
		fmt.Printf("Block Hash: %s\n", currentNode.data.currentBlockHash)
		currentNode = currentNode.next
	}
}

// Calculating the hash of current Block
func (block1 *Block) blockHashCalculation() string {
	//Block header consist of prevBlockhash,nonce,timestamp,merkleroot and trasactions in that block
	blockHeader := fmt.Sprintf("%s%d%d%s%s", block1.prevBlockHash, block1.timestamp, block1.nonce, block1.merkleroot, block1.BlockTransactions)
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
func blockCreation(prevBlockHash string, trasactions []Transaction) *Block {
	block := &Block{
		prevBlockHash:     prevBlockHash,
		timestamp:         time.Now().Unix(),
		nonce:             0,
		BlockTransactions: trasactions,
	}

	block.merkleroot = merkleRoot(trasactions).hash
	mined := block.mineBlock()
	if mined {
		return block
	} else {
		fmt.Println("Block is not minned")
	}
	return nil
}

func (b *Block) mineBlock() bool {
	b.currentBlockHash = b.blockHashCalculation()
	for {
		//Calculating the trailing zero in block hash by iterating over last number of zero's
		for i := len(b.currentBlockHash) - 1; i >= len(b.currentBlockHash)-_MINING_DIFFICULTY_; i-- {
			if b.currentBlockHash[i] != '0' {
				b.nonce++
				b.currentBlockHash = b.blockHashCalculation()
			}
		}
		return true

	}
}

// Checking the validity of block along with chain
func (obj *BlockChain) validityCheck() bool {
	currentNode := obj.head
	for currentNode != nil && currentNode.next != nil {
		currentBlock := currentNode.data
		nextBlock := currentNode.next.data

		if currentBlock.currentBlockHash != nextBlock.prevBlockHash {
			return false
		}

		if currentBlock.currentBlockHash != currentBlock.blockHashCalculation() && nextBlock.currentBlockHash != nextBlock.blockHashCalculation() {
			return false
		}

		currentNode = currentNode.next
	}
	return true
}

// Changing the block
func changeBlock(b *Block, transactions []Transaction) {
	b.BlockTransactions = transactions
	b.merkleroot = merkleRoot(transactions).hash
	b.currentBlockHash = b.blockHashCalculation()
}

// Displaying Block
func (b *Block) String() string {
	return fmt.Sprintf("Block:\n"+
		"|  Previous Block Hash: %s\n"+
		"|  Current Block Hash:  %s\n"+
		"|  Timestamp:           %s\n"+
		"|  Nonce:               %d\n"+
		"|  Merkle Root:         %s\n"+
		"|  Transactions:        %v\n",
		b.prevBlockHash,
		b.currentBlockHash,
		time.Unix(b.timestamp, 0).Format("2006-01-02 15:04:05"),
		b.nonce,
		b.merkleroot,
		b.BlockTransactions,
	)
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
