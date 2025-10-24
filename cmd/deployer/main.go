package main

import (
	"log"

	"github.com/eliasfloreteng/github-auto-deployer/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		log.Fatal(err)
	}
}
