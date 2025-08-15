package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	// Setup signal handling
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	// Print start message
	fmt.Println("PROGRAM_STARTED")

	// Create a ticker to print periodic messages
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	counter := 0
	for {
		select {
		case <-ticker.C:
			counter++
			fmt.Printf("TICK_%d\n", counter)
			if counter >= 10 {
				fmt.Println("PROGRAM_FINISHED")
				return
			}
		case sig := <-c:
			fmt.Printf("SIGNAL_RECEIVED_%s\n", sig.String())
			fmt.Println("PROGRAM_GRACEFUL_EXIT")
			return
		}
	}
}
