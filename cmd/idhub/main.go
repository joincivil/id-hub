package main

import (
	"fmt"
	"os"

	"github.com/joincivil/id-hub/pkg/idhubmain"
)

func main() {
	// To be implemented
	err := idhubmain.RunServer()
	if err != nil {
		fmt.Printf("Error: %v", err)
		os.Exit(1)
	}
}
