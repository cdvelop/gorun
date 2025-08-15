package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("ERROR_PROGRAM_STARTED")
	fmt.Println("This program will exit with error code 1")
	fmt.Println("ERROR_PROGRAM_FINISHED")
	os.Exit(1)
}
