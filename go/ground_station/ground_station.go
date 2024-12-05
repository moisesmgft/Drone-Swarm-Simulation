package main

import (
	"encoding/json"
	"fmt"
	"math"
	"net"
	"os"
	"strconv"
	"time"
)

const (
	VISITED  int     = 1000
	MAX_DIST float64 = 10.0
)

func distance(pos1, pos2 [2]int) float64 {
	x := float64(pos1[0] - pos2[0])
	y := float64(pos1[1] - pos2[1])
	return math.Sqrt(math.Pow(x, 2) + math.Pow(y, 2))
}

// GroundStation struct to manage drones and received messages
type GroundStation struct {
	TotalDrones int
	GridSize    int
	Address     string
	DroneConn   []*net.UDPConn
	ServerConn  *net.UDPConn
	PythonConn  *net.UDPConn
	Position    [2]int
	DronesPos   [][2]int
	Graph       [][]bool
	Grid        [][]int
	Down        []bool
}

// Message struct to handle incoming and outgoing communication
type Message struct {
	Sender  int
	Content string
	Point   [2]int
	Path    [][2]int
}

// CheckError utility function to handle errors
func CheckError(err error) {
	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(1)
	}
}

// Initialize sets up the UDP connection for the GroundStation
func (gs *GroundStation) Initialize() {
	// My server where I listen
	serverAddr, err := net.ResolveUDPAddr("udp", gs.Address)
	CheckError(err)
	gs.ServerConn, err = net.ListenUDP("udp", serverAddr)
	CheckError(err)

	// Python
	pythonAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:9999")
	CheckError(err)
	gs.PythonConn, err = net.DialUDP("udp", nil, pythonAddr)
	CheckError(err)

	fmt.Printf("Ground Station started on %s\n", gs.Address)
	// Initialize connections to all drones
	gs.DroneConn = make([]*net.UDPConn, gs.TotalDrones+1)
	for i := 1; i <= gs.TotalDrones; i++ {
		droneAddr := fmt.Sprintf("127.0.0.1:%d", 10000+i)
		udpAddr, err := net.ResolveUDPAddr("udp", droneAddr)
		CheckError(err)

		conn, err := net.DialUDP("udp", nil, udpAddr)
		CheckError(err)
		gs.DroneConn[i] = conn

		fmt.Printf("Connected to Drone %d at %s\n", i, droneAddr)
	}
}

// Listen receives messages from drones
func (gs *GroundStation) Listen() {
	for {
		message, err := gs.ReceiveMessage()
		if err != nil {
			continue
		}
		if message.Content == "REACHED" {
			gs.Grid[message.Point[0]][message.Point[1]] = VISITED
			gs.CheckForCompletition()
			nxt := gs.GetClosest(message.Sender)
			if nxt[0] != -1 {
				msg := Message{
					Content: "GO",
					Point:   nxt,
				}
				gs.SendMessage(msg, message.Sender)
			}
		} else if message.Content == "DOWN" {
			gs.Down[message.Sender] = true
		}
	}
}

func (gs *GroundStation) ProcessReached(i int, j int, id int) {
	gs.Grid[i][j] = VISITED
	gs.CheckForCompletition()
	nxt := gs.GetClosest(id)
	if nxt[0] != -1 {
		msg := Message{
			Content: "GO",
			Point:   nxt,
		}
		gs.SendMessage(msg, id)
	}
}

func (gs *GroundStation) inside(i int, j int) bool {
	return i >= 0 && j >= 0 && i < gs.GridSize && j < gs.GridSize
}

func (gs *GroundStation) GetClosest(id int) [2]int {
	vis := make([][]bool, gs.GridSize)
	for i := range vis {
		vis[i] = make([]bool, gs.GridSize)
	}
	queue := make([][2]int, 0)

	queue = append(queue, gs.DronesPos[id])
	vis[gs.DronesPos[id][0]][gs.DronesPos[id][1]] = true

	directions := [][2]int{{-1, 0}, {1, 0}, {0, -1}, {0, 1}}

	for len(queue) > 0 {
		i, j := queue[0][0], queue[0][1]
		queue = queue[1:]
		if gs.Grid[i][j] != VISITED {
			return [2]int{i, j}
		}
		for _, d := range directions {
			ii := i + d[0]
			jj := j + d[1]
			if !gs.inside(ii, jj) || vis[ii][jj] {
				continue
			}
			queue = append(queue, [2]int{ii, jj})
			vis[ii][jj] = true
		}
	}

	return [2]int{-1, -1}
}

// ReceiveMessage listens for a single message from the UDP connection
func (gs *GroundStation) ReceiveMessage() (Message, error) {
	buf := make([]byte, 1024)
	n, _, err := gs.ServerConn.ReadFromUDP(buf)
	if err != nil {
		fmt.Printf("Failed to receive message: %v\n", err)
		return Message{}, err
	}

	var message Message
	err = json.Unmarshal(buf[:n], &message)
	if err != nil {
		fmt.Printf("Failed to parse message: %v\n", err)
		return Message{}, err
	}

	fmt.Printf("Received message from Drone %d: %s\n", message.Sender, message.Content)
	return message, nil
}

func (gs *GroundStation) SendMessage(message Message, id int) {
	// Check if the connection for the given ID is valid
	if gs.DroneConn[id] == nil {
		fmt.Printf("Ground Station: No connection found for drone %d, skipping message send\n", id)
		return
	}

	jsonMessage, err := json.Marshal(message)
	CheckError(err)
	_, err = gs.DroneConn[id].Write(jsonMessage)
	CheckError(err)
}

func (gs *GroundStation) Act() {
	for {
		updated := gs.UpdateConnectivity()
		if updated {
			gs.PostGraph()
		}
		time.Sleep(25 * time.Millisecond)
	}
}

func (gs *GroundStation) UpdateConnectivity() bool {
	ret := false
	for i := 1; i <= gs.TotalDrones; i++ {
		connected := false
		if !gs.Down[i] && distance(gs.Position, gs.DronesPos[i]) < MAX_DIST {
			connected = true
		}
		if gs.Graph[0][i] != connected {
			ret = true
		}
		gs.Graph[0][i] = connected
		gs.Graph[i][0] = connected
	}
	for i := 1; i <= gs.TotalDrones; i++ {
		for j := i + 1; j <= gs.TotalDrones; j++ {
			connected := false
			if !gs.Down[i] && !gs.Down[j] && distance(gs.DronesPos[i], gs.DronesPos[j]) < MAX_DIST {
				connected = true
			}
			if gs.Graph[i][j] != connected {
				ret = true
			}
			gs.Graph[i][j] = connected
			gs.Graph[j][i] = connected
		}
	}

	return ret
}

func (gs *GroundStation) PostGraph() {

	// Build the message
	message := ""

	// First line: total number of drones
	totalDrones := len(gs.Graph) - 1 // Exclude the ground station
	message += fmt.Sprintf("%d\n", totalDrones)

	// Print the adjacency matrix without row and column 0
	for i := 1; i <= totalDrones; i++ {
		for j := 1; j <= totalDrones; j++ {
			if gs.Down[i] || gs.Down[j] {
				message += "0 "
			} else {
				message += fmt.Sprintf("%t ", gs.Graph[i][j])
			}
		}
		message = message[:len(message)-1] + "\n" // Remove trailing space and add newline
	}

	// Add a line for the drones connected to the ground station (index 0)

	for i := 1; i <= totalDrones; i++ {
		if gs.Graph[0][i] && !gs.Down[i] {
			message += "1 "
		} else {
			message += "0 "
		}
	}
	message = message[:len(message)-1] + "\n"

	// Send the message
	_, err := gs.PythonConn.Write([]byte(message))
	if err != nil {
		fmt.Printf("GroundStation: Failed to send graph: %v\n", err)
	}

}

func (gs *GroundStation) CheckForCompletition() {
	for i := 0; i < gs.GridSize; i++ {
		for j := 0; j < gs.GridSize; j++ {
			if gs.Grid[i][j] != VISITED {
				return
			}
		}
	}
	msg := "SUCCESS"
	gs.PythonConn.Write([]byte(msg))
}

func (gs *GroundStation) UpdatePositions() {
	for {
		buf := make([]byte, 1024)
		n, _, err := gs.ServerConn.ReadFromUDP(buf)
		if err != nil {
			fmt.Printf("Failed to receive position: %v\n", err)
			continue
		}

		message := string(buf[:n])
		fmt.Printf("msg = %s\n", message)
		if len(message) > 0 {
			var id, x, y int
			_, err := fmt.Sscanf(message, "(POS,%d,%d,%d)", &id, &x, &y)
			if err == nil {
				gs.DronesPos[id] = [2]int{x, y}
				fmt.Printf("Updated Drone %d position to (%d, %d)\n", id, x, y)
			} else {
				_, err = fmt.Sscanf(message, "(REACHED,%d,%d,%d)", &id, &x, &y)
				if err == nil {
					gs.ProcessReached(x, y, id)
					fmt.Printf("Drone %d reached (%d,%d)\n", id, x, y)
				}
			}

		}
	}
}

func initialize2DArray(rows, cols int) [][]int {
	array := make([][]int, rows)
	for i := 0; i < rows; i++ {
		array[i] = make([]int, cols)
	}
	return array
}
func initializeBoolArray(rows, cols int) [][]bool {
	array := make([][]bool, rows)
	for i := 0; i < rows; i++ {
		array[i] = make([]bool, cols)
	}
	return array
}

// Main function to run the GroundStation
func main() {
	// Checking if the required arguments are passed
	if len(os.Args) != 3 {
		fmt.Println("Usage: ground_station.go <totalDrones> <gridSize>")
		return
	}

	totalDrones, err := strconv.Atoi(os.Args[1])
	CheckError(err)
	gridSize, err := strconv.Atoi(os.Args[2])
	CheckError(err)

	address := "127.0.0.1:10000"

	gs := GroundStation{
		TotalDrones: totalDrones,
		GridSize:    gridSize,
		Address:     address,
		Graph:       initializeBoolArray(totalDrones+1, totalDrones+1),
		DronesPos:   make([][2]int, totalDrones+1),
		Down:        make([]bool, totalDrones+1),
		Grid:        initialize2DArray(gridSize, gridSize),
	}

	// Initialize and connect to drones
	gs.Initialize()
	time.Sleep(1 * time.Second)

	// Start listening for incoming messages
	// go gs.Listen()

	// Act
	go gs.Act()

	gs.UpdatePositions()

	msg := Message{
		Content: "GO",
		Point:   [2]int{7, 7},
	}
	gs.SendMessage(msg, 1)
}
