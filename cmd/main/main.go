package main

import (
	"Chamael/internal/bft"
	"Chamael/internal/party"
	"Chamael/pkg/config"
	"Chamael/pkg/txs"
	"Chamael/pkg/utils/db"
	"Chamael/pkg/utils/division"
	"Chamael/pkg/utils/logger"
	"time"

	"fmt"
	"log"
	"os"
	"strconv"
)

func main() {
	B, err := strconv.Atoi(os.Args[1])
	ConfigFile := os.Args[2]
	if err != nil {
		log.Fatalln(err)
	}

	c, err := config.NewHonestConfig(ConfigFile, true)
	if err != nil {
		fmt.Println(err)
	}

	p := party.NewHonestParty(uint32(c.N), uint32(c.F), uint32(c.M), uint32(c.PID), uint32(c.Snumber), uint32(c.SID), c.IPList, c.PortList, division.CalculateShards(c.M, c.N, c.PID), c.PK, c.SK)
	p.InitReceiveChannel()

	//fmt.Println(p.PID, p.ShardList)
	time.Sleep(time.Second * time.Duration(c.PrepareTime/10))

	p.InitSendChannel()

	txlength := 32

	if B == 0 {
		B = c.N
	}

	isTxnum := int(float64(c.Txnum) * (1 - c.Crate))
	csTxnum := c.Txnum - isTxnum

	const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	var Txs []string
	for i := 0; i < isTxnum; i++ {
		tx := txs.InterTxGenerator(txlength, int(p.Snumber), int(p.PID), chars)
		Txs = append(Txs, tx)
	}

	itxdb := fmt.Sprintf("/home/hiddener/Chamael/db/inter_txs_node%d.db", p.PID)
	//fmt.Println(itxdb)
	db.SaveTxsToSQL(Txs, itxdb)
	fmt.Println("Inner-Shard Transactions saved to SQLite database.")

	/*
		testMessage := core.Encapsulation("Execute", make([]byte, 1), p.PID, &protobuf.Execute{
			Unknown: make([]byte, 1),
		})
		p.Send(testMessage, 0)
		p.GetMessage("Execute", make([]byte, 1))
	*/

	//time.Sleep(time.Second * time.Duration(c.PrepareTime/10))
	//ctxdb := fmt.Sprintf("/home/hiddener/Chamael/cross_txs.db", p.PID)\
	ctxdb := "/home/hiddener/Chamael/cross_txs.db"

	itx_inputChannel := make(chan []string, 1024)
	ctx_inputChannel := make(chan []string, 1024)
	outputChannel := make(chan []string, 1024)

	//预先装入一些交易
	for e := 1; e <= c.TestEpochs; e++ {
		itxs, _ := db.LoadAndDeleteTxsFromDB(itxdb, isTxnum)
		itx_inputChannel <- itxs
		ctxs, _ := db.LoadAndDeleteTxsFromDB(ctxdb, csTxnum)
		ctx_inputChannel <- ctxs
	}

	//go bft.MainProcess(p, inputChannel, outputChannel) //把节点独立出来,inputChannel放入Txs,OutputChannel接取
	//go bft.MainProcess(p, c.TestEpochs, itx_inputChannel, ctx_inputChannel, outputChannel)

	//go bft.HotStuffProcess(p, c.TestEpochs, itx_inputChannel, outputChannel)
	/*for i := 1; i <= c.TestEpochs; i++ {
		bft.HotStuffProcess(p, i, itx_inputChannel, outputChannel)
	}*/

	timeChannel := make(chan time.Time, 1024)
	timeChannel <- time.Now()
	go bft.KronosProcess(p, c.TestEpochs, itx_inputChannel, ctx_inputChannel, outputChannel, timeChannel)

	time.Sleep(time.Second * 15)
	logger.CalculateTPS(c, *p, "/home/hiddener/Chamael/log/", timeChannel, outputChannel)
	//logger.RenameHonest(c, *p, "/home/hiddener/Chamael/log/")
	log.Println("exit safely", p.PID)
}
