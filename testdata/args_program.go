package main

import (
	"fmt"
	"os"
	"strings"
	"time"
)

func main() {
	fmt.Println("ARGS_PROGRAM_STARTED")
	fmt.Println("ARGS:" + strings.Join(os.Args[1:], ","))

	// Keep it alive for a short time to allow testing
	time.Sleep(50 * time.Millisecond)

	fmt.Println("ARGS_PROGRAM_FINISHED")
}
