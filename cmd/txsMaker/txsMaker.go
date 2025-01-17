package main

import (
	"Chamael/pkg/txs"
	"Chamael/pkg/utils/db"
	"flag"
	"fmt"
	"log"
	"os"
)

func main() {
	// Parse command-line arguments
	shardNum := flag.Int("shard_num", 0, "Number of shards")
	txNum := flag.Int("tx_num", 0, "Number of transactions")
	rRate := flag.Int("Rrate", 10, "Percentage of true input validity")
	PID := flag.Int("PID", 10, "User set")
	flag.Parse()

	if *shardNum <= 0 || *txNum <= 0 {
		fmt.Println("Invalid arguments: shard_num and tx_num must be positive integers.")
		os.Exit(1)
	}

	const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	var Txs []string
	for i := 0; i < *txNum; i++ {
		tx := txs.CrossTxGenerator(32, *shardNum, *rRate, *PID, chars)
		Txs = append(Txs, tx)
	}

	db.SaveTxsToSQL(Txs, "/home/hiddener/Chamael/cross_txs.db")
	log.Println("Cross-Shard Transactions saved to SQLite database.")
}
