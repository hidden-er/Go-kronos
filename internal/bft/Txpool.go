package bft

import (
	"Chamael/pkg/txs"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"reflect"
	"sort"
	"sync"
)

// 交易池数据结构
type TransactionPool struct {
	mu           sync.Mutex
	transactions map[string]*TransactionRecord
}

// 交易记录数据结构
type TransactionRecord struct {
	Transaction    *txs.Transaction
	ReceivedShards []int
}

// 创建一个新的交易池
func NewTransactionPool() *TransactionPool {
	return &TransactionPool{
		transactions: make(map[string]*TransactionRecord),
	}
}

// 添加交易到交易池
func (tp *TransactionPool) AddTransaction(tx string, shardID int) error {
	tp.mu.Lock()
	defer tp.mu.Unlock()

	// 提取交易详情
	transactionDetails, err := txs.ExtractTransactionDetails(tx)
	if err != nil {
		return err
	}

	// 唯一标识交易的键
	txKey := generateTransactionKey(tx)

	// 检查交易池中是否已存在
	if record, exists := tp.transactions[txKey]; exists {
		// 更新已接收到的分片列表
		if !contains(record.ReceivedShards, shardID) {
			record.ReceivedShards = append(record.ReceivedShards, shardID)
		}
	} else {
		// 新交易，加入交易池
		tp.transactions[txKey] = &TransactionRecord{
			Transaction:    transactionDetails,
			ReceivedShards: []int{shardID},
		}
	}

	return nil
}

// 检查交易池并移除满足条件的交易
func (tp *TransactionPool) CheckAndRemoveTransactions() []string {
	tp.mu.Lock()
	defer tp.mu.Unlock()

	var completedTransactions []string

	for key, record := range tp.transactions {
		// 判断是否所有输入分片都已发送了该交易
		sort.Ints(record.ReceivedShards)
		sort.Ints(record.Transaction.InputShard)
		if reflect.DeepEqual(record.ReceivedShards, record.Transaction.InputShard) {
			// 条件满足，构建格式化字符串
			formattedTx := fmt.Sprintf(
				"<Dummy TX: %s, Userset: 10, Input Shard: %v, Input Valid: %v, Output Shard: %d, Output Valid: %d>",
				key,
				record.Transaction.InputShard,
				record.Transaction.InputValid,
				record.Transaction.OutputShard,
				record.Transaction.OutputValid,
			)
			// 加入完成列表
			completedTransactions = append(completedTransactions, formattedTx)
			// 从交易池中移除
			delete(tp.transactions, key)
		}
	}

	return completedTransactions
}

// 判断切片是否包含指定元素
func contains(slice []int, value int) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}

// 生成交易的唯一标识键
func generateTransactionKey(tx string) string {
	hash := sha256.Sum256([]byte(tx))
	return hex.EncodeToString(hash[:]) // 返回交易的 SHA-256 哈希值作为唯一键
}

// 打印交易池详情
func (tp *TransactionPool) PrintTxPoolDetail() {
	tp.mu.Lock()
	defer tp.mu.Unlock()

	if len(tp.transactions) == 0 {
		fmt.Println("Transaction Pool is empty.")
		return
	}

	fmt.Println("Transaction Pool Details:")
	for key, record := range tp.transactions {
		fmt.Printf("Transaction Hash: %s\n", key)
		fmt.Printf("  Transaction: %+v\n", record.Transaction)
		fmt.Printf("  Received Shards: %v\n", record.ReceivedShards)
		fmt.Println("-----------------------------------")
	}
}
