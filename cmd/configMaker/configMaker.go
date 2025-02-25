package main

import (
	"Chamael/pkg/config"
	"flag"
	"log"
)

func main() {
	configPath := flag.String("config_path", "", "Original config file path")
	flag.Parse()

	if *configPath == "" {
		log.Fatalln("Usage configMaker -config_path <config_path>")
	}

	c, err := config.NewHonestConfig(*configPath, true)
	if err != nil {
		log.Fatalln(err)
	}
	c.RemoteHonestGen("./configs")
}
