package crypto

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"strings"

	m "github.com/cbergoon/merkletree"
	"golang.org/x/crypto/sha3"
)

// 注册类型到 gob
func init() {
	gob.Register(&implContent{})
}

// implement of m.Content
type implContent struct {
	x []string
}

func buildImplContent(x []string) *implContent {
	return &implContent{x: x}
}

func (i *implContent) CalculateHash() ([]byte, error) {
	hash := sha3.Sum512([]byte(strings.Join(i.x, "")))
	return hash[:], nil
}

func (i *implContent) Equals(other m.Content) (bool, error) {
	hash1, _ := other.CalculateHash()
	hash2, _ := i.CalculateHash()
	if bytes.Equal(hash1, hash2) {
		return true, nil
	}
	return false, nil
}

//MerkleTree is a kind of vector commitment
type MerkleTree struct {
	mktree   *m.MerkleTree
	contents []m.Content
}

//NewMerkleTree generates a merkletree
func NewMerkleTree(data [][]string) (*MerkleTree, error) {
	contents := []m.Content{}
	for _, d := range data {
		c := buildImplContent(d)
		contents = append(contents, c)
	}
	mk, err := m.NewTreeWithHashStrategy(contents, sha3.New512)
	if err != nil {
		return nil, err
	}
	return &MerkleTree{
		mktree:   mk,
		contents: contents,
	}, nil
}

// GetMerkleTreeRoot returns a Merkle tree root
func (t *MerkleTree) GetMerkleTreeRoot() []byte {
	if t == nil {
		fmt.Println("Warning: Merkle tree is nil")
		return nil // 返回 nil 或者一个默认值
	}
	return t.mktree.MerkleRoot()
}

// GetMerkleTreeProof returns a vector commitment
func (t *MerkleTree) GetMerkleTreeProof(id int) ([][]byte, []int64) {
	if t == nil {
		fmt.Println("Waring: Merkle tree is nil")
		return nil, nil // 返回空值，表示失败
	}
	if id < 0 || id >= len(t.contents) {
		fmt.Printf("Error: Invalid ID %d, must be within [0, %d)\n", id, len(t.contents))
		return nil, nil
	}
	path, indicator, _ := t.mktree.GetMerklePath(t.contents[id])
	return path, indicator
}

//VerifyMerkleTreeProof returns a vector commitment
func VerifyMerkleTreeProof(root []byte, proof [][]byte, indicator []int64, msg []string) bool {
	if len(proof) != len(indicator) {
		return false
	}
	itHash, _ := (&implContent{x: msg}).CalculateHash()
	for i, p := range proof {
		s := sha3.New512()
		if indicator[i] == 1 {
			s.Write(append(itHash, p...))
		} else if indicator[i] == 0 {
			s.Write(append(p, itHash...))
		} else {
			return false
		}
		itHash = s.Sum(nil)
	}
	return bytes.Equal(itHash, root)
}
