package main

import (
	"log"

	"github.com/knaka/querysan"
)

func main() {
	err := querysan.Initialize()
	if err != nil {
		log.Fatal(err)
	}
}
