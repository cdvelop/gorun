package main

import (
	"fmt"
	"time"
)

func main() {
	fmt.Println("LONG_PROGRAM_STARTED")

	// Run for a longer time (useful for testing forced termination)
	for i := 0; i < 100; i++ {
		time.Sleep(50 * time.Millisecond)
		fmt.Printf("LONG_TICK_%d\n", i)
	}

	fmt.Println("LONG_PROGRAM_FINISHED")
}
