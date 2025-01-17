package txs

import (
	"fmt"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Transaction struct {
	DummyTX     string `json:"DummyTX"`
	InputShard  []int  `json:"InputShard"`
	InputValid  []int  `json:"InputValid"`
	OutputShard int    `json:"OutputShard"`
	OutputValid int    `json:"OutputValid"`
}

func randomString(size int, chars string) string {
	result := make([]byte, size)
	for i := range result {
		result[i] = chars[rand.Intn(len(chars))]
	}
	return string(result)
}
func randomSample(min, max, count int) []int {
	all := rand.Perm(max - min)
	result := all[:count]
	for i := range result {
		result[i] += min
	}
	return result
}
func contains(slice []int, value int) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}

func InterTxGenerator(size int, shardID int, PID int, chars string) string {
	rand.Seed(time.Now().UnixNano())
	randomString := randomString(size, chars)
	shardInfo := fmt.Sprintf(", Userset: %d, Input Shard: [%d], Input Valid: [1], Output Shard: %d, Output Valid: 2", PID, shardID, shardID)
	//↑目前只考虑合法交易
	return fmt.Sprintf("<Dummy TX: %s%s >", randomString, shardInfo)
}
func CrossTxGenerator(size, shardNum, Rrate int, PID int, chars string) string {
	rand.Seed(time.Now().UnixNano())

	// Determine the number of input shards
	inputShardNum := rand.Intn(3) + 1
	if inputShardNum >= shardNum {
		inputShardNum = shardNum - 1
	}

	// Select input shards
	inputShards := randomSample(0, shardNum, inputShardNum)

	// Generate input validity
	inputValid := make([]int, inputShardNum)
	for i := range inputValid {
		/*
			if rand.Intn(100) < Rrate {
				inputValid[i] = 1
			} else {
				inputValid[i] = 0
			}
		*/
		inputValid[i] = 1 //目前只考虑合法交易
	}

	// Choose output shard
	outputShard := -1
	for {
		candidate := rand.Intn(shardNum)
		if !contains(inputShards, candidate) {
			outputShard = candidate
			break
		}
	}

	randomString := randomString(size, chars)
	shardInfo := fmt.Sprintf(", Userset: %d, Input Shard: %v, Input Valid: %v, Output Shard: %d, Output Valid: 0",
		PID, inputShards, inputValid, outputShard)

	return fmt.Sprintf("<Dummy TX: %s%s >", randomString, shardInfo)
}

/*
func GroupAndSortTransactions(txBatchList []string) ([][]string, []string) {
	// 创建一个 map，用于按 Output Shard 分组交易
	groupedcTx := make(map[string][]string)
	var groupediTx []string

	// 遍历交易列表，按 Output Shard 分组
	for _, tx := range txBatchList {

		// 提取 Output Vaild 的值, 据此确定片内交易
		// "【......Output Shard: 0, Output Valid: 1 >】那个表明输出可用性的数字后面固定跟2个字符
		shardStart := strings.Index(tx, "Output Shard: ") + len("Output Shard: ")
		shardEnd := len(tx) - 2
		outputVaild := tx[shardStart:shardEnd]
		if outputVaild[0] == '2' {
			groupediTx = append(groupediTx, tx)
		} else {
			// 提取 Output Shard 的值
			// "【......Output Shard: 0, Output Valid: 1 >】那个表明输出分片的数字后面固定跟19个字符
			shardStart := strings.Index(tx, "Output Shard: ") + len("Output Shard: ")
			shardEnd := len(tx) - 19
			outputShard := tx[shardStart:shardEnd]

			// 将交易添加到对应分组
			if _, exists := groupedcTx[outputShard]; !exists {
				groupedcTx[outputShard] = []string{}
			}
			groupedcTx[outputShard] = append(groupedcTx[outputShard], tx)
		}
	}

	// 对分组的 Output Shard 进行排序
	var sortedShards []string
	for shard := range groupedcTx {
		sortedShards = append(sortedShards, shard)
	}
	sort.Strings(sortedShards) // 按字典序对 shard 排序

	// 构造最终的输出结果
	var batchctx [][]string
	for _, shard := range sortedShards {
		batchctx = append(batchctx, groupedcTx[shard])
	}

	return batchctx, groupediTx
}
*/

func ExtractTransactionDetails(tx string) (*Transaction, error) {
	// 定义正则表达式模式
	re := regexp.MustCompile(
		`Input Shard: \[([0-9 ]+)\], Input Valid: \[([0-9 ]+)\], Output Shard: ([0-9]+), Output Valid: ([0-9]+)`,
	)

	// 查找匹配
	matches := re.FindStringSubmatch(tx)
	if len(matches) < 5 {
		return nil, fmt.Errorf("transaction format is invalid")
	}

	// 提取和解析数据
	inputShardsStr := matches[1]
	inputValidsStr := matches[2]
	outputShardStr := matches[3]
	outputValidStr := matches[4]

	// 解析 InputShard 列表
	inputShards := parseIntList(inputShardsStr)
	// 解析 InputValid 列表
	inputValids := parseIntList(inputValidsStr)
	// 解析 OutputShard 和 OutputValid
	outputShard, err := strconv.Atoi(outputShardStr)
	if err != nil {
		return nil, fmt.Errorf("invalid Output Shard: %v", err)
	}
	outputValid, err := strconv.Atoi(outputValidStr)
	if err != nil {
		return nil, fmt.Errorf("invalid Output Valid: %v", err)
	}

	// 返回交易数据结构
	return &Transaction{
		InputShard:  inputShards,
		InputValid:  inputValids,
		OutputShard: outputShard,
		OutputValid: outputValid,
	}, nil
}

// 辅助函数：解析一个以逗号分隔的数字字符串为整型列表
func parseIntList(str string) []int {
	str = strings.Trim(str, "[]")           // 去掉开头和结尾的方括号
	str = strings.ReplaceAll(str, ",", " ") // 替换逗号为空格（支持逗号分隔格式）
	str = strings.TrimSpace(str)            // 去掉首尾空格
	numStrs := strings.Fields(str)          // 根据空格分割
	var nums []int
	for _, numStr := range numStrs {
		num, err := strconv.Atoi(numStr)
		if err == nil {
			nums = append(nums, num)
		}
	}
	return nums
}
