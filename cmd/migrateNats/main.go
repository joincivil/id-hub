package main

import (
	"fmt"
	"os"

	"github.com/joincivil/id-hub/pkg/idhubmain"
)

func main() {
	err := idhubmain.RunMigration()
	if err != nil {
		fmt.Printf("Error running the migration: err: %v\n", err)
		os.Exit(1)
	}
}
