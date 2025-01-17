package main

import (
	"Chamael/pkg/config"
	"log"
)

func main() {
	c, err := config.NewHonestConfig("./cmd/main/config.yaml", true)
	if err != nil {
		log.Fatalln(err)
	}
	c.RemoteHonestGen("./configs")
}
