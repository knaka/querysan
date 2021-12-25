package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/knaka/querysan/db"
)

func main() {
	if len(os.Args) == 1 {
		_, err := fmt.Fprintf(os.Stderr, "No sub command specified\n")
		if err != nil {

		}
		os.Exit(1)
	}
	subCommand := os.Args[1]
	args := os.Args[2:]
	switch subCommand {
	case "query":
		fmt.Println("cp0")
		// words := tokenizer.Words(args[0])
		words := args
		fmt.Println(words)
		_ = db.SimpleConnect("/Users/knaka/.local/share/querysan/main.db")
		fmt.Println("cp1")
		paths, _ := db.Query(strings.Join(words, " "))
		fmt.Println("cp2")
		for _, path := range paths {
			fmt.Println(path)
		}
	default:
		_, err := fmt.Fprintf(os.Stderr, "No such command: %s\n", subCommand)
		if err != nil {

		}
	}
}
