package division

import (
	"reflect"
	"testing"
)

// 测试函数
func TestCalculateShards(t *testing.T) {
	tests := []struct {
		m           int   // 分片数
		n           int   // 片内节点数
		currentNode int   // 当前节点编号
		expected    []int // 期望的分片结果
	}{
		{m: 2, n: 4, currentNode: 3, expected: []int{1}},
		{m: 2, n: 4, currentNode: 4, expected: []int{0}},
		{m: 3, n: 4, currentNode: 2, expected: []int{2}},
		{m: 3, n: 4, currentNode: 5, expected: []int{0}},
		{m: 8, n: 4, currentNode: 17, expected: []int{2, 3}},
		{m: 8, n: 4, currentNode: 27, expected: []int{7}},
		{m: 16, n: 4, currentNode: 37, expected: []int{4, 5, 6, 7}},
		{m: 16, n: 4, currentNode: 38, expected: []int{8, 10, 11, 12}},
		{m: 16, n: 4, currentNode: 39, expected: []int{13, 14, 15}},
		{m: 32, n: 8, currentNode: 100, expected: []int{17, 18, 19, 20}},
		{m: 32, n: 8, currentNode: 103, expected: []int{29, 30, 31}},
	}

	for _, tt := range tests {
		t.Run(
			// 描述测试名称
			t.Name(),
			func(t *testing.T) {
				result := CalculateShards(tt.m, tt.n, tt.currentNode)
				if !reflect.DeepEqual(result, tt.expected) {
					t.Errorf("For m=%d, n=%d, currentNode=%d: expected %v, got %v",
						tt.m, tt.n, tt.currentNode, tt.expected, result)
				}
			},
		)
	}
}
