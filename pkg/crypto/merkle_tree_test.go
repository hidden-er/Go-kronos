package crypto

import (
	"bytes"
	"fmt"
	m "github.com/cbergoon/merkletree"
	"testing"
)

func generateDummyMerkleTree() (*MerkleTree, error) {
	data := [][]string{
		{"<Dummy TX: xxxx01\n", "<Dummy TX: xxxx02\n"},
		{"<Dummy TX: xxxx03\n", "<Dummy TX: xxxx04\n"},
		{"<Dummy TX: xxxx05\n", "<Dummy TX: xxxx06\n"},
		{"<Dummy TX: xxxx07\n", "<Dummy TX: xxxx08\n"},
		{"<Dummy TX: xxxx09\n", "<Dummy TX: xxxx10\n"},
		{"<Dummy TX: xxxx11\n", "<Dummy TX: xxxx12\n"},
		{"<Dummy TX: xxxx13\n", "<Dummy TX: xxxx14\n"},
		{"<Dummy TX: xxxx15\n", "<Dummy TX: xxxx16\n"},
	}
	return NewMerkleTree(data)
}

/*
func validateMerkleTrees(t *testing.T, tree1 *m.MerkleTree, tree2 *m.MerkleTree) {
	// 验证根节点是否一致
	validateNodes(t, tree1.Root, tree2.Root)

	// 验证叶子节点是否一致
	if len(tree1.Leafs) != len(tree2.Leafs) {
		t.Fatalf("Leaf count mismatch: got %d, want %d", len(tree2.Leafs), len(tree1.Leafs))
	}
	for i := 0; i < len(tree1.Leafs); i++ {
		validateNodes(t, tree1.Leafs[i], tree2.Leafs[i])
	}
} */

func validateNodes(t *testing.T, node1 *m.Node, node2 *m.Node) {
	if node1 == nil && node2 == nil {
		return
	}
	if node1 == nil || node2 == nil {
		t.Errorf("Node mismatch: one of the nodes is nil")
		return
	}

	// 验证哈希值是否一致
	if !bytes.Equal(node1.Hash, node2.Hash) {
		t.Errorf("Node hash mismatch: got %x, want %x", node2.Hash, node1.Hash)
	}

	// 验证子节点
	validateNodes(t, node1.Left, node2.Left)
	validateNodes(t, node1.Right, node2.Right)
}

//<Dummy TX: DXMEZ80U753ISQPWEIRBZSVZ2ALAU9H6, Userset: 0, Input Shard: [0], Input Valid: [1], Output Shard: 0, Output Valid: 1 >
func TestNewMerkleTree(t *testing.T) {
	tre, err := generateDummyMerkleTree()
	if err != nil {
		t.Errorf("%s", err.Error())
	}
	b0, _ := tre.GetMerkleTreeProof(0)
	b1, _ := tre.GetMerkleTreeProof(1)

	if !bytes.Equal(b0[1], b1[1]) {
		t.Errorf("proof generate fail")
	}

	b2, _ := tre.GetMerkleTreeProof(2)
	b3, _ := tre.GetMerkleTreeProof(3)

	if !bytes.Equal(b2[1], b3[1]) {
		t.Errorf("proof generate fail")
	}
}

func TestVerifyMerkleTreeProof(t *testing.T) {
	data := [][]string{
		{"<Dummy TX: xxxx01\n", "<Dummy TX: xxxx02\n"},
		{"<Dummy TX: xxxx03\n", "<Dummy TX: xxxx04\n"},
		{"<Dummy TX: xxxx05\n", "<Dummy TX: xxxx06\n"},
		{"<Dummy TX: xxxx07\n", "<Dummy TX: xxxx08\n"},
		{"<Dummy TX: xxxx09\n", "<Dummy TX: xxxx10\n"},
		{"<Dummy TX: xxxx11\n", "<Dummy TX: xxxx12\n"},
		{"<Dummy TX: xxxx13\n", "<Dummy TX: xxxx14\n"},
		{"<Dummy TX: xxxx15\n", "<Dummy TX: xxxx16\n"},
	}
	tre, err := NewMerkleTree(data)
	if err != nil {
		t.Errorf("%s", err.Error())
	}
	root := tre.GetMerkleTreeRoot()

	b0, indic0 := tre.GetMerkleTreeProof(0)
	fmt.Println(indic0)
	if !VerifyMerkleTreeProof(root, b0, indic0, data[0]) {
		t.Errorf("verify fail")
	}
	b1, indic1 := tre.GetMerkleTreeProof(1)
	fmt.Println(indic1)
	if !VerifyMerkleTreeProof(root, b1, indic1, data[1]) {
		t.Errorf("verify fail")
	}
	b2, indic2 := tre.GetMerkleTreeProof(2)
	fmt.Println(indic2)
	if !VerifyMerkleTreeProof(root, b2, indic2, data[2]) {
		t.Errorf("verify fail")
	}
	b3, indic3 := tre.GetMerkleTreeProof(3)
	fmt.Println(indic3)
	if !VerifyMerkleTreeProof(root, b3, indic3, data[3]) {
		t.Errorf("verify fail")
	}
}

/*
func TestMerkleTreeSerialization(t *testing.T) {
	// 使用公共函数生成 Merkle Tree
	tre, err := generateDummyMerkleTree()
	if err != nil {
		t.Fatalf("Failed to generate Merkle Tree: %s", err.Error())
	}

	// 测试序列化
	data, err := tre.Marshal()
	if err != nil {
		t.Fatalf("Failed to serialize Merkle Tree: %s", err.Error())
	}

	// 测试反序列化
	var newTree MerkleTree
	err = newTree.Unmarshal(data)
	if err != nil {
		t.Fatalf("Failed to deserialize Merkle Tree: %s", err.Error())
	}

	// 验证 contents 数据是否一致
	if len(tre.contents) != len(newTree.contents) {
		t.Fatalf("Contents length mismatch: got %d, want %d", len(newTree.contents), len(tre.contents))
	}
	for i := 0; i < len(tre.contents); i++ {
		origContent, ok1 := tre.contents[i].(*implContent)
		newContent, ok2 := newTree.contents[i].(*implContent)

		if !ok1 || !ok2 {
			t.Errorf("Content type mismatch at index %d", i)
			continue
		}

		// 比较内容长度
		if len(origContent.x) != len(newContent.x) {
			t.Errorf("Content length mismatch at index %d: got %d, want %d", i, len(newContent.x), len(origContent.x))
			continue
		}

		// 比较具体内容
		for j := 0; j < len(origContent.x); j++ {
			if origContent.x[j] != newContent.x[j] {
				t.Errorf("Content mismatch at index %d, item %d: got %s, want %s", i, j, newContent.x[j], origContent.x[j])
			}
		}
	}

	// 验证 mktree 是否完全相同
	validateMerkleTrees(t, tre.mktree, newTree.mktree)
}*/
