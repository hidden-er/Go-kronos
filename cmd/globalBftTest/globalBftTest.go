package main

import (
	"Chamael/internal/bft"
	"Chamael/internal/party"
	"Chamael/pkg/config"
	"Chamael/pkg/utils/division"
	"time"

	"fmt"
	"log"
	"os"
)

func main() {
	ConfigFile := os.Args[1]
	Mode := os.Args[2]
	var Debug bool
	if Mode == "Debug" {
		Debug = true
	} else {
		Debug = false
	}

	c, err := config.NewHonestConfig(ConfigFile, true)
	if err != nil {
		fmt.Println(err)
	}

	p := party.NewHonestParty(uint32(c.N), uint32(c.F), uint32(c.M), uint32(c.PID), uint32(c.Snumber), uint32(c.SID), c.IPList, c.PortList, division.CalculateShards(c.M, c.N, c.PID), c.PK, c.SK, Debug)
	p.InitReceiveChannel()

	//fmt.Println(p.PID, p.ShardList)
	time.Sleep(time.Second * time.Duration(c.PrepareTime/10))

	p.InitSendChannel()

	inputChannel := make(chan []string, 1024)
	outputChannel := make(chan []string, 1024)

	txs := []string{"test-txs1", "tx2", "tx369"}
	inputChannel <- txs

	fmt.Println("Start HotStuffProcess", p.PID)
	bft.HotStuffProcess(p, 1, inputChannel, outputChannel, true)

	txs_out := <-outputChannel
	fmt.Println("txs_out:", txs_out, p.PID)

	time.Sleep(time.Second * 5)

	log.Println("exit safely", p.PID)
}
