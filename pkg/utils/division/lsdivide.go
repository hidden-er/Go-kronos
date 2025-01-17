package division

import "math"

// 计算当前节点需要负责的分片
func CalculateShards(m, n, currentNode int) []int {
	currentShard := currentNode / n
	var shards []int // 存储当前节点负责的分片编号

	if m <= n {
		// 分片数小于或等于节点数
		nodesPerShard := n / (m - 1) // 每个分片的节点数

		shard := 0
		cnt := 0
		for node := 0; node < n; node++ {
			if shard == currentShard { //自己的分片不需要负责
				shard++
			}
			cnt++

			if node == currentNode%n { // 分配当前节点
				shards = append(shards, shard)
				break
			}

			if cnt == nodesPerShard {
				cnt = 0
				shard++
			}
		}
	} else {
		// 分片数大于节点数
		shardsPerNode := int(math.Ceil(float64(m-1) / float64(n))) // 每个节点负责的分片数

		shard := 0
		for node := 0; node < n; node++ {
			cnt := 0
			for {
				cnt++

				if node == currentNode%n { // 分配当前节点
					shards = append(shards, shard)
				}
				shard++
				if shard == currentShard { //自己的分片不需要负责
					shard++
				}

				if shard >= m {
					return shards
				}
				if cnt == shardsPerNode {
					break
				}
			}
		}
	}

	return shards
}
