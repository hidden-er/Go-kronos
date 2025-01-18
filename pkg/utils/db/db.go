package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

func SaveTxsToSQL(txs []string, filename string) {
	if _, err := os.Stat(filename); err == nil {
		err := os.Remove(filename)
		if err != nil {
			log.Fatalf("Error deleting SQLite database file: %v\n", err)
		}
		log.Printf("Existing database file '%s' removed.\n", filename)
	}

	db, err := sql.Open("sqlite3", filename)
	if err != nil {
		log.Fatalf("Error opening SQLite database: %v\n", err)
	}
	defer db.Close()

	// Create the table for transactions
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS transactions (
	    id INTEGER PRIMARY KEY,
		tx TEXT NOT NULL
	);`
	_, err = db.Exec(createTableSQL)
	if err != nil {
		log.Fatalf("Error creating table: %v\n", err)
	}

	// Clear the table before inserting new data
	clearTableSQL := `DELETE FROM transactions;`
	_, err = db.Exec(clearTableSQL)
	if err != nil {
		log.Fatalf("Error clearing table: %v\n", err)
	}

	// Insert transactions into the table
	insertSQL := `INSERT INTO transactions (tx) VALUES (?)`
	stmt, err := db.Prepare(insertSQL)
	if err != nil {
		log.Fatalf("Error preparing insert statement: %v\n", err)
	}
	defer stmt.Close()

	for _, tx := range txs {
		_, err = stmt.Exec(tx)
		if err != nil {
			log.Fatalf("Error inserting transaction: %v\n", err)
		}
	}
}

func LoadAndDeleteTxsFromDB(dbPath string, limit int) ([]string, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}
	defer db.Close()

	// 使用 LIMIT 限制返回的事务数量
	rows, err := db.Query("SELECT * FROM transactions LIMIT ?", limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query database: %v", err)
	}
	defer rows.Close()

	var txs []string
	var txIDs []int

	// 读取数据库中的事务
	for rows.Next() {
		var id int
		var tx string
		if err := rows.Scan(&id, &tx); err != nil {
			return nil, fmt.Errorf("failed to scan row: %v", err)
		}
		txs = append(txs, tx)
		txIDs = append(txIDs, id) // 记录事务的 ID 以便后续删除
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %v", err)
	}

	// 删除已读取的事务
	for _, id := range txIDs {
		_, err := db.Exec("DELETE FROM transactions WHERE id = ?", id)
		if err != nil {
			return nil, fmt.Errorf("failed to delete transaction with id %d: %v", id, err)
		}
	}

	return txs, nil
}
