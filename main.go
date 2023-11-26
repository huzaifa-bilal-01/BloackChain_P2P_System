package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"net"
	"sync"
	"time"
)

// Node represents a peer in the P2P network
type Node struct {
	ID           int
	Port         int
	IP           string
	Neighbors    []string
	CurrentBlock Block
	mu           sync.Mutex
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

// handling the transaction receiveing from the client
func (node *Node) handleClient(conn net.Conn) {
	defer conn.Close()
	scanner := bufio.NewScanner(conn)
	trx := []Transaction{}
	for scanner.Scan() {
		transactionData := scanner.Text()
		fmt.Printf("Node %d: Received transaction from neighbor: %s\n", node.ID, transactionData)
		trx = append(trx, Transaction{Data: transactionData})
		// node.mu.Lock()
		// node.Transactions = append(node.Transactions, Transaction{Data: transactionData})
		// node.mu.Unlock()
		// fmt.Printf("ID=%d, IP=%s, Port=%d, Neighbors=%v, Transactions=%v\n", node.ID, node.IP, node.Port, node.Neighbors, node.Transactions)
	}
	flood_arr := []Transaction{}
	for _, upcoming_trx := range trx {
		check := false
		for _, exisiting_trx := range node.CurrentBlock.Transactions {
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
	if len(node.CurrentBlock.Transactions) >= 4 {
		node.CurrentBlock.mineBlock()
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

		//time.Sleep(5 * time.Second) // Simulating periodic contact
	}
}

// contactNeighbor simulates the client contacting a neighbor
func (node *Node) contactNeighbor(neighbor string) {
}

// Function to broadcast transactions
func (node *Node) floodingTrx(transactions []Transaction) {
	node.mu.Lock()
	defer node.mu.Unlock()
	node.CurrentBlock.Transactions = append(node.CurrentBlock.Transactions, transactions...)

	//Check for the duplicated Transactions

	for _, neighbor := range node.Neighbors {
		go node.brodcastingToNeigborNodes(neighbor, node.CurrentBlock.Transactions)
	}
	node.CurrentBlock.merkleroot = merkleRoot(node.CurrentBlock.Transactions).hash
}

func (node *Node) brodcastingToNeigborNodes(neighbor string, transactions []Transaction) {

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

	for _, trx := range transactions {
		_, err := fmt.Fprintf(conn, "%s\n", trx.Data)
		if err != nil {
			fmt.Printf("Node %d: There was error sending transaction to neigbor: %v\n", node.ID, err)
			return
		}
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
	genesisBlock := Block{
		prevBlockHash: "",
		timestamp:     time.Now().Unix(),
		nonce:         0,
		merkleroot:    "",
		Transactions:  []Transaction{},
	}

	genesisBlock.currentBlockHash = genesisBlock.blockHashCalculation()
	blockchain.addBlock(&genesisBlock)

	// Example usage with 8 nodes
	transactions := []Transaction{
		{Data: "1500$ Sent"},
		{Data: "1600$ Sent"},
		{Data: "1700$ Sent"},
		{Data: "1800$ Sent"},
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
	// Display P2P network
	time.Sleep(12 * time.Second)
	DisplayP2PNetwork(nodes)

	// Keep the main goroutine running
	select {}
}
