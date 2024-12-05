package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"strconv"
	"sync"
	"time"
)

func CheckError(err error) {
	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(1)
	}
}

// Drone struct holds the drone's ID, position, neighbors, and mutex for thread safety
type Drone struct {
	ID int

	Position     [2]int
	GoalPosition [2]int
	Speed        int
	Stop         bool

	Mutex sync.Mutex

	ServerConn *net.UDPConn
	SimConn    *net.UDPConn
}

type Message struct {
	Sender  int
	Content string
	Point   [2]int
}

// ConnectToNetwork initializes UDP connections with neighbors, retrying until successful
func (d *Drone) ConnectToNetwork() {

	myAddress := fmt.Sprintf("127.0.0.1:%d", 10000+d.ID)
	serverAddr, err := net.ResolveUDPAddr("udp", myAddress)
	CheckError(err)
	d.ServerConn, err = net.ListenUDP("udp", serverAddr)
	CheckError(err)

	// Connect to simulation
	server := "127.0.0.1:10000"
	serverAddr, err = net.ResolveUDPAddr("udp", server)
	CheckError(err)
	conn, err := net.DialUDP("udp", nil, serverAddr)
	CheckError(err)
	d.SimConn = conn
}

// Listen initializes the server for receiving messages from other drones
func (d *Drone) Listen() {
	for {
		message, err := d.ReceiveMessage()
		CheckError(err)
		fmt.Printf("Received message: {%s, {%d,%d}}\n",
			message.Content, message.Point[0], message.Point[1])
		if message.Content == "GO" {
			d.GoalPosition = d.Position
			d.Stop = false
		}
	}
}

// Receive message
func (d *Drone) ReceiveMessage() (Message, error) {
	buf := make([]byte, 1024)
	// Ler (uma vez somente) da conexão UDP a mensagem
	n, _, err := d.ServerConn.ReadFromUDP(buf)
	CheckError(err)
	var message Message
	err = json.Unmarshal(buf[:n], &message)
	CheckError(err)
	return message, err
}

func (d *Drone) PublishPosition() {
	conn, err := net.Dial("udp", "127.0.0.1:9999") // Port 9999 for publishing
	if err != nil {
		fmt.Printf("Drone %d: Failed to connect to UDP socket: %v\n", d.ID, err)
		return
	}
	defer conn.Close()

	for {
		time.Sleep(1 * time.Second)
		d.Mutex.Lock()
		position := fmt.Sprintf("(POS,%d,%d,%d)", d.ID, d.Position[0], d.Position[1])
		d.Mutex.Unlock()

		_, err := conn.Write([]byte(position))
		if err != nil {
			fmt.Printf("Drone %d: Failed to send position to 9999: %v\n", d.ID, err)
		}
		_, err = d.SimConn.Write([]byte(position))
		if err != nil {
			fmt.Printf("Drone %d: Failed to send position to Sim: %v\n", d.ID, err)
		}
	}
}

// Act performs drone actions based on received messages
func (d *Drone) Act() {
	for {
		if d.Stop {
			continue
		}
		d.Move()
		d.CheckForArrival()
		time.Sleep(1 * time.Second)
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func (d *Drone) Move() {
	d.Mutex.Lock()
	defer d.Mutex.Unlock()

	if d.Position[0] != d.GoalPosition[0] {
		step := d.Speed
		if d.GoalPosition[0] < d.Position[0] {
			step = -d.Speed
		}
		if abs(d.GoalPosition[0]-d.Position[0]) < abs(step) {
			d.Position[0] = d.GoalPosition[0]
		} else {
			d.Position[0] += step
		}
	} else if d.Position[1] != d.GoalPosition[1] {
		step := d.Speed
		if d.GoalPosition[1] < d.Position[1] {
			step = -d.Speed
		}
		if abs(d.GoalPosition[1]-d.Position[1]) < abs(step) {
			d.Position[1] = d.GoalPosition[1]
		} else {
			d.Position[1] += step
		}
	}
}

func (d *Drone) CheckForArrival() {
	if d.GoalPosition == d.Position {
		d.Stop = true

		// Send message of reached position
		msg := fmt.Sprintf("(\"REACHED\",%d,%d,%d)", d.ID, d.Position[0], d.Position[1])
		d.SimConn.Write([]byte(msg))
	}
}

func main() {
	// Checking if the required arguments are passed
	if len(os.Args) != 4 {
		fmt.Println("Usage: drone.go <droneID> <xPosition> <yPosition>")
		return
	}

	// Parse arguments
	droneID, err := strconv.Atoi(os.Args[1])
	if err != nil {
		fmt.Println("Invalid drone ID")
		return
	}

	xPos, err := strconv.Atoi(os.Args[2])
	if err != nil {
		fmt.Println("Invalid x position")
		return
	}

	yPos, err := strconv.Atoi(os.Args[3])
	if err != nil {
		fmt.Println("Invalid y position")
		return
	}

	// Initialize the drone
	Point := [2]int{xPos, yPos}
	drone := Drone{
		ID:           droneID,
		Position:     Point,
		GoalPosition: Point,
		Stop:         true,
		Speed:        1,
	}

	fmt.Printf("Initializing Drone %d at position (%d, %d)\n", drone.ID, drone.Position[0], drone.Position[1])

	// Connect to neighbors
	time.Sleep(2 * time.Second)
	drone.ConnectToNetwork()
	time.Sleep(2 * time.Second)

	// Start receiving messages
	go drone.Listen()

	// Start acting on messages
	go drone.Act()

	// Publish position
	drone.PublishPosition()
}
