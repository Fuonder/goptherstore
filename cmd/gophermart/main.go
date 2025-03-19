package main

import (
	"fmt"
	"log"
)

func main() {
	err := parseFlags()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(CliOptions.String())

	err = run()
	if err != nil {
		log.Fatal(err)
	}
}

func run() error {

	return nil
}
