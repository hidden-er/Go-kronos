package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// 累加的六项数值
func AccumulateTPSStats(dir string) (int, int, int, float64, float64, float64, error) {
	// 正则表达式用于匹配文件中的数据
	totalTxReg := regexp.MustCompile(`Total Transactions:\s*(\d+)`)
	internalTxReg := regexp.MustCompile(`Internal Transactions:\s*(\d+)`)
	crossShardTxReg := regexp.MustCompile(`Cross-Shard Transactions:\s*(\d+)`)
	totalTPSReg := regexp.MustCompile(`Total TPS:\s*([\d\.]+)`)
	internalTPSReg := regexp.MustCompile(`Internal TPS:\s*([\d\.]+)`)
	crossShardTPSReg := regexp.MustCompile(`Cross-Shard TPS:\s*([\d\.]+)`)

	var totalTransactions, internalTransactions, crossShardTransactions int
	var totalTPS, internalTPS, crossShardTPS float64

	// 遍历目录下的所有文件
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 只处理以 (Performance) 开头的文件
		if strings.HasPrefix(info.Name(), "(Performance)") {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			// 逐行读取文件
			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				line := scanner.Text()

				// 匹配每一项并累加
				if matches := totalTxReg.FindStringSubmatch(line); matches != nil {
					total, err := strconv.Atoi(matches[1])
					if err == nil {
						totalTransactions += total
					}
				}

				if matches := internalTxReg.FindStringSubmatch(line); matches != nil {
					internal, err := strconv.Atoi(matches[1])
					if err == nil {
						internalTransactions += internal
					}
				}

				if matches := crossShardTxReg.FindStringSubmatch(line); matches != nil {
					crossShard, err := strconv.Atoi(matches[1])
					if err == nil {
						crossShardTransactions += crossShard
					}
				}

				if matches := totalTPSReg.FindStringSubmatch(line); matches != nil {
					tps, err := strconv.ParseFloat(matches[1], 64)
					if err == nil {
						totalTPS += tps
					}
				}

				if matches := internalTPSReg.FindStringSubmatch(line); matches != nil {
					tps, err := strconv.ParseFloat(matches[1], 64)
					if err == nil {
						internalTPS += tps
					}
				}

				if matches := crossShardTPSReg.FindStringSubmatch(line); matches != nil {
					tps, err := strconv.ParseFloat(matches[1], 64)
					if err == nil {
						crossShardTPS += tps
					}
				}
			}

			if err := scanner.Err(); err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return 0, 0, 0, 0, 0, 0, err
	}

	return totalTransactions, internalTransactions, crossShardTransactions, totalTPS, internalTPS, crossShardTPS, nil
}

func main() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Error getting home directory:", err)
		return
	}

	totalTx, internalTx, crossShardTx, totalTps, internalTps, crossShardTps, err := AccumulateTPSStats(homeDir + "/Chamael/log/")
	if err != nil {
		fmt.Println("Error accumulating stats:", err)
	} else {
		fmt.Printf("Total Transactions: %d\nInternal Transactions: %d\nCross-Shard Transactions: %d\n", totalTx, internalTx, crossShardTx)
		fmt.Printf("Total TPS: %.2f\nInternal TPS: %.2f\nCross-Shard TPS: %.2f\n", totalTps, internalTps, crossShardTps)
	}
}
