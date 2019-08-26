package main

import (
	"fmt"
	"os"

	"github.com/joincivil/id-hub/pkg/idhubmain"
)

func main() {
	err := idhubmain.RunCLI()
	if err != nil {
		fmt.Printf("Error running the cli: err: %v\n", err)
		os.Exit(1)
	}
}
