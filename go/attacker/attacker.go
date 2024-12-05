package main

import (
	"fmt"
	"time"
)

type Attacker struct {
	Bandwidth int
}

func (a *Attacker) LaunchDDOS(target string) {
	fmt.Printf("Launching DDOS on %s with %d bandwidth\n", target, a.Bandwidth)
	time.Sleep(1 * time.Second) // Simulate attack
}

func main() {
	attacker := Attacker{Bandwidth: 10}
	attacker.LaunchDDOS("Drone 1")
}
