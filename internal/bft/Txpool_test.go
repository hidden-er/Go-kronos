package bft

import (
	"testing"
)

// 测试交易池的添加和移除功能
func TestTransactionPool(t *testing.T) {
	// 创建交易池
	pool := NewTransactionPool()

	// 示例交易
	tx1 := "<Dummy TX: CZZDB66YF4F42B1I6W0FM2OHTIC96SRM, Userset: 10, Input Shard: [0 1 2], Input Valid: [1], Output Shard: 1, Output Valid: 0 >"

	// 添加交易到分片 0 和 1
	err := pool.AddTransaction(tx1, 0)
	if err != nil {
		t.Fatalf("failed to add transaction to shard 0: %v", err)
	}
	err = pool.AddTransaction(tx1, 1)
	if err != nil {
		t.Fatalf("failed to add transaction to shard 1: %v", err)
	}

	// 检查池中未完成的交易
	if len(pool.transactions) != 1 {
		t.Errorf("expected 1 transaction in pool, got %d", len(pool.transactions))
	}

	// 添加剩余的分片
	err = pool.AddTransaction(tx1, 2)
	if err != nil {
		t.Fatalf("failed to add transaction to shard 2: %v", err)
	}

	// 检查并移除完成的交易
	completed := pool.CheckAndRemoveTransactions()
	if len(completed) != 1 {
		t.Errorf("expected 1 completed transaction, got %d", len(completed))
	}

	// 检查池中剩余交易
	if len(pool.transactions) != 0 {
		t.Errorf("expected 0 transactions in pool, got %d", len(pool.transactions))
	}
}
