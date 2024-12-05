package main

import (
	"encoding/json"
	"fmt"
	"math"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	MAX_DIST = 10.0 // Maximum distance for connectivity
	DRONE_PORT_BASE = 10000
)

type Drone struct {
	ID       int
	Position [2]float64
}

// Message for shortest path updates
type Message struct {
	Sender  int
	Content string
	Queue   []int
}

// Simulation struct to manage drones and connectivity
type Simulation struct {
	Drones map[int]Drone
	Graph  map[int][]int
	Mutex  sync.Mutex
}

// Calculate Euclidean distance between two points
func distance(pos1, pos2 [2]float64) float64 {
	return math.Sqrt(math.Pow(pos1[0]-pos2[0], 2) + math.Pow(pos1[1]-pos2[1], 2))
}

// Listen to drone positions
func (sim *Simulation) ListenForPositions() {
	addr, err := net.ResolveUDPAddr("udp", ":9999")
	if err != nil {
		fmt.Println("Error resolving UDP address:", err)
		return
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		fmt.Println("Error starting UDP listener:", err)
		return
	}
	defer conn.Close()

	buffer := make([]byte, 1024)

	for {
		n, _, err := conn.ReadFromUDP(buffer)
		if err != nil {
			fmt.Println("Error reading from UDP:", err)
			continue
		}

		data := string(buffer[:n])
		parts := strings.Split(data, ",")
		if len(parts) != 3 {
			fmt.Println("Invalid position data:", data)
			continue
		}

		id, _ := strconv.Atoi(parts[0])
		x, _ := strconv.ParseFloat(parts[1], 64)
		y, _ := strconv.ParseFloat(parts[2], 64)

		sim.Mutex.Lock()
		sim.Drones[id] = Drone{ID: id, Position: [2]float64{x, y}}
		sim.Mutex.Unlock()
	}
}

// Update connectivity graph based on positions
func (sim *Simulation) UpdateConnectivity() {
	for {
		time.Sleep(1 * time.Second)

		sim.Mutex.Lock()
		newGraph := make(map[int][]int)

		for id1, drone1 := range sim.Drones {
			for id2, drone2 := range sim.Drones {
				if id1 != id2 && distance(drone1.Position, drone2.Position) <= MAX_DIST {
					newGraph[id1] = append(newGraph[id1], id2)
				}
			}
		}

		// Check for changes in connectivity
		if !graphsEqual(sim.Graph, newGraph) {
			sim.Graph = newGraph
			sim.UpdateShortestPaths()
		}
		sim.Mutex.Unlock()
	}
}

// Compare two graphs for equality
func graphsEqual(g1, g2 map[int][]int) bool {
	if len(g1) != len(g2) {
		return false
	}
	for key, val1 := range g1 {
		val2, exists := g2[key]
		if !exists || !slicesEqual(val1, val2) {
			return false
		}
	}
	return true
}

// Compare two slices for equality
func slicesEqual(s1, s2 []int) bool {
	if len(s1) != len(s2) {
		return false
	}
	m := make(map[int]int)
	for _, v := range s1 {
		m[v]++
	}
	for _, v := range s2 {
		if m[v] == 0 {
			return false
		}
		m[v]--
	}
	return true
}

// Update shortest paths to the ground station
func (sim *Simulation) UpdateShortestPaths() {
	paths := sim.BFS(0)

	for droneID, path := range paths {
		if droneID == 0 {
			continue
		}
		sim.SendPathToDrone(droneID, path)
	}
}

// Perform BFS to find shortest paths from source
func (sim *Simulation) BFS(source int) map[int][]int {
	queue := []int{source}
	visited := make(map[int]bool)
	paths := make(map[int][]int)

	visited[source] = true
	paths[source] = []int{}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		for _, neighbor := range sim.Graph[current] {
			if !visited[neighbor] {
				visited[neighbor] = true
				paths[neighbor] = append(paths[current], current)
				queue = append(queue, neighbor)
			}
		}
	}

	return paths
}

// Send shortest path to a drone
func (sim *Simulation) SendPathToDrone(droneID int, path []int) {
	message := Message{
		Sender:  0,
		Content: "PATH",
		Queue:   path,
	}

	addr := fmt.Sprintf("127.0.0.1:%d", DRONE_PORT_BASE+droneID)
	conn, err := net.Dial("udp", addr)
	if err != nil {
		fmt.Printf("Error connecting to drone %d: %v\n", droneID, err)
		return
	}
	defer conn.Close()

	data, err := json.Marshal(message)
	if err != nil {
		fmt.Printf("Error marshaling message for drone %d: %v\n", droneID, err)
		return
	}

	_, err = conn.Write(data)
	if err != nil {
		fmt.Printf("Error sending message to drone %d: %v\n", droneID, err)
	}
}

func main() {
	sim := &Simulation{
		Drones: make(map[int]Drone),
		Graph:  make(map[int][]int),
	}

	go sim.ListenForPositions()
	go sim.UpdateConnectivity()

	select {} // Block forever
}