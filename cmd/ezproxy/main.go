// cmd/ezproxy/main.go
package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: ezproxy <init|apply|remove|status>")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "init":
		fmt.Println("init: not yet implemented")
	case "apply":
		fmt.Println("apply: not yet implemented")
	case "remove":
		fmt.Println("remove: not yet implemented")
	case "status":
		fmt.Println("status: not yet implemented")
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", os.Args[1])
		os.Exit(1)
	}
}
