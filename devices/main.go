package main

import (
	"fmt"
	"math/rand"
	"time"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	fmt.Println("Starting all device simulators...")

	for i := 1; i <= 5; i++ {
		go SimulateFTD(i)
		go SimulateWSA(i)
		go SimulateESA(i)
		go SimulateISE(i)
	}

	select {} // keep program running forever
}
