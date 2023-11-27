package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"regexp"
	"strconv"
	"sync"
	"time"
)

// Node represents a peer in the P2P network
type Node struct {
	ID            int
	Port          int
	IP            string
	Neighbors     []string
	CurrentBlock  Block
	ReceivedBlock Block
	mu            sync.Mutex
}

// BootstrapNode represents the bootstrap node
var BootstrapNode = Node{
	ID:   0,
	Port: 5000,
	IP:   "127.0.0.1",
}

var nodeIDCounter = 1
var portCounter = 6000

// RegisterNode registers a new node with the bootstrap node
func RegisterNode(node *Node) {
	BootstrapNode.mu.Lock()
	defer BootstrapNode.mu.Unlock()

	// Assign a unique ID and port to the new node
	node.ID = nodeIDCounter
	nodeIDCounter++

	node.Port = portCounter
	portCounter++

	// Update the BootstrapNode's Neighbors list
	BootstrapNode.Neighbors = append(BootstrapNode.Neighbors, fmt.Sprintf("%s:%d", node.IP, node.Port))
}

func assigningNeighbor(node *Node) {
	// Connect with a random subset of existing nodes
	existingNodes := append([]string{}, BootstrapNode.Neighbors...)
	rand.Shuffle(len(existingNodes), func(i, j int) {
		existingNodes[i], existingNodes[j] = existingNodes[j], existingNodes[i]
	})

	maxNeighbors := 5 // You can adjust the number of neighbors as needed
	for _, potentialNeighbor := range existingNodes {
		if potentialNeighbor != node.IP {
			node.Neighbors = append(node.Neighbors, potentialNeighbor)
			maxNeighbors--
		}

		if maxNeighbors == 0 {
			break
		}
	}
}

// StartNode starts a new node as both a server and a client
func StartNode(node *Node) {
	node.CurrentBlock = Block{
		prevBlockHash: blockchain.head.data.currentBlockHash,
		timestamp:     time.Now().Unix(),
		nonce:         0,
	}
	go node.startServer()
	go node.startClient()
	go node.mineCheck()
}

// startServer starts the server for the node
func (node *Node) startServer() {
	listen, err := net.Listen("tcp", fmt.Sprintf(":%d", node.Port))
	if err != nil {
		fmt.Printf("Node %d: There was error starting server: %v\n", node.ID, err)
		return
	}
	defer listen.Close()

	for {
		conn, err := listen.Accept()
		if err != nil {
			fmt.Printf("Node %d: There was error accepting connection: %v\n", node.ID, err)
			continue
		}

		go node.handleClient(conn)
	}
}

func (node *Node) handleTranscation(trx []Transaction) {

	flood_arr := []Transaction{}
	for _, upcoming_trx := range trx {
		check := false
		for _, exisiting_trx := range node.CurrentBlock.BlockTransactions {
			if exisiting_trx.Data == upcoming_trx.Data {
				check = true
				return
			}
		}
		if !check {
			flood_arr = append(flood_arr, upcoming_trx)
		}
	}
	node.floodingTrx(flood_arr)

}

func (node *Node) mineCheck() {
	if len(node.CurrentBlock.BlockTransactions) >= 4 {
		node.CurrentBlock.mineBlock()
		if node.ReceivedBlock.prevBlockHash == "" {
			fmt.Println("Node ", node.ID, " is the first node to brodcast")
			node.floodingBlock(node.CurrentBlock)
		}
	}
	node.CurrentBlock = Block{}
}

func (node *Node) handleBlock(block Block) {
	if block.prevBlockHash != node.ReceivedBlock.prevBlockHash {
		node.ReceivedBlock = block
		if node.ReceivedBlock.currentBlockHash == node.ReceivedBlock.blockHashCalculation() {
			fmt.Println("Block is valid")
			node.floodingBlock(node.ReceivedBlock)
		} else {
			fmt.Println("Block is invalid : ", node.ReceivedBlock.currentBlockHash, " | ", node.ReceivedBlock.blockHashCalculation())
		}
	}
	node.ReceivedBlock = Block{}
}

// handling the transaction receiveing from the client
func (node *Node) handleClient(conn net.Conn) {
	defer conn.Close()
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		data := scanner.Text()
		//fmt.Printf("Node %d: Received data from neighbor: %s\n", node.ID, data)
		var receivedTransactions []Transaction
		if err := json.Unmarshal([]byte(data), &receivedTransactions); err == nil {
			node.handleTranscation(receivedTransactions)
			continue
		} else {
			re := regexp.MustCompile(`\{([^ ]+) ([^ ]+) %!s\(int64=([0-9]+)\) %!s\(int=([0-9]+)\) ([^ ]+) \[([^\]]+)\]\}`)

			// Find matches in the input string
			matches := re.FindStringSubmatch(data)
			if matches == nil || len(matches) != 7 {
				//sreturn nil, fmt.Errorf("invalid input format")
			}

			// Extract values from matches
			currentBlockHash := matches[1]
			prevBlockHash := matches[2]
			timestamp, _ := strconv.ParseInt(matches[3], 10, 64)
			nonce, _ := strconv.Atoi(matches[4])
			merkleRoot := matches[5]
			transactionsStr := matches[6]

			re1 := regexp.MustCompile(`\{([^}]+)\}`)

			// Find matches in the transactions string
			matches1 := re1.FindAllStringSubmatch(transactionsStr, -1)
			if matches1 == nil {
			}

			// Extract values from matches
			var extractedTransactions []Transaction
			for _, match := range matches1 {
				extractedTransactions = append(extractedTransactions, Transaction{Data: match[1]})
			}
			// Create and return the Block structure
			block := &Block{
				currentBlockHash:  currentBlockHash,
				prevBlockHash:     prevBlockHash,
				timestamp:         timestamp,
				nonce:             nonce,
				merkleroot:        merkleRoot,
				BlockTransactions: extractedTransactions,
			}

			//fmt.Println("received block: ", block)
			node.handleBlock(*block)

		}

	}
}

// startClient simulates the client functionality by periodically contacting neighbors
func (node *Node) startClient() {
	for {
		node.mu.Lock()
		neighbors := append([]string{}, node.Neighbors...)
		node.mu.Unlock()

		for _, neighbor := range neighbors {
			go node.contactNeighbor(neighbor)
		}
	}
}

// contactNeighbor simulates the client contacting a neighbor
func (node *Node) contactNeighbor(neighbor string) {
}

// Function to broadcast transactions
func (node *Node) floodingTrx(transactions []Transaction) {
	node.mu.Lock()
	defer node.mu.Unlock()
	node.CurrentBlock.BlockTransactions = append(node.CurrentBlock.BlockTransactions, transactions...)

	for _, neighbor := range node.Neighbors {
		go node.brodcastingTrxToNeigborNodes(neighbor, node.CurrentBlock.BlockTransactions)
	}
	node.CurrentBlock.merkleroot = merkleRoot(node.CurrentBlock.BlockTransactions).hash
}

func (node *Node) floodingBlock(block Block) {
	node.mu.Lock()
	defer node.mu.Unlock()
	node.CurrentBlock = block
	for _, neighbor := range node.Neighbors {
		go node.brodcastingBlockToNeigborNodes(neighbor, node.CurrentBlock)
	}
}

func (node *Node) brodcastingBlockToNeigborNodes(neighbor string, block Block) {
	fmt.Printf("Nocde %d: Broadcasting block to neighbor %s\n", node.ID, neighbor)
	node.mu.Lock()
	defer node.mu.Unlock()
	time.Sleep(1 * time.Second)
	conn, err := net.Dial("tcp", neighbor)
	if err != nil {
		fmt.Printf("Node %d: There was error connecting to neighbor: %v\n", node.ID, err)
		return
	}
	defer conn.Close()

	fmt.Fprintf(conn, "%s\n", block)

}

func (node *Node) brodcastingTrxToNeigborNodes(neighbor string, transactions []Transaction) {

	fmt.Printf("Nocde %d: Broadcasting transactions to neighbor %s\n", node.ID, neighbor)
	node.mu.Lock()
	defer node.mu.Unlock()
	time.Sleep(1 * time.Second)
	conn, err := net.Dial("tcp", neighbor)
	if err != nil {
		fmt.Printf("Node %d: There was error connecting to neighbor: %v\n", node.ID, err)
		return
	}
	defer conn.Close()
	encoder := json.NewEncoder(conn)
	if err := encoder.Encode(transactions); err != nil {
		fmt.Printf("Node %d: There was an error encoding and sending transactions to neighbor: %v\n", node.ID, err)
		return
	}
}

// DisplayP2PNetwork prints the details of the P2P network
func DisplayP2PNetwork(nodes []Node) {
	fmt.Println("P2P Network:")
	for i, node := range nodes {
		fmt.Printf("Node %d: ID=%d, IP=%s, Port=%d, Neighbors=%v, Block=%v\n", i+1, node.ID, node.IP, node.Port, node.Neighbors, node.CurrentBlock)
	}
	fmt.Println("Bootstrap Node:", BootstrapNode)
}

var blockchain = BlockChain{}

func main() {
	var choice int

	fmt.Println("Select an option:")
	fmt.Println("1. Part #01")
	fmt.Println("2. Part #02")

	fmt.Print("Enter your choice (1 or 2): ")
	fmt.Scanln(&choice)

	switch choice {
	case 1:
		part01()
	case 2:
		part02()
	default:
		fmt.Println("Invalid choice. Please enter 1 or 2.")
	}
}

func part01() {

	transactions1 := []Transaction{
		{Data: "Huzaifa"},
		{Data: "Hamza"},
		{Data: "Ahmed"},
		{Data: "Daniyal"},
	}

	transactions2 := []Transaction{
		{Data: "frustrated"},
		{Data: "happy"},
		{Data: "mad"},
		{Data: "sad"},
	}

	changedTransactions := []Transaction{
		{Data: "meowww"},
		{Data: "woffff"},
		{Data: "krrrrr"},
		{Data: "shhhhh"},
	}

	prevBlockHash := ""
	block1 := blockCreation(prevBlockHash, transactions1)
	block2 := blockCreation(block1.currentBlockHash, transactions2)
	fmt.Println(block1)
	fmt.Println(block2)

	blockchain := BlockChain{}
	blockchain.addBlock(block1)
	blockchain.addBlock(block2)

	fmt.Println("*****BLOCK CHAIN*****")
	blockchain.displayBlockChain()

	if blockchain.validityCheck() {
		fmt.Println("VALID BLOCKS :)")
	} else {
		fmt.Println("The transaction in block has been tampered")
	}

	changeBlock(block1, changedTransactions)
	fmt.Println(block1)

	if blockchain.validityCheck() {
		fmt.Println("VALID BLOCKS :)")
	} else {
		fmt.Println("The transaction in block has been tampered")
	}

	fmt.Println("********MERKLE TREE********")
	displayMerkleTree(merkleRoot(block1.BlockTransactions), "")
}

func part02() {

	genesisBlock := Block{
		prevBlockHash:     "",
		timestamp:         time.Now().Unix(),
		nonce:             0,
		merkleroot:        "",
		BlockTransactions: []Transaction{},
	}

	genesisBlock.currentBlockHash = genesisBlock.blockHashCalculation()
	blockchain.addBlock(&genesisBlock)

	transactions := []Transaction{
		{Data: "1500USD Sent"},
		{Data: "1600USD Sent"},
		{Data: "1700USD Sent"},
		{Data: "1800USD Sent"},
	}
	nodes := make([]Node, 8)

	for i := range nodes {
		nodes[i].IP = "127.0.0.1"
		RegisterNode(&nodes[i])
	}
	for i := range nodes {
		assigningNeighbor(&nodes[i])
		StartNode(&nodes[i])
	}
	nodes[4].floodingTrx(transactions)

	time.Sleep(12 * time.Second)
	DisplayP2PNetwork(nodes)

	select {}
}
